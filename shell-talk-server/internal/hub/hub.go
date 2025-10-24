package hub

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"shell-talk-server/internal/domain"
	"shell-talk-server/internal/service"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// ClientRequest bundles a client with their incoming message.
type ClientRequest struct {
	Client  *Client
	Message domain.WebSocketMessage
}

// Hub maintains the set of active clients and broadcasts messages.
type Hub struct {
	connections          map[*Client]bool
	authenticatedClients map[uuid.UUID]*Client
	messages             chan *ClientRequest
	register             chan *Client
	unregister           chan *Client
	userService          service.IUserService
	roomService          service.IRoomService
	messageRepo          service.IMessageRepository
}

func NewHub(userService service.IUserService, roomService service.IRoomService, messageRepo service.IMessageRepository) *Hub {
	return &Hub{
		connections:          make(map[*Client]bool),
		authenticatedClients: make(map[uuid.UUID]*Client),
		messages:             make(chan *ClientRequest),
		register:             make(chan *Client),
		unregister:           make(chan *Client),
		userService:          userService,
		roomService:          roomService,
		messageRepo:          messageRepo,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.connections[client] = true
		case client := <-h.unregister:
			if _, ok := h.connections[client]; ok {
				if client.AuthInfo != nil {
					delete(h.authenticatedClients, client.AuthInfo.UserID)
				}
				delete(h.connections, client)
				close(client.Send)
			}
		case request := <-h.messages:
			h.handleMessage(request)
		}
	}
}

func (h *Hub) ServeWs(conn *websocket.Conn) {
	client := &Client{Hub: h, Conn: conn, Send: make(chan []byte, 256)}
	h.register <- client
	go client.writePump()
	go client.readPump()
}

func (h *Hub) handleMessage(req *ClientRequest) {
	switch req.Message.Type {
	case "register":
		h.handleRegister(req)
		return
	case "login":
		h.handleLogin(req)
		return
	}

	if req.Client.AuthInfo == nil {
		req.Client.sendSystemMessage("error_message", "Authentication required.")
		return
	}

	switch req.Message.Type {
	case "send_direct_message":
		h.handleSendDirectMessage(req)
	case "create_room":
		h.handleCreateRoom(req)
	case "join_room":
		h.handleJoinRoom(req)
	case "leave_room":
		h.handleLeaveRoom(req)
	case "list_rooms":
		h.handleListRooms(req)
	case "list_members":
		h.handleListMembers(req)
	case "send_room_message":
		h.handleSendRoomMessage(req)
	default:
		req.Client.sendSystemMessage("error_message", fmt.Sprintf("Unknown message type: %s", req.Message.Type))
	}
}

// ... Auth handlers ...
func (h *Hub) handleRegister(req *ClientRequest) {
	var payload domain.RegisterPayload
	if err := parsePayload(req.Message.Payload, &payload); err != nil {
		req.Client.sendSystemMessage("error_message", "Invalid register payload.")
		return
	}
	user, err := h.userService.Register(payload.Nickname, payload.Password)
	if err != nil {
		req.Client.sendSystemMessage("error_message", fmt.Sprintf("Registration failed: %v", err))
		return
	}
	h.authenticateClient(req.Client, user)
}

func (h *Hub) handleLogin(req *ClientRequest) {
	var payload domain.LoginPayload
	if err := parsePayload(req.Message.Payload, &payload); err != nil {
		req.Client.sendSystemMessage("error_message", "Invalid login payload.")
		return
	}
	user, err := h.userService.Login(payload.Nickname, payload.Password)
	if err != nil {
		req.Client.sendSystemMessage("error_message", fmt.Sprintf("Login failed: %v", err))
		return
	}
	h.authenticateClient(req.Client, user)
}

func (h *Hub) authenticateClient(client *Client, user *domain.User) {
	if existingClient, ok := h.authenticatedClients[user.ID]; ok {
		existingClient.sendSystemMessage("error_message", "You have been logged in from another location.")
		close(existingClient.Send)
	}
	client.AuthInfo = &Auth{UserID: user.ID, Nickname: user.Nickname}
	h.authenticatedClients[user.ID] = client
	loginSuccessPayload := domain.LoginSuccessPayload{UserID: user.ID, Nickname: user.Nickname}
	msg, _ := json.Marshal(domain.WebSocketMessage{Type: "login_success", Payload: loginSuccessPayload})
	client.Send <- msg
}

// --- Message Handlers ---

func (h *Hub) handleSendDirectMessage(req *ClientRequest) {
	var payload domain.SendDirectMessagePayload
	if err := parsePayload(req.Message.Payload, &payload); err != nil {
		req.Client.sendSystemMessage("error_message", "Invalid DM payload.")
		return
	}
	recipientUser, err := h.userService.GetUserByNickname(payload.RecipientNickname)
	if err != nil || recipientUser == nil {
		req.Client.sendSystemMessage("error_message", fmt.Sprintf("User '%s' not found.", payload.RecipientNickname))
		return
	}
	convoID := generateDMConversationID(req.Client.AuthInfo.UserID, recipientUser.ID)
	chatMsg := &domain.ChatMessage{ConversationID: convoID, SenderID: req.Client.AuthInfo.UserID.String(), SenderNickname: req.Client.AuthInfo.Nickname, Content: payload.Content, Timestamp: time.Now()}
	h.messageRepo.SaveMessage(context.Background(), chatMsg)
	if recipientClient, ok := h.authenticatedClients[recipientUser.ID]; ok {
		dmPayload := domain.DirectMessagePayload{Sender: req.Client.AuthInfo.Nickname, Content: payload.Content, Timestamp: chatMsg.Timestamp}
		msg, _ := json.Marshal(domain.WebSocketMessage{Type: "new_direct_message", Payload: dmPayload})
		recipientClient.Send <- msg
	}
}

func (h *Hub) handleCreateRoom(req *ClientRequest) {
	var payload domain.CreateRoomPayload
	if err := parsePayload(req.Message.Payload, &payload); err != nil {
		req.Client.sendSystemMessage("error_message", "Invalid create_room payload.")
		return
	}
	user := &domain.User{ID: req.Client.AuthInfo.UserID, Nickname: req.Client.AuthInfo.Nickname}
	room, err := h.roomService.CreateRoom(payload.Name, payload.Password, user)
	if err != nil {
		req.Client.sendSystemMessage("error_message", fmt.Sprintf("Failed to create room: %v", err))
		return
	}
	joinSuccessPayload := domain.JoinSuccessPayload{RoomID: room.ID.String(), RoomName: room.Name}
	msg, _ := json.Marshal(domain.WebSocketMessage{Type: "join_success", Payload: joinSuccessPayload})
	req.Client.Send <- msg
}

func (h *Hub) handleJoinRoom(req *ClientRequest) {
	var payload domain.JoinRoomPayload
	if err := parsePayload(req.Message.Payload, &payload); err != nil {
		req.Client.sendSystemMessage("error_message", "Invalid join_room payload.")
		return
	}
	user := &domain.User{ID: req.Client.AuthInfo.UserID}
	room, err := h.roomService.JoinRoom(payload.RoomName, payload.Password, user)
	if err != nil {
		req.Client.sendSystemMessage("error_message", fmt.Sprintf("Failed to join room: %v", err))
		return
	}
	joinSuccessPayload := domain.JoinSuccessPayload{RoomID: room.ID.String(), RoomName: room.Name}
	msg, _ := json.Marshal(domain.WebSocketMessage{Type: "join_success", Payload: joinSuccessPayload})
	req.Client.Send <- msg
}

func (h *Hub) handleLeaveRoom(req *ClientRequest) {
	var payload domain.LeaveRoomPayload
	if err := parsePayload(req.Message.Payload, &payload); err != nil {
		req.Client.sendSystemMessage("error_message", "Invalid leave_room payload.")
		return
	}
	user := &domain.User{ID: req.Client.AuthInfo.UserID}
	room, err := h.roomService.LeaveRoom(payload.RoomName, user)
	if err != nil {
		req.Client.sendSystemMessage("error_message", fmt.Sprintf("Failed to leave room: %v", err))
		return
	}
	leaveSuccessPayload := domain.LeaveSuccessPayload{RoomID: room.ID.String()}
	msg, _ := json.Marshal(domain.WebSocketMessage{Type: "leave_success", Payload: leaveSuccessPayload})
	req.Client.Send <- msg
}

func (h *Hub) handleListRooms(req *ClientRequest) {
	rooms, err := h.roomService.ListRooms()
	if err != nil {
		req.Client.sendSystemMessage("error_message", "Failed to retrieve room list.")
		return
	}
	roomInfos := make([]domain.RoomInfo, len(rooms))
	for i, r := range rooms {
		roomInfos[i] = domain.RoomInfo{ID: r.ID.String(), Name: r.Name}
	}
	payload := domain.RoomListPayload{Rooms: roomInfos}
	msg, _ := json.Marshal(domain.WebSocketMessage{Type: "room_list", Payload: payload})
	req.Client.Send <- msg
}

func (h *Hub) handleListMembers(req *ClientRequest) {
	var payload domain.ListMembersPayload
	if err := parsePayload(req.Message.Payload, &payload); err != nil {
		req.Client.sendSystemMessage("error_message", "Invalid list_members payload.")
		return
	}
	members, err := h.roomService.GetRoomMembers(payload.RoomName)
	if err != nil {
		req.Client.sendSystemMessage("error_message", fmt.Sprintf("Failed to get members for room '%s': %v", payload.RoomName, err))
		return
	}
	membersPayload := domain.RoomMembersPayload{RoomName: payload.RoomName, Members: members}
	msg, _ := json.Marshal(domain.WebSocketMessage{Type: "room_members", Payload: membersPayload})
	req.Client.Send <- msg
}

func (h *Hub) handleSendRoomMessage(req *ClientRequest) {
	var payload domain.SendRoomMessagePayload
	if err := parsePayload(req.Message.Payload, &payload); err != nil {
		req.Client.sendSystemMessage("error_message", "Invalid room message payload.")
		return
	}

	user := &domain.User{ID: req.Client.AuthInfo.UserID}
	isMember, err := h.roomService.IsRoomMember(payload.RoomName, user)
	if err != nil || !isMember {
		req.Client.sendSystemMessage("error_message", fmt.Sprintf("You are not a member of room '%s'.", payload.RoomName))
		return
	}

	room, err := h.roomService.GetRoomByName(payload.RoomName)
	if err != nil || room == nil {
		req.Client.sendSystemMessage("error_message", "Room not found.")
		return
	}

	// Save to DB
	chatMsg := &domain.ChatMessage{ConversationID: room.ID.String(), SenderID: req.Client.AuthInfo.UserID.String(), SenderNickname: req.Client.AuthInfo.Nickname, Content: payload.Content, Timestamp: time.Now()}
	h.messageRepo.SaveMessage(context.Background(), chatMsg)

	// Broadcast to online members
	roomMsgPayload := domain.RoomMessagePayload{RoomName: payload.RoomName, SenderNickname: req.Client.AuthInfo.Nickname, Content: payload.Content, Timestamp: chatMsg.Timestamp}
	msg, _ := json.Marshal(domain.WebSocketMessage{Type: "room_message", Payload: roomMsgPayload})

	memberIDs, err := h.roomService.GetRoomMemberIDs(room.Name)
	if err != nil {
		return
	}

	for _, memberID := range memberIDs {
		if onlineClient, ok := h.authenticatedClients[memberID]; ok {
			onlineClient.Send <- msg
		}
	}
}

// --- Helper Functions ---

func parsePayload(payload interface{}, result interface{}) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return errors.New("failed to marshal payload")
	}
	return json.Unmarshal(payloadBytes, result)
}

func generateDMConversationID(userID1, userID2 uuid.UUID) string {
	ids := []string{userID1.String(), userID2.String()}
	sort.Strings(ids)
	return strings.Join(ids, "_")
}
