package network

// WebSocketMessage WebSocketMessage는 서버와 통신하는 표준 포맷입니다.
// (서버의 domain/message.go와 일치해야 함)
type WebSocketMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// SendDirectMessagePayload SendDirectMessagePayload는 'send_direct_message' 요청의 페이로드입니다.
type SendDirectMessagePayload struct {
	Recipient string `json:"recipient"`
	Content   string `json:"content"`
}

// CreateRoomPayload is the payload for the 'create_room' message.
type CreateRoomPayload struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

// JoinRoomPayload is the payload for the 'join_room' message.
type JoinRoomPayload struct {
	RoomID   string `json:"room_id"`
	Password string `json:"password"`
}

// LeaveRoomPayload is the payload for the 'leave_room' message.
type LeaveRoomPayload struct {
	RoomID string `json:"room_id"`
}

// SendRoomMessagePayload is the payload for the 'send_room_message' message.
type SendRoomMessagePayload struct {
	RoomID  string `json:"room_id"`
	Content string `json:"content"`
}

// RoomInfo represents basic information about a room.
type RoomInfo struct {
	ID   string `json:"id"`
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
