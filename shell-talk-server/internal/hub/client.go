package hub

import (
	"encoding/json"
	"log"
	"shell-talk-server/internal/domain"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	Hub      *Hub
	Conn     *websocket.Conn
	Send     chan []byte
	AuthInfo *Auth // Holds authenticated user info, nil if not authenticated
}

// Auth holds the authenticated user's data.
type Auth struct {
	UserID   uuid.UUID
	Nickname string
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs one readPump per connection. The readPump ensures
// that there is at most one reader on a connection by executing all
// reads from this goroutine. It does this by sending all messages
// to the hub for processing.
func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	for {
		var req domain.WebSocketMessage
		if err := c.Conn.ReadJSON(&req); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		// If client is not authenticated, only allow login or register messages
		if c.AuthInfo == nil {
			if req.Type != "login" && req.Type != "register" {
				c.sendSystemMessage("error_message", "Authentication required. Please /login or /register.")
				continue
			}
		}

		// Add client context to the request and send to hub
		request := &ClientRequest{
			Client:  c,
			Message: req,
		}
		c.Hub.messages <- request
	}
}

// writePump pumps messages from the hub to the websocket connection.
func (c *Client) writePump() {
	defer c.Conn.Close()
	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				// The hub closed the channel.
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		}
	}
}

// sendSystemMessage sends a system message to this client.
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
		select {
		case c.Send <- jsonMsg:
		default:
			log.Printf("Could not send system message to client (channel closed or full)")
		}
	}
}
