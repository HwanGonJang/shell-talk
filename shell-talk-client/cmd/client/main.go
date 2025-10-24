package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"shell-talk-client/internal/config"
	"shell-talk-client/internal/network"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "shell-talk-client",
		Short: "ShellTalk CLI Client",
		Run:   runClient,
	}

	cobra.OnInitialize(config.LoadConfig)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runClient(cmd *cobra.Command, args []string) {
	serverURL := config.Cfg.Server.URL
	netClient := network.NewClient()

	if err := netClient.Connect(serverURL); err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}

	// Start the authentication flow
	authenticated := false
	go func() {
		for auth := range netClient.AuthCh {
			if auth {
				authenticated = true
				// Once authenticated, the main input handler will take over
				return
			}
		}
	}()

	reader := bufio.NewReader(os.Stdin)

	for !authenticated {
		fmt.Println("\nPlease login or register.")
		fmt.Println("Usage: /login <nickname> <password> | /register <nickname> <password>")
		fmt.Print("> ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		parts := strings.Split(input, " ")

		if len(parts) != 3 {
			fmt.Println("[ERROR] Invalid command format.")
			continue
		}

		command, nickname, password := parts[0], parts[1], parts[2]

		switch command {
		case "/login":
			netClient.Send <- network.WebSocketMessage{
				Type:    "login",
				Payload: network.LoginPayload{Nickname: nickname, Password: password},
			}
		case "/register":
			netClient.Send <- network.WebSocketMessage{
				Type:    "register",
				Payload: network.RegisterPayload{Nickname: nickname, Password: password},
			}
		default:
			fmt.Println("[ERROR] Invalid command. Use /login or /register.")
			continue
		}
		// Wait a moment for the server to respond
		// In a real app, you'd handle this more gracefully
		time.Sleep(1 * time.Second)
	}

	// Once authenticated, start the main input handler
	netClient.HandleStdin()
}
