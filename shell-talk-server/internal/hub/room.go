package hub

import (
	"sync"

	"github.com/google/uuid"
)

// Room represents a chat room.
type Room struct {
	ID       string
	Name     string
	Password string // 4-digit numeric password
	clients  map[*Client]bool
	mu       sync.RWMutex
	Hub      *Hub
}

// NewRoom creates a new Room.
func NewRoom(name, password string, hub *Hub) *Room {
	return &Room{
		ID:       uuid.NewString(),
		Name:     name,
		Password: password,
		clients:  make(map[*Client]bool),
		Hub:      hub,
	}
}

// addClient adds a client to the room.
func (r *Room) addClient(client *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.clients[client] = true
}

// removeClient removes a client from the room.
func (r *Room) removeClient(client *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.clients[client]; ok {
		delete(r.clients, client)
	}
}

// hasClient checks if a client is in the room.
func (r *Room) hasClient(client *Client) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.clients[client]
	return ok
}

// broadcast sends a message to all clients in the room, except the sender.
func (r *Room) broadcast(message []byte, sender *Client) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for client := range r.clients {
		if client != sender {
			select {
			case client.Send <- message:
			default:
				// If the client's send channel is full, unregister the client.
				// This is a simple way to handle slow clients.
				r.Hub.unregister <- client
			}
		}
	}
}
