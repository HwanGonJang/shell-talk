package network

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Conversation holds the state of a single chat (DM or Room).
type Conversation struct {
	ID      string // Nickname for DMs, Room Name for rooms
	Type    string // "DM" or "ROOM"
	Joined  bool   // Only relevant for rooms
	History []string
	mu      sync.RWMutex
}

// Client manages the WebSocket connection and the user interface state.
type Client struct {
	Conn                *websocket.Conn
	Send                chan WebSocketMessage
	AuthInfo            *LoginSuccessPayload // Populated after successful login
	AuthCh              chan bool            // Signals successful authentication
	conversations       map[string]*Conversation
	currentConversation *Conversation
	mu                  sync.RWMutex
}

// NewClient creates a new network client.
func NewClient() *Client {
	return &Client{
		Send:          make(chan WebSocketMessage, 256),
		AuthCh:        make(chan bool),
		conversations: make(map[string]*Conversation),
	}
}

// Connect establishes a WebSocket connection to the server.
func (c *Client) Connect(serverURL string) error {
	u, err := url.Parse(serverURL)
	if err != nil {
		return err
	}

	log.Printf("Connecting to %s...", u.String())
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}
	c.Conn = conn

	go c.readPump()
	go c.writePump()

	return nil
}

func (c *Client) readPump() {
	defer c.Conn.Close()
	for {
		var msg WebSocketMessage
		err := c.Conn.ReadJSON(&msg)
		if err != nil {
			clearScreen()
			log.Printf("Connection to server lost: %v", err)
			os.Exit(0)
		}
		c.handleServerMessage(msg)
	}
}

func (c *Client) writePump() {
	defer c.Conn.Close()
	for msg := range c.Send {
		err := c.Conn.WriteJSON(msg)
		if err != nil {
			log.Printf("Write error: %v", err)
			return
		}
	}
}

// HandleStdin is the main loop for reading user input post-authentication.
func (c *Client) HandleStdin() {
	reader := bufio.NewReader(os.Stdin)
	c.redrawView()

	for {
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			c.prompt()
			continue
		}

		if strings.HasPrefix(input, "/") {
			c.handleCommand(input)
		} else {
			c.handleChatMessage(input)
		}
	}
}

func (c *Client) handleCommand(input string) {
	parts := strings.Split(input, " ")
	command := parts[0]

	switch command {
	case "/help":
		c.redrawView()
		return
	case "/create":
		if len(parts) < 3 {
			c.printToScreen("[ERROR] Usage: /create <room_name> <password>")
		} else {
			c.Send <- WebSocketMessage{Type: "create_room", Payload: CreateRoomPayload{Name: parts[1], Password: strings.Join(parts[2:], " ")}}
		}
	case "/join":
		if len(parts) < 3 {
			c.printToScreen("[ERROR] Usage: /join <room_name> <password>")
		} else {
			c.Send <- WebSocketMessage{Type: "join_room", Payload: JoinRoomPayload{RoomName: parts[1], Password: strings.Join(parts[2:], " ")}}
		}
	case "/leave":
		if len(parts) < 2 {
			c.printToScreen("[ERROR] Usage: /leave <room_name>")
		} else {
			c.Send <- WebSocketMessage{Type: "leave_room", Payload: LeaveRoomPayload{RoomName: parts[1]}}
		}
	case "/members":
		if len(parts) < 2 {
			c.printToScreen("[ERROR] Usage: /members <room_name>")
		} else {
			c.Send <- WebSocketMessage{Type: "list_members", Payload: ListMembersPayload{RoomName: parts[1]}}
		}
	case "/list":
		c.Send <- WebSocketMessage{Type: "list_rooms"}
	case "/myrooms":
		c.listMyRooms()
		return
	case "/switch":
		if len(parts) < 3 {
			c.printToScreen("[ERROR] Usage: /switch <dm|room> <name>")
			return
		}
		convType, name := parts[1], parts[2]
		if err := c.switchConversation(convType, name); err != nil {
			c.printToScreen(fmt.Sprintf("[ERROR] %s", err.Error()))
		} else {
			c.redrawView()
		}
		return
	case "/exit":
		c.mu.Lock()
		c.currentConversation = nil
		c.mu.Unlock()
		c.redrawView()
		return
	default:
		c.printToScreen(fmt.Sprintf("[ERROR] Unknown command: %s", command))
	}

	c.prompt()
}

func (c *Client) handleChatMessage(input string) {
	c.mu.RLock()
	conv := c.currentConversation
	c.mu.RUnlock()

	if conv == nil {
		c.printToScreen("[ERROR] Not in a conversation. Use /switch <dm|room> <name>.")
		return
	}

	var msgType string
	switch conv.Type {
	case "DM":
		msgType = "send_direct_message"
		c.Send <- WebSocketMessage{Type: msgType, Payload: SendDirectMessagePayload{RecipientNickname: conv.ID, Content: input}}
	case "ROOM":
		msgType = "send_room_message"
		c.Send <- WebSocketMessage{Type: msgType, Payload: SendRoomMessagePayload{RoomName: conv.ID, Content: input}}
	}

	formattedMsg := fmt.Sprintf("[%s] [Me]: %s", time.Now().Format("15:04:05"), input)
	conv.addHistory(formattedMsg)
	c.printToScreen(formattedMsg)
}

func (c *Client) handleServerMessage(msg WebSocketMessage) {
	payloadBytes, _ := json.Marshal(msg.Payload)
	timestamp := time.Now().Format("15:04:05")

	switch msg.Type {
	case "login_success":
		var payload LoginSuccessPayload
		_ = json.Unmarshal(payloadBytes, &payload)
		c.AuthInfo = &payload
		fmt.Printf("\r[SYSTEM] Welcome, %s! Login successful.\n", payload.Nickname)
		c.AuthCh <- true
		return

	case "new_direct_message":
		var payload map[string]interface{}
		_ = json.Unmarshal(payloadBytes, &payload)
		sender := payload["sender"].(string)
		content := payload["content"].(string)
		conv := c.getOrCreateConversation(sender, "DM")
		formattedMsg := fmt.Sprintf("[%s] [%s]: %s", timestamp, sender, content)
		conv.addHistory(formattedMsg)
		c.notifyOrUpdate(sender, "DM", formattedMsg)

	case "room_message":
		var payload map[string]interface{}
		_ = json.Unmarshal(payloadBytes, &payload)
		roomName := payload["room_name"].(string)
		sender := payload["sender_nickname"].(string)
		content := payload["content"].(string)
		conv := c.getOrCreateConversation(roomName, "ROOM")
		formattedMsg := fmt.Sprintf("[%s] [%s]: %s", timestamp, sender, content)
		conv.addHistory(formattedMsg)
		c.notifyOrUpdate(roomName, "ROOM", formattedMsg)

	case "join_success":
		var payload JoinSuccessPayload
		_ = json.Unmarshal(payloadBytes, &payload)
		conv := c.getOrCreateConversation(payload.RoomName, "ROOM")
		conv.mu.Lock()
		conv.Joined = true
		conv.mu.Unlock()
		c.printToScreen(fmt.Sprintf("[SYSTEM] Successfully joined room: %s", payload.RoomName))

	case "leave_success":
		var payload LeaveSuccessPayload
		_ = json.Unmarshal(payloadBytes, &payload)
		c.printToScreen(fmt.Sprintf("[SYSTEM] Successfully left a room."))

	case "room_list":
		var payload RoomListPayload
		_ = json.Unmarshal(payloadBytes, &payload)
		var builder strings.Builder
		builder.WriteString(fmt.Sprintf("\r[%s] [All Rooms]:\n", timestamp))
		if len(payload.Rooms) == 0 {
			builder.WriteString("  No rooms available.")
		} else {
			for _, room := range payload.Rooms {
				builder.WriteString(fmt.Sprintf("  - %s\n", room.Name))
				c.getOrCreateConversation(room.Name, "ROOM")
			}
		}
		c.printToScreen(builder.String())

	case "room_members":
		var payload RoomMembersPayload
		_ = json.Unmarshal(payloadBytes, &payload)
		var builder strings.Builder
		builder.WriteString(fmt.Sprintf("\r[%s] [Members of %s]:\n", timestamp, payload.RoomName))
		for _, member := range payload.Members {
			builder.WriteString(fmt.Sprintf("  - %s\n", member))
		}
		c.printToScreen(builder.String())

	case "error_message", "system_message":
		var payload map[string]interface{}
		_ = json.Unmarshal(payloadBytes, &payload)
		content := payload["content"].(string)
		prefix := "[SYSTEM]"
		if msg.Type == "error_message" {
			prefix = "[ERROR]"
		}
		c.printToScreen(fmt.Sprintf("%s %s", prefix, content))

	default:
		c.printToScreen(fmt.Sprintf("[UNKNOWN] %v", msg))
	}
}

func (c *Client) getOrCreateConversation(name string, convType string) *Conversation {
	key := convType + "_" + name
	c.mu.Lock()
	defer c.mu.Unlock()
	if conv, exists := c.conversations[key]; exists {
		return conv
	}
	conv := &Conversation{ID: name, Type: convType, History: []string{}}
	c.conversations[key] = conv
	return conv
}

func (c *Client) switchConversation(convType, name string) error {
	convType = strings.ToUpper(convType)
	if convType != "DM" && convType != "ROOM" {
		return fmt.Errorf("invalid switch type: must be 'dm' or 'room'")
	}

	key := convType + "_" + name
	c.mu.Lock()
	defer c.mu.Unlock()

	conv, exists := c.conversations[key]
	if convType == "ROOM" {
		if !exists {
			return fmt.Errorf("unknown room: %s. Use /list to see available rooms", name)
		}
		conv.mu.RLock()
		joined := conv.Joined
		conv.mu.RUnlock()
		if !joined {
			return fmt.Errorf("you have not joined room '%s'. Use /join first", name)
		}
	}

	if !exists {
		conv = c.getOrCreateConversation_internal(name, convType)
	}

	c.currentConversation = conv
	return nil
}

func (c *Client) getOrCreateConversation_internal(name, convType string) *Conversation {
	key := convType + "_" + name
	if conv, exists := c.conversations[key]; exists {
		return conv
	}
	conv := &Conversation{ID: name, Type: convType, History: []string{}}
	c.conversations[key] = conv
	return conv
}

func (c *Client) redrawView() {
	clearScreen()
	c.mu.RLock()
	conv := c.currentConversation
	c.mu.RUnlock()

	if conv == nil {
		c.printHelp()
	} else {
		fmt.Printf("--- Conversation with %s (%s) ---\n", conv.ID, conv.Type)
		conv.mu.RLock()
		for _, msg := range conv.History {
			fmt.Println(msg)
		}
		conv.mu.RUnlock()
	}
	c.prompt()
}

func (c *Client) listMyRooms() {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("\r[%s] [My Joined Rooms]:\n", time.Now().Format("15:04:05")))

	found := false
	for key, conv := range c.conversations {
		if strings.HasPrefix(key, "ROOM_") {
			conv.mu.RLock()
			joined := conv.Joined
			conv.mu.RUnlock()
			if joined {
				builder.WriteString(fmt.Sprintf("  - %s\n", conv.ID))
				found = true
			}
		}
	}

	if !found {
		builder.WriteString("  You have not joined any rooms.")
	}
	c.printToScreen(builder.String())
}

func (c *Client) printToScreen(msg string) {
	fmt.Printf("\r%s\n", msg)
	c.prompt()
}

func (c *Client) prompt() {
	c.mu.RLock()
	prompt := "[Auth] > "
	if c.AuthInfo != nil {
		if c.currentConversation != nil {
			prompt = fmt.Sprintf("[%s] > ", c.currentConversation.ID)
		} else {
			prompt = "[Lobby] > "
		}
	}
	c.mu.RUnlock()
	fmt.Print(prompt)
}

func (c *Client) notifyOrUpdate(name, convType, message string) {
	key := convType + "_" + name
	c.mu.RLock()
	isCurrent := false
	if c.currentConversation != nil {
		currentKey := c.currentConversation.Type + "_" + c.currentConversation.ID
		isCurrent = (currentKey == key)
	}
	inLobby := c.currentConversation == nil
	c.mu.RUnlock()

	if isCurrent {
		c.printToScreen(message)
	} else if inLobby {
		c.printToScreen(fmt.Sprintf("New message in %s (%s)", name, convType))
	}
}

func (c *Client) printHelp() {
	fmt.Println("--- ShellTalk Help ---")
	fmt.Println("  /help                  - Show this help message")
	fmt.Println("  /list                  - List all available rooms")
	fmt.Println("  /myrooms               - List rooms you have joined")
	fmt.Println("  /create <name> <pass>  - Create a new room")
	fmt.Println("  /join <name> <pass>    - Join a room by its name")
	fmt.Println("  /leave <name>          - Leave a room by its name")
	fmt.Println("  /members <name>        - List members of a room")
	fmt.Println("  /switch dm <nickname>  - Switch to a DM conversation")
	fmt.Println("  /switch room <name>    - Switch to a room conversation")
	fmt.Println("  /exit                  - Exit the current conversation to the lobby")
}

func (conv *Conversation) addHistory(msg string) {
	conv.mu.Lock()
	defer conv.mu.Unlock()
	conv.History = append(conv.History, msg)
}

func clearScreen() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		fmt.Print(strings.Repeat("\n", 50))
	}
}
