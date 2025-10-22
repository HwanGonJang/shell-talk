package network

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// Client Client는 WebSocket 연결을 관리합니다.
type Client struct {
	Conn *websocket.Conn
	Send chan WebSocketMessage // 메시지 전송 채널 (동시 쓰기 방지)
}

// NewClient NewClient는 새 네트워크 클라이언트를 생성합니다.
func NewClient() *Client {
	return &Client{
		Send: make(chan WebSocketMessage, 256),
	}
}

// Connect Connect는 서버에 WebSocket 연결을 시도합니다.
func (c *Client) Connect(serverURL string, nickname string) error {
	u, err := url.Parse(serverURL)
	if err != nil {
		return err
	}
	q := u.Query()
	q.Set("nickname", nickname)
	u.RawQuery = q.Encode()

	log.Printf("Connecting to %s...", u.String())
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}
	log.Println("Connection successful!")
	c.Conn = conn

	// 메시지 수신 및 송신 고루틴 시작
	go c.readPump()
	go c.writePump()

	return nil
}

// readPump 서버로부터 메시지를 읽어 stdout에 출력합니다.
func (c *Client) readPump() {
	defer c.Conn.Close()
	for {
		var msg WebSocketMessage
		err := c.Conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("Connection closed: %v", err)
			os.Exit(0) // 연결 끊기면 프로그램 종료
			return
		}

		// 수신한 메시지 파싱 및 출력
		c.handleServerMessage(msg)
	}
}

// writePump Send 채널의 메시지를 서버로 전송합니다. (동시 쓰기 방지)
func (c *Client) writePump() {
	defer c.Conn.Close()
	for msg := range c.Send {
		err := c.Conn.WriteJSON(msg)
		if err != nil {
			log.Printf("Write error: %v", err)
			return
		}
	}
}

// HandleStdin HandleStdin은 터미널 입력을 읽어 Send 채널로 보냅니다.
func (c *Client) HandleStdin() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Enter message (e.g., /dm Bob Hello!):")
	fmt.Print("> ")

	for {
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			fmt.Print("> ")
			continue
		}

		// 명령어 파싱
		if strings.HasPrefix(input, "/dm") {
			parts := strings.SplitN(input, " ", 3)
			if len(parts) < 3 {
				fmt.Printf("\r[ERROR] Invalid command format. Use: /dm [nickname] [message]\n")
			} else {
				wsMsg := WebSocketMessage{
					Type: "send_direct_message",
					Payload: SendDirectMessagePayload{
						Recipient: parts[1],
						Content:   parts[2],
					},
				}
				c.Send <- wsMsg
				// 내가 보낸 메시지도 로컬에 표시
				fmt.Printf("[%s] [Me -> %s]: %s\n", time.Now().Format("15:04:05"), parts[1], parts[2])
			}
		} else {
			fmt.Printf("\r[ERROR] Invalid command. Only /dm is supported.\n")
		}
		fmt.Print("> ")
	}
}

// handleServerMessage 서버 메시지를 파싱하여 콘솔에 이쁘게 출력합니다.
func (c *Client) handleServerMessage(msg WebSocketMessage) {
	payloadBytes, _ := json.Marshal(msg.Payload)
	var payload map[string]interface{}
	_ = json.Unmarshal(payloadBytes, &payload)

	timestamp := time.Now().Format("15:04:05")

	var output string

	switch msg.Type {
	case "new_direct_message":
		sender := payload["sender"].(string)
		content := payload["content"].(string)
		output = fmt.Sprintf("[%s] [DM from %s]: %s", timestamp, sender, content)

	case "system_message":
		content := payload["content"].(string)
		output = fmt.Sprintf("[%s] [SYSTEM]: %s", timestamp, content)

	case "error_message":
		content := payload["content"].(string)
		output = fmt.Sprintf("[%s] [SERVER ERROR]: %s", timestamp, content)

	default:
		output = fmt.Sprintf("[%s] [UNKNOWN]: %v", timestamp, msg)
	}

	// \n (줄바꿈), 메시지, \n (줄바꿈) 만 출력하고 프롬프트(>)를 제거합니다.
	// \r (캐리지 리턴)은 현재 입력 중인 줄을 지우고 새로 그리기 위해 사용합니다. (터미널 UI 트릭)
	fmt.Printf("\r%s\n> ", output)
}
