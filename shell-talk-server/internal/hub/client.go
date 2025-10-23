package hub

import (
	"encoding/json"
	"log"
	"shell-talk-server/internal/domain"
	"time"

	"github.com/gorilla/websocket"
)

// Client WebSocket 연결과 Hub 간의 중개자입니다.
type Client struct {
	ID       string
	Nickname string
	Hub      *Hub // 중앙 Hub에 대한 참조
	Conn     *websocket.Conn
	Send     chan []byte // 메시지를 받는 채널
}

// readPump 클라이언트에서 WebSocket을 통해 메시지를 읽습니다.
func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	for {
		// JSON 메시지 읽기
		var req domain.WebSocketMessage
		err := c.Conn.ReadJSON(&req)
		if err != nil {
			log.Printf("readPump ReadJSON error (client: %s): %v", c.Nickname, err)
			break
		}

		// 메시지 타입에 따라 처리
		switch req.Type {
		case "send_direct_message":
			var payload domain.SendDirectMessagePayload
			payloadBytes, _ := json.Marshal(req.Payload)
			if err := json.Unmarshal(payloadBytes, &payload); err != nil {
				log.Printf("Invalid DM payload from %s: %v", c.Nickname, err)
				continue
			}
			dmRequest := &DirectMessage{
				SenderNickname:    c.Nickname,
				RecipientNickname: payload.Recipient,
				Content:           payload.Content,
			}
			c.Hub.directMessage <- dmRequest
		case "create_room":
			var payload domain.CreateRoomPayload
			payloadBytes, _ := json.Marshal(req.Payload)
			if err := json.Unmarshal(payloadBytes, &payload); err != nil {
				log.Printf("Invalid create_room payload from %s: %v", c.Nickname, err)
				continue
			}
			c.Hub.createRoom(payload.Name, payload.Password, c)
		case "join_room":
			var payload domain.JoinRoomPayload
			payloadBytes, _ := json.Marshal(req.Payload)
			if err := json.Unmarshal(payloadBytes, &payload); err != nil {
				log.Printf("Invalid join_room payload from %s: %v", c.Nickname, err)
				continue
			}
			c.Hub.joinRoom(payload.RoomID, payload.Password, c)
		case "leave_room":
			var payload domain.LeaveRoomPayload
			payloadBytes, _ := json.Marshal(req.Payload)
			if err := json.Unmarshal(payloadBytes, &payload); err != nil {
				log.Printf("Invalid leave_room payload from %s: %v", c.Nickname, err)
				continue
			}
			c.Hub.leaveRoom(payload.RoomID, c)
		case "send_room_message":
			var payload domain.SendRoomMessagePayload
			payloadBytes, _ := json.Marshal(req.Payload)
			if err := json.Unmarshal(payloadBytes, &payload); err != nil {
				log.Printf("Invalid send_room_message payload from %s: %v", c.Nickname, err)
				continue
			}
			c.Hub.handleRoomMessage(payload.RoomID, payload.Content, c)
		case "list_rooms":
			c.Hub.listRooms <- c
		}
	}
}

// writePump Send 채널의 메시지를 클라이언트의 WebSocket으로 보냅니다.
func (c *Client) writePump() {
	defer func() {
		c.Conn.Close()
	}()

	for message := range c.Send {
		err := c.Conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Printf("writePump error (client: %s): %v", c.Nickname, err)
			return
		}
	}
}

// sendSystemMessage 특정 시스템 메시지를 클라이언트에게 전송합니다.
func (c *Client) sendSystemMessage(msgType string, content string) {
	payload := domain.SystemPayload{
		Content:   content,
		Timestamp: time.Now(),
	}
	respMsg := domain.WebSocketMessage{
		Type:    msgType,
		Payload: payload,
	}
	jsonMsg, err := json.Marshal(respMsg)
	if err == nil {
		// send 채널이 닫혔을 수도 있으므로 안전하게 전송
		// (닉네임 중복 시 채널이 바로 닫힘)
		select {
		case c.Send <- jsonMsg:
		default:
			log.Printf("Could not send system message to %s (channel closed)", c.Nickname)
		}
	}
}
