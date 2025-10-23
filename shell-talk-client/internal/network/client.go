package network

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Conversation holds the state of a single chat (DM or Room).
type Conversation struct {
	ID      string // Nickname for DMs, RoomID for rooms
	Name    string
	Type    string // "DM" or "ROOM"
	Joined  bool   // Only relevant for rooms
	History []string
	mu      sync.RWMutex
}

// Client manages the WebSocket connection and the user interface state.
type Client struct {
	Conn                *websocket.Conn
	Send                chan WebSocketMessage
	Nickname            string
	conversations       map[string]*Conversation
	currentConversation *Conversation
	mu                  sync.RWMutex
}

var uuidRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// NewClient creates a new network client.
func NewClient() *Client {
	return &Client{
		Send:          make(chan WebSocketMessage, 256),
		conversations: make(map[string]*Conversation),
	}
}

// Connect establishes a WebSocket connection to the server.
func (c *Client) Connect(serverURL string, nickname string) error {
	c.Nickname = nickname
	u, err := url.Parse(serverURL)
	if err != nil {
		return err
	}
	q := u.Query()
	q.Set("nickname", nickname)
	u.RawQuery = q.Encode()

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

// HandleStdin is the main loop for reading user input.
func (c *Client) HandleStdin() {
	reader := bufio.NewReader(os.Stdin)
	time.Sleep(500 * time.Millisecond) // Wait for initial messages
	c.updateView()

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
		c.updateView()
		return
	case "/create":
		if len(parts) < 3 {
			c.printToScreen("[ERROR] Usage: /create <room_name> <password>")
		} else {
			c.Send <- WebSocketMessage{Type: "create_room", Payload: CreateRoomPayload{Name: parts[1], Password: strings.Join(parts[2:], " ")}}
		}
	case "/join":
		if len(parts) < 3 {
			c.printToScreen("[ERROR] Usage: /join <room_id> <password>")
		} else {
			c.Send <- WebSocketMessage{Type: "join_room", Payload: JoinRoomPayload{RoomID: parts[1], Password: strings.Join(parts[2:], " ")}}
		}
	case "/leave":
		if len(parts) < 2 {
			c.printToScreen("[ERROR] Usage: /leave <room_id>")
		} else {
			c.Send <- WebSocketMessage{Type: "leave_room", Payload: LeaveRoomPayload{RoomID: parts[1]}}
		}
	case "/list":
		c.Send <- WebSocketMessage{Type: "list_rooms"}
	case "/myrooms":
		c.listMyRooms()
		return
	case "/switch":
		if len(parts) < 2 {
			c.printToScreen("[ERROR] Usage: /switch <context_id>")
			return
		}
		if err := c.switchConversation(parts[1]); err != nil {
			c.printToScreen(fmt.Sprintf("[ERROR] %s", err.Error()))
		} else {
			c.updateView()
		}
		return
	case "/exit":
		c.mu.Lock()
		c.currentConversation = nil
		c.mu.Unlock()
		c.updateView()
		return
	default:
		c.printToScreen(fmt.Sprintf("[ERROR] Unknown command: %s", command))
	}

	// For commands that don't manage their own UI update, just show a prompt
	c.prompt()
}

func (c *Client) handleChatMessage(input string) {
	c.mu.RLock()
	conv := c.currentConversation
	c.mu.RUnlock()

	if conv == nil {
		c.printToScreen("[ERROR] Not in a conversation. Use /switch <id> or /list.")
		return
	}

	var msgType string
	switch conv.Type {
	case "DM":
		msgType = "send_direct_message"
		c.Send <- WebSocketMessage{Type: msgType, Payload: SendDirectMessagePayload{Recipient: conv.ID, Content: input}}
	case "ROOM":
		msgType = "send_room_message"
		c.Send <- WebSocketMessage{Type: msgType, Payload: SendRoomMessagePayload{RoomID: conv.ID, Content: input}}
	}

	formattedMsg := fmt.Sprintf("[%s] [Me]: %s", time.Now().Format("15:04:05"), input)
	conv.addHistory(formattedMsg)
	c.updateView()
}

func (c *Client) handleServerMessage(msg WebSocketMessage) {
	payloadBytes, _ := json.Marshal(msg.Payload)
	timestamp := time.Now().Format("15:04:05")
	var output, contextID, convType, convName string

	switch msg.Type {
	case "new_direct_message":
		var payload map[string]interface{}
		_ = json.Unmarshal(payloadBytes, &payload)
		sender := payload["sender"].(string)
		content := payload["content"].(string)
		contextID, convType, convName = sender, "DM", sender
		output = fmt.Sprintf("[%s] [%s]: %s", timestamp, sender, content)

	case "room_message":
		var payload map[string]interface{}
		_ = json.Unmarshal(payloadBytes, &payload)
		roomID := payload["room_id"].(string)
		sender := payload["sender"].(string)
		content := payload["content"].(string)
		contextID, convType, convName = roomID, "ROOM", roomID
		output = fmt.Sprintf("[%s] [%s]: %s", timestamp, sender, content)

	case "join_success":
		var payload JoinSuccessPayload
		_ = json.Unmarshal(payloadBytes, &payload)
		conv := c.getOrCreateConversation(payload.RoomID, "ROOM", payload.RoomName)
		conv.mu.Lock()
		conv.Joined = true
		conv.Name = payload.RoomName // Ensure name is updated
		conv.mu.Unlock()
		c.printToScreen(fmt.Sprintf("[SYSTEM] Successfully joined room: %s", payload.RoomName))
		return

	case "leave_success":
		var payload LeaveSuccessPayload
		_ = json.Unmarshal(payloadBytes, &payload)
		c.mu.RLock()
		conv, exists := c.conversations[payload.RoomID]
		c.mu.RUnlock()
		if exists {
			conv.mu.Lock()
			conv.Joined = false
			conv.mu.Unlock()
		}
		c.printToScreen(fmt.Sprintf("[SYSTEM] Successfully left room: %s", payload.RoomID))
		return

	case "system_message", "error_message":
		var payload map[string]interface{}
		_ = json.Unmarshal(payloadBytes, &payload)
		content := payload["content"].(string)
		prefix := "[SYSTEM]"
		if msg.Type == "error_message" {
			prefix = "[SERVER ERROR]"
		}
		output = fmt.Sprintf("[%s] %s: %s", timestamp, prefix, content)
		c.printToScreen(output)
		return

	case "room_list":
		var payload RoomListPayload
		_ = json.Unmarshal(payloadBytes, &payload)
		var builder strings.Builder
		builder.WriteString(fmt.Sprintf("\r[%s] [All Rooms]:\n", timestamp))
		if len(payload.Rooms) == 0 {
			builder.WriteString("  No rooms available.")
		} else {
			for _, room := range payload.Rooms {
				builder.WriteString(fmt.Sprintf("  - %s (ID: %s)\n", room.Name, room.ID))
				c.getOrCreateConversation(room.ID, "ROOM", room.Name)
			}
		}
		c.printToScreen(builder.String())
		return

	default:
		c.printToScreen(fmt.Sprintf("[%s] [UNKNOWN]: %v", timestamp, msg))
		return
	}

	if contextID != "" {
		conv := c.getOrCreateConversation(contextID, convType, convName)
		conv.addHistory(output)

		c.mu.RLock()
		isCurrent := c.currentConversation != nil && c.currentConversation.ID == contextID
		c.mu.RUnlock()

		if isCurrent {
			c.updateView()
		} else {
			notification := fmt.Sprintf("New message from %s", conv.Name)
			c.printToScreen(notification)
		}
	}
}

// getOrCreateConversation_internal is a helper that assumes the caller holds the necessary lock.
func (c *Client) getOrCreateConversation_internal(id, convType, name string) *Conversation {
	if conv, exists := c.conversations[id]; exists {
		if conv.Name == conv.ID && name != id {
			conv.Name = name
		}
		return conv
	}
	conv := &Conversation{ID: id, Type: convType, Name: name, History: []string{}}
	c.conversations[id] = conv
	return conv
}

// getOrCreateConversation acquires a lock and creates a conversation.
func (c *Client) getOrCreateConversation(id, convType, name string) *Conversation {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.getOrCreateConversation_internal(id, convType, name)
}

func (c *Client) switchConversation(contextID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	conv, exists := c.conversations[contextID]
	if exists {
		conv.mu.RLock()
		convType, convJoined := conv.Type, conv.Joined
		conv.mu.RUnlock()

		if convType == "ROOM" && !convJoined {
			return fmt.Errorf("you have not joined room '%s'. Use /join first", conv.Name)
		}
		c.currentConversation = conv
		return nil
	} else {
		if uuidRegex.MatchString(contextID) {
			return fmt.Errorf("unknown room ID: %s. Use /list to see available rooms", contextID)
		}
		c.currentConversation = c.getOrCreateConversation_internal(contextID, "DM", contextID)
		return nil
	}
}

// updateView is the single source of truth for rendering the UI.
func (c *Client) updateView() {
	clearScreen()
	c.mu.RLock()
	conv := c.currentConversation
	c.mu.RUnlock()

	if conv == nil {
		// We are in the lobby
		c.printHelp()
	} else {
		// We are in a conversation
		fmt.Printf("--- Conversation with %s (%s) ---\n", conv.Name, conv.Type)
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
	for _, conv := range c.conversations {
		conv.mu.RLock()
		if conv.Type == "ROOM" && conv.Joined {
			builder.WriteString(fmt.Sprintf("  - %s (ID: %s)\n", conv.Name, conv.ID))
			found = true
		}
		conv.mu.RUnlock()
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
	defer c.mu.RUnlock()

	prompt := "[Lobby] > "
	if c.currentConversation != nil {
		prompt = fmt.Sprintf("[%s] > ", c.currentConversation.Name)
	}
	fmt.Print(prompt)
}

func (c *Client) printHelp() {
	fmt.Println("--- ShellTalk Help ---")
	fmt.Println("  /help                  - Show this help message")
	fmt.Println("  /list                  - List all available rooms")
	fmt.Println("  /myrooms               - List rooms you have joined")
	fmt.Println("  /create <name> <pass>  - Create a new room (and auto-join)")
	fmt.Println("  /join <id> <pass>      - Join a room")
	fmt.Println("  /leave <id>            - Leave a room")
	fmt.Println("  /switch <id>           - Switch to a conversation (room ID or user nickname)")
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
		// Fallback if clear command fails
		fmt.Print("\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n")
	}
}