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
