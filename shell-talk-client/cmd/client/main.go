package main

import (
	"fmt"
	"log"
	"os"
	"shell-talk-client/internal/config"
	"shell-talk-client/internal/network"

	"github.com/spf13/cobra"
)

var nickname string // 닉네임 플래그

func main() {
	var rootCmd = &cobra.Command{
		Use:   "shell-talk-client",
		Short: "ShellTalk (Simple CLI) 채팅 클라이언트",
		Run:   runClient,
	}

	cobra.OnInitialize(config.LoadConfig)

	rootCmd.Flags().StringVarP(&nickname, "nickname", "n", "", "채팅에서 사용할 닉네임 (필수)")
	rootCmd.MarkFlagRequired("nickname")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// runClient는 실제 클라이언트 로직을 실행합니다.
func runClient(cmd *cobra.Command, args []string) {
	// 1. 설정 로드
	serverURL := config.Cfg.Server.URL

	// 2. 네트워크 클라이언트 생성
	netClient := network.NewClient()

	// 3. 서버 연결 (내부적으로 read/write 고루틴 시작)
	if err := netClient.Connect(serverURL, nickname); err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}

	// 4. 메인 고루틴에서 터미널 입력 처리 시작
	//    이 함수는 무한 루프이므로 프로그램이 종료되지 않습니다.
	netClient.HandleStdin()
}
