package main

import (
	"log"
	"net/http"
	"shell-talk-server/internal/handler"
	"shell-talk-server/internal/hub"

	"github.com/gorilla/mux"
)

func main() {
	// 1. 의존성 생성 (DI)
	hub := hub.NewHub()
	// Hub의 메인 루프를 고루틴으로 실행
	go hub.Run()

	wsHandler := handler.NewWebsocketHandler(hub)

	// 2. 라우터 설정
	r := mux.NewRouter()

	// WebSocket 라우트 (POST, GET API 제거)
	// /ws?nickname=myname 형식으로 접속
	r.HandleFunc("/ws", wsHandler.HandleConnection).Methods("GET")

	// 3. 서버 시작
	port := ":8080"
	log.Printf("Server starting on port %s (DM Only Mode)", port)
	if err := http.ListenAndServe(port, r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
