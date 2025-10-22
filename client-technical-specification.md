네, 알겠습니다. 제공해주신 서버 기술 명세서의 구조와 스타일에 정확히 맞춰서 **클라이언트 기술 명세서**를 다시 작성해 드릴게요.

-----

# **ShellTalk 클라이언트(CLI) 기술 명세서**

## **1. 개요**

본 문서는 ShellTalk 서버와 통신하는 **터미널(CLI) 클라이언트**의 아키텍처, 코드 컨벤션, 통신 프로토콜 등 기술적인 표준을 정의한다. 목표는 사용자에게 **직관적인 TUI(Terminal User Interface)를 제공**하고, 서버와 안정적인 실시간 통신을 보장하는 것이다.

## **2. 기술 스택 및 주요 라이브러리**

  - **언어**: **Go** (최신 안정 버전)
  - **웹소켓**: **gorilla/websocket**
      - 이유: 서버와 동일한 라이브러리를 사용하여 일관성을 유지하고, 안정적인 클라이언트 측 연결을 관리함.
  - **CLI 프레임워크**: **cobra**
      - 이유: `git`, `kubectl` 등에서 사용하는 표준 CLI 라이브러리로, 명령어, 하위 명령어, 플래그(flag) 관리가 매우 용이함.
  - **TUI (Terminal UI)**: **Bubble Tea**
      - 이유: Elm 아키텍처(MVU)를 기반으로 하여, 실시간으로 변하는 채팅 UI의 상태(State)를 선언적이고 안전하게 관리하는 데 최적화됨.
  - **설정 관리**: **spf13/viper**
      - 이유: 서버 주소, 기본 닉네임 등 사용자 설정을 파일이나 환경 변수에서 손쉽게 관리할 수 있음.
  - **로깅**: **zerolog**
      - 이유: 클라이언트 측에서 발생하는 디버그 로그 및 에러를 구조화하여 파일로 기록하는 데 유용함.

## **3. 아키텍처**

**Elm 아키텍처 (Model-View-Update)** 패턴을 채택하여 TUI의 상태를 관리한다. 이는 Bubble Tea 라이브러리의 핵심 철학이다.

  - **핵심 원칙 (단방향 데이터 흐름)**: 모든 상태(State)는 단일 `Model`에 저장된다. 외부 이벤트(키 입력, 서버 메시지)는 `Update` 함수로 전달되어 새로운 `Model`을 반환하고, `View` 함수는 이 `Model`을 바탕으로 화면을 그린다.

  - **구성 요소 역할**:

      - **Model**: 애플리케이션의 모든 상태를 담고 있는 단일 Struct. (예: `메시지 목록`, `현재 입력 중인 텍스트`, `서버 연결 상태`, `사용자 목록`)
      - **Update**: 이벤트(`tea.Msg`)를 수신하여 `Model`을 갱신하는 유일한 통로. 사용자의 키 입력, 서버로부터 수신된 메시지 등을 처리한다.
      - **View**: 현재 `Model`의 상태를 바탕으로 터미널에 표시될 화면(UI)을 문자열로 렌더링한다.

## **4. 디렉토리 구조 및 패키지 역할**

Standard Go Project Layout을 기반으로 하며, TUI와 네트워크 로직을 분리한다.

```
shell-talk-client/
├── cmd/
│   └── client/
│       └── main.go         # 🚀 애플리케이션 시작점, Cobra 명령어 정의
├── configs/
│   └── config.yaml         # ⚙️ 서버 주소, 기본 닉네임 등 설정 파일
├── internal/
│   ├── tui/                # ❤️‍🔥 Bubble Tea (Model, View, Update) 로직
│   ├── network/            # 🔌 WebSocket 연결 및 메시지 송수신 관리
│   └── config/             # 🛠️ Viper 설정 로드
├── go.mod                  # 📦 의존성 관리
├── go.sum
└── Dockerfile              # 🐳 (선택 사항) 클라이언트 실행 환경
```

  - **`cmd/client/main.go`**: `cobra`를 사용해 CLI 명령어(예: `join`, `list`)를 정의하고, 플래그를 파싱하여 `tui`와 `network` 모듈을 초기화하고 실행한다.
  - **`internal/tui`**: 애플리케이션의 핵심. Bubble Tea의 `Model`, `View`, `Update` 함수를 구현하여 전체 UI와 상태를 관리한다.
  - **`internal/network`**: 서버와의 WebSocket 연결, 메시지 송신, 메시지 수신(리스닝)을 담당한다. 서버로부터 메시지를 수신하면, Bubble Tea가 이해할 수 있는 `tea.Msg` 형태로 변환하여 `tui`의 `Update` 함수로 전달한다.
  - **`internal/config`**: `viper`를 사용해 `configs/config.yaml` 파일이나 환경 변수에서 설정을 불러온다.

## **5. 통신 프로토콜**

### **5.1. HTTP REST API (채팅방 관리)**

  - 서버의 `Content-Type: application/json`을 준수한다.
  - 채팅방 참여 전, 서버 명세서에 정의된 API를 호출하여 필요한 정보를 얻는다.
      - `POST /api/rooms` (채팅방 생성 요청)
      - `GET /api/rooms` (채팅방 목록 조회)

### **5.2. WebSocket 메시지 프로토콜**

  - 서버 명세서에 정의된 **JSON 형식**의 메시지 프로토콜을 **준수**한다.

  - `type`과 `payload` 필드를 가진 메시지를 **송신**하고 **수신**한다.

  - **주요 메시지 타입 예시**:

      - **클라이언트 → 서버 (송신)**:
          - `send_public_message`: `{ "content": "..." }`
          - `send_direct_message`: `{ "recipient": "...", "content": "..." }`
          - `list_users`: `{}`
      - **서버 → 클라이언트 (수신)**:
          - `new_public_message`: `{ "sender": "...", "content": "...", "timestamp": "..." }`
          - `new_direct_message`: `{ "sender": "...", "content": "...", "timestamp": "..." }`
          - `system_message`: `{ "content": "...", "timestamp": "..." }`
          - `user_list`: `{ "users": [...] }`

## **6. 동시성 모델**

  - **메인 고루틴**: **Bubble Tea의 이벤트 루프**(`tui.Run()`)를 실행한다. 이 고루틴은 사용자의 키 입력과 같은 UI 이벤트를 처리하고 화면을 갱신하는 역할만 한다. **절대 블로킹(Blocking)되어서는 안 된다.**
  - **네트워크 고루틴**: `internal/network` 모듈이 별도의 고루틴을 생성하여 WebSocket 서버로부터 메시지를 **지속적으로 수신 대기**한다 (`readPump` 역할).
  - **안전한 상태 업데이트**: 네트워크 고루틴이 서버로부터 메시지를 수신하면, TUI의 `Model`을 직접 수정하지 않는다. 대신, 수신한 데이터를 `tea.Msg`로 감싸 Bubble Tea의 이벤트 루프(메인 고루틴)로 **채널을 통해 전달**한다. `Update` 함수가 이 메시지를 받아 안전하게 UI 상태를 갱신한다. (Race Condition 방지)

## **7. 코드 컨벤션**

  - **포맷팅**: \*\*`goimports`\*\*를 사용하여 포맷팅과 import 문을 자동으로 정리한다. (IDE 저장 시 자동 실행 설정 권장)
  - **네이밍**:
      - 패키지명: 짧고 간결한 소문자.
      - 변수/함수명: Go 표준인 \*\*카멜케이스(CamelCase)\*\*를 따른다.
  - **에러 처리**: `if err != nil` 패턴을 철저히 따른다. 에러는 무시하지 않고, 로깅하거나 TUI의 상태바 등을 통해 사용자에게 명확히 알린다.
  - **주석**: 외부로 공개되는 모든 함수와 타입에는 `godoc` 표준에 맞는 주석을 작성한다.

## **8. 빌드 및 배포**

  - **크로스 컴파일**: Go의 강력한 크로스 컴파일 기능을 활용하여 **다양한 OS와 아키텍처**용 바이너리를 생성한다.
      - `GOOS=linux GOARCH=amd64 go build ...`
      - `GOOS=windows GOARCH=amd64 go build ...`
      - `GOOS=darwin GOARCH=amd64 go build ...`
  - **배포**: 컴파일된 단일 실행 바이너리 파일들을 **GitHub Releases**를 통해 압축 파일로 배포하여 사용자들이 자신의 환경에 맞게 다운로드할 수 있도록 한다.
  - **Dockerfile**: (선택 사항) `alpine` 기반의 경량 이미지에서 클라이언트를 실행할 수 있는 Dockerfile을 제공하여 컨테이너 환경 사용자도 지원한다.