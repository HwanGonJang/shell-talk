package domain

import (
	"time"

	"github.com/google/uuid"
)

// WebSocketMessage WebSocketMessage는 클라이언트와 서버 간의 표준 통신 포맷입니다.
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

// SendDirectMessagePayload SendDirectMessagePayload는 'send_direct_message' 요청의 페이로드입니다.
type SendDirectMessagePayload struct {
	RecipientNickname string `json:"recipient_nickname"`
	Content           string `json:"content"`
}

// DirectMessagePayload DirectMessagePayload는 'new_direct_message' 타입의 페이로드입니다.
type DirectMessagePayload struct {
	Sender    string    `json:"sender"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// SendRoomMessagePayload is the payload for the 'send_room_message' message.
type SendRoomMessagePayload struct {
	RoomName string `json:"room_name"`
	Content  string `json:"content"`
}

// RoomMessagePayload is the payload for the 'room_message' message.
type RoomMessagePayload struct {
	RoomName       string    `json:"room_name"`
	SenderNickname string    `json:"sender_nickname"`
	Content        string    `json:"content"`
	Timestamp      time.Time `json:"timestamp"`
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

// --- System & Error Payloads ---

// SystemPayload SystemPayload는 'system_message' 또는 'error_message' 타입의 페이로드입니다.
type SystemPayload struct {
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}
