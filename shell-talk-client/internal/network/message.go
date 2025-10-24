package network

import (
	"github.com/google/uuid"
)

// WebSocketMessage is the standard communication format.
type WebSocketMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// --- Authentication Payloads ---

// RegisterPayload is the payload for the 'register' message.
type RegisterPayload struct {
	Nickname string `json:"nickname"`
	Password string `json:"password"`
}

// LoginPayload is the payload for the 'login' message.
type LoginPayload struct {
	Nickname string `json:"nickname"`
	Password string `json:"password"`
}

// LoginSuccessPayload is the payload for the 'login_success' message.
type LoginSuccessPayload struct {
	UserID   uuid.UUID `json:"user_id"`
	Nickname string    `json:"nickname"`
}

// --- DM & Room Message Payloads ---

// SendDirectMessagePayload is the payload for the 'send_direct_message' request.
type SendDirectMessagePayload struct {
	RecipientNickname string `json:"recipient_nickname"`
	Content           string `json:"content"`
}

// SendRoomMessagePayload is the payload for the 'send_room_message' message.
type SendRoomMessagePayload struct {
	RoomName string `json:"room_name"`
	Content  string `json:"content"`
}

// --- Room Management Payloads ---

// CreateRoomPayload is the payload for the 'create_room' message.
type CreateRoomPayload struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

// JoinRoomPayload is the payload for the 'join_room' message.
type JoinRoomPayload struct {
	RoomName string `json:"room_name"`
	Password string `json:"password"`
}

// LeaveRoomPayload is the payload for the 'leave_room' message.
type LeaveRoomPayload struct {
	RoomName string `json:"room_name"`
}

// ListMembersPayload is the payload for the 'list_members' message.
type ListMembersPayload struct {
	RoomName string `json:"room_name"`
}

// RoomMembersPayload is the payload for the 'room_members' message.
type RoomMembersPayload struct {
	RoomName string   `json:"room_name"`
	Members  []string `json:"members"` // List of nicknames
}

// RoomInfo represents basic information about a room.
type RoomInfo struct {
	ID   string `json:"id"` // This is the UUID
	Name string `json:"name"`
}

// RoomListPayload is the payload for the 'room_list' message.
type RoomListPayload struct {
	Rooms []RoomInfo `json:"rooms"`
}

// JoinSuccessPayload is the payload for the 'join_success' message.
type JoinSuccessPayload struct {
	RoomID   string `json:"room_id"`
	RoomName string `json:"room_name"`
}

// LeaveSuccessPayload is the payload for the 'leave_success' message.
type LeaveSuccessPayload struct {
	RoomID string `json:"room_id"`
}
