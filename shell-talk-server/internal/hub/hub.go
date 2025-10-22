package hub

import (
	"encoding/json"
	"fmt"
	"log"
	"shell-talk-server/internal/domain"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// DirectMessage 1:1 메시지 전송 요청을 나타냅니다.
type DirectMessage struct {
	SenderNickname    string
	RecipientNickname string
	Content           string
}

// Hub 모든 활성화된 클라이언트를 관리하고 메시지를 중계합니다.
type Hub struct {
	// 닉네임을 키로 사용하여 클라이언트를 저장합니다.
	clients    map[string]*Client
	clientsMu  sync.RWMutex // clients 맵을 위한 Mutex
	register   chan *Client
	unregister chan *Client

	// DM 처리를 위한 채널
	directMessage chan *DirectMessage
}

// NewHub 새 Hub를 생성합니다.
func NewHub() *Hub {
	return &Hub{
		clients:       make(map[string]*Client),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		directMessage: make(chan *DirectMessage),
	}
}

// Run Hub의 메인 이벤트 루프를 실행합니다.
func (h *Hub) Run() {
	log.Println("Hub is now running...")
	for {
		select {
		case client := <-h.register:
			// 클라이언트 등록
			if !h.registerClient(client) {
				// 닉네임 중복
				client.sendSystemMessage("error_message", "이미 사용 중인 닉네임입니다.")
				close(client.Send) // 클라이언트 종료 처리
			} else {
				client.sendSystemMessage("system_message", "서버에 성공적으로 연결되었습니다.")
			}

		case client := <-h.unregister:
			// 클라이언트 등록 해제
			h.unregisterClient(client)

		case dm := <-h.directMessage:
			// 1:1 메시지 처리
			h.handleDirectMessage(dm)
		}
	}
}

// HandleNewClient handler로부터 새 연결을 받아 클라이언트를 생성하고 실행합니다.
func (h *Hub) HandleNewClient(conn *websocket.Conn, nickname string) {
	// 1. 새 클라이언트 생성
	client := &Client{
		ID:       uuid.NewString(),
		Nickname: nickname,
		Hub:      h,
		Conn:     conn,
		Send:     make(chan []byte, 256),
	}

	// 2. Hub의 Run() 루프에 등록 요청
	h.register <- client

	// 3. 클라이언트의 고루틴 실행
	go client.writePump()
	go client.readPump()
}

// registerClient Hub에 클라이언트를 등록합니다. 닉네임 중복 시 false 반환.
func (h *Hub) registerClient(client *Client) bool {
	h.clientsMu.Lock()
	defer h.clientsMu.Unlock()

	if _, exists := h.clients[client.Nickname]; exists {
		log.Printf("Nickname conflict: %s", client.Nickname)
		return false // 닉네임 중복
	}

	h.clients[client.Nickname] = client
	log.Printf("Client registered: %s (ID: %s)", client.Nickname, client.ID)
	return true
}

// unregisterClient Hub에서 클라이언트를 등록 해제합니다.
func (h *Hub) unregisterClient(client *Client) {
	h.clientsMu.Lock()
	defer h.clientsMu.Unlock()

	if _, ok := h.clients[client.Nickname]; ok {
		delete(h.clients, client.Nickname)
		close(client.Send)
		log.Printf("Client unregistered: %s", client.Nickname)
	}
}

// handleDirectMessage DM을 찾아 수신자에게 전송합니다.
func (h *Hub) handleDirectMessage(dm *DirectMessage) {
	h.clientsMu.RLock()
	recipient, ok := h.clients[dm.RecipientNickname]
	h.clientsMu.RUnlock()

	// 수신자가 온라인 상태인지 확인
	if !ok {
		log.Printf("DM Failed: Recipient '%s' not found.", dm.RecipientNickname)

		// 발신자에게 에러 메시지 전송
		h.clientsMu.RLock()
		sender, senderOk := h.clients[dm.SenderNickname]
		h.clientsMu.RUnlock()
		if senderOk {
			sender.sendSystemMessage("error_message", fmt.Sprintf("'%s'님을 찾을 수 없습니다.", dm.RecipientNickname))
		}
		return
	}

	// 수신자에게 보낼 메시지 페이로드 생성
	payload := domain.DirectMessagePayload{
		Sender:    dm.SenderNickname,
		Content:   dm.Content,
		Timestamp: time.Now(),
	}
	respMsg := domain.WebSocketMessage{
		Type:    "new_direct_message",
		Payload: payload,
	}
	jsonMsg, err := json.Marshal(respMsg)
	if err != nil {
		log.Printf("DM Marshal error: %v", err)
		return
	}

	// 수신자의 Send 채널로 메시지 전송
	select {
	case recipient.Send <- jsonMsg:
		log.Printf("DM Sent: %s -> %s", dm.SenderNickname, dm.RecipientNickname)
	default:
		// 수신자의 채널이 꽉 찼거나 닫힌 경우
		log.Printf("DM Failed: Recipient '%s' channel is full or closed.", dm.RecipientNickname)
	}
}
