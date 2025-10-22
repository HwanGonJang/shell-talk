package handler

import (
	"log"
	"net/http"
	"shell-talk-server/internal/hub"

	"github.com/gorilla/websocket"
)

// WebSocket 업그레이더 설정
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// 모든 오리진 허용 (개발용)
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// WebsocketHandler WebSocket 연결 요청을 처리합니다.
type WebsocketHandler struct {
	hub *hub.Hub
}

// NewWebsocketHandler 새 WebsocketHandler를 생성합니다.
func NewWebsocketHandler(h *hub.Hub) *WebsocketHandler {
	return &WebsocketHandler{
		hub: h,
	}
}

// HandleConnection 핸들러 (GET /ws)
func (h *WebsocketHandler) HandleConnection(w http.ResponseWriter, r *http.Request) {
	// 닉네임은 쿼리 파라미터로 받습니다 (e.g., /ws?nickname=gopher)
	nickname := r.URL.Query().Get("nickname")

	if nickname == "" {
		http.Error(w, "Nickname is required", http.StatusBadRequest)
		return
	}

	// 1. WebSocket 연결 업그레이드
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrade error: %v", err)
		return
	}

	// 2. 새 클라이언트 생성
	h.hub.HandleNewClient(conn, nickname)
}
