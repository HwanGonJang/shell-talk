package main

import (
	"log"
	"net/http"
	"shell-talk-server/internal/config"
	"shell-talk-server/internal/repository/postgres"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool { return true },
}

func main() {
	cfg := config.Load()

	// Run database migrations before starting the main app
	if err := postgres.RunMigrations(cfg.PostgresURL); err != nil {
		log.Fatalf("failed to run database migrations: %v", err)
	}
	log.Println("Database migrations completed successfully.")

	app, cleanup, err := InitializeApp()
	if err != nil {
		log.Fatalf("failed to initialize app: %v", err)
	}
	defer cleanup()

	go app.Hub.Run()

	r := mux.NewRouter()
	r.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("failed to upgrade connection: %v", err)
			return
		}
		app.Hub.ServeWs(conn)
	})

	log.Println("Server started on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
