package domain

import "time"

// WebSocketMessage WebSocketMessage는 클라이언트와 서버 간의 표준 통신 포맷입니다.
type WebSocketMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// SendDirectMessagePayload SendDirectMessagePayload는 'send_direct_message' 요청의 페이로드입니다.
type SendDirectMessagePayload struct {
	Recipient string `json:"recipient"` // 메시지 받을 사람의 닉네임
	Content   string `json:"content"`
}

// DirectMessagePayload DirectMessagePayload는 'new_direct_message' 타입의 페이로드입니다.
type DirectMessagePayload struct {
	Sender    string    `json:"sender"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// SystemPayload SystemPayload는 'system_message' 또는 'error_message' 타입의 페이로드입니다.
type SystemPayload struct {
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}
