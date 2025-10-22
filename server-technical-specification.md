# **채팅 서버 기술 명세서**

## **1. 개요**

본 문서는 ShellTalk 서버의 아키텍처, 디렉토리 구조, 코드 컨벤션, 통신 프로토콜 등 기술적인 표준을 정의한다. 목표는 **유지보수가 용이하고, 테스트 가능하며, 확장성 있는** 코드 베이스를 구축하는 것이다.

---

## **2. 기술 스택 및 주요 라이브러리**

  - **언어**: **Go** (최신 안정 버전)
  - **웹소켓**: **gorilla/websocket**
      - 이유: Go 커뮤니티에서 사실상의 표준으로 널리 사용되며, 안정성과 유연성이 검증됨.
  - **HTTP 라우터**: **gorilla/mux**
      - 이유: 표준 라이브러리보다 유연한 라우팅(경로 변수, 메소드 지정 등)을 지원하여 API 구현에 용이함.
  - **설정 관리**: **spf13/viper**
      - 이유: 파일(YAML, JSON 등), 환경 변수, 원격 K/V 저장소 등 다양한 소스에서 설정을 손쉽게 관리할 수 있음.
  - **로깅**: **zerolog**
      - 이유: 빠른 성능과 구조화된 로깅(Structured Logging)을 지원하여 로그 분석 및 추적에 매우 유리함.

---

## **3. 아키텍처**

\*\*클린 아키텍처(Clean Architecture)\*\*를 채택하여 각 계층의 역할을 명확히 분리하고 의존성을 관리한다.

  - **핵심 원칙 (의존성 규칙)**: 모든 의존성은 바깥쪽에서 안쪽(`Handler` → `Service` → `Repository`)으로만 향한다. 안쪽 계층은 바깥쪽 계층을 절대 알지 못한다. 이는 \*\*인터페이스(Interface)\*\*를 통해 구현된다.

  - **계층별 역할**:

      - **Domain**: 애플리케이션의 핵심 데이터 구조(Struct). 모든 계층에서 사용된다.
      - **Service (Usecase)**: 순수한 비즈니스 로직. Repository 인터페이스에 의존한다.
      - **Repository**: 데이터 영속성 처리. Service에 정의된 인터페이스를 구현한다. (초기 버전은 **In-Memory**로 구현)
      - **Handler (Delivery)**: HTTP 요청 및 WebSocket 연결을 처리. Service를 호출하여 비즈니스 로직을 실행한다.

---

## **4. 디렉토리 구조 및 패키지 역할**

Standard Go Project Layout을 따르며, 클린 아키텍처를 반영하여 구성한다.

```
shell-talk-server/
├── cmd/
│   └── server/
│       └── main.go         # 🚀 애플리케이션 시작점, 의존성 주입
├── configs/
│   └── config.yaml         # ⚙️ 서버 설정 파일
├── internal/
│   ├── domain/             # 📄 핵심 데이터 모델 (Room, Message, Client)
│   ├── hub/                # ❤️‍🔥 채팅방(Hub)의 동시성 처리 로직
│   ├── handler/            # 🎮 HTTP, WebSocket 요청 처리
│   ├── service/            # 🧠 비즈니스 로직, Repository 인터페이스 정의
│   └── repository/         # 🗄️ 데이터 영속성 로직 (In-memory 구현)
├── pkg/
│   └── utils/              # 🛠️ 프로젝트 전반에서 사용될 수 있는 유틸리티
├── go.mod                  # 📦 의존성 관리
├── go.sum
└── Dockerfile              # 🐳 배포용 도커 파일
```

  - **`cmd/server/main.go`**: 각 계층의 구현체(Repository, Service, Handler)를 생성하고 의존성을 주입(`main` 함수)하여 서버를 실행한다.
  - **`internal/domain`**: `Room`, `Message` 등 순수한 데이터 구조(Struct)만 정의한다.
  - **`internal/hub`**: **채팅방의 심장**. 각 채팅방의 클라이언트 관리, 메시지 브로드캐스팅 등 **동시성** 관련 로직을 채널(Channel)을 통해 안전하게 처리한다.
  - **`internal/handler`**: 외부 세계와의 접점. HTTP 핸들러와 WebSocket 핸들러를 포함한다. JSON 요청을 파싱하여 Service에 전달하고, 결과를 JSON으로 응답한다.
  - **`internal/service`**: `ChatService`와 같은 인터페이스와 그 구현체를 포함한다. 비즈니스 규칙을 실행하며, `Repository` 인터페이스를 통해 데이터에 접근한다.
  - **`internal/repository`**: `service`에 정의된 `Repository` 인터페이스를 구현한다. 초기에는 In-memory (메모리 맵 `map`) 방식으로 채팅방 데이터를 관리한다.

---

## **5. 통신 프로토콜**

### **5.1. HTTP REST API (채팅방 관리)**

  - `Content-Type`: `application/json`

| Method | Endpoint | 설명 |
| :--- | :--- | :--- |
| `POST` | `/api/rooms` | 새 채팅방 생성 |
| `GET` | `/api/rooms` | 전체 채팅방 목록 조회 |

### **5.2. WebSocket 메시지 프로토콜**

  - 클라이언트와 서버는 **JSON 형식**의 메시지를 주고받는다.
  - 모든 메시지는 `type`과 `payload` 필드를 가진다.

<!-- end list -->

```json
{
  "type": "MESSAGE_TYPE_IN_SNAKE_CASE",
  "payload": {
    // 메시지 데이터
  }
}
```

  - **주요 메시지 타입 예시**:
      - **클라이언트 → 서버**:
          - `send_public_message`: `{ "content": "안녕하세요" }`
          - `send_direct_message`: `{ "recipient": "user2", "content": "귓속말입니다" }`
          - `list_users`: `{}`
      - **서버 → 클라이언트**:
          - `new_public_message`: `{ "sender": "user1", "content": "안녕하세요", "timestamp": "..." }`
          - `new_direct_message`: `{ "sender": "user1", "content": "귓속말입니다", "timestamp": "..." }`
          - `system_message`: `{ "content": "'user3'님이 입장하셨습니다.", "timestamp": "..." }`
          - `user_list`: `{ "users": ["user1", "user2", "user3"] }`

---

## **6. 동시성 모델**

  - 각 클라이언트의 WebSocket 연결마다 2개의 \*\*고루틴(Goroutine)\*\*이 생성된다.
      - **`readPump`**: 클라이언트로부터 메시지를 읽어 Hub의 채널로 전달.
      - **`writePump`**: Hub의 채널로부터 메시지를 받아 클라이언트에게 전송.
  - **Hub**는 중앙 처리자로서 동작하며, 채널을 통해 모든 메시지를 순차적으로 처리한다. 이를 통해 `map`과 같은 공유 데이터에 대한 동시 접근을 안전하게 관리하고 **Race Condition을 방지**한다.

---

## **7. 코드 컨벤션**

  - **포맷팅**: \*\*`goimports`\*\*를 사용하여 포맷팅과 import 문을 자동으로 정리한다. (IDE 저장 시 자동 실행 설정 권장)
  - **네이밍**:
      - 패키지명: 짧고 간결한 소문자. (`chat_service` (X) → `service` (O))
      - 변수/함수명: Go 표준인 \*\*카멜케이스(CamelCase)\*\*를 따른다. (`PascalCase`는 외부 공개, `camelCase`는 내부 사용)
  - **에러 처리**: `if err != nil` 패턴을 철저히 따른다. 에러는 무시하지 않고, 로깅하거나 상위 호출자로 반환하여 명시적으로 처리한다. `panic`은 정말 복구 불가능한 상황이 아니면 사용하지 않는다.
  - **주석**: 외부로 공개되는 모든 함수와 타입에는 `godoc` 표준에 맞는 주석을 작성한다.

---

## **8. 빌드 및 배포**

  - **Dockerfile**: \*\*멀티 스테이지 빌드(Multi-stage builds)\*\*를 활용한다.
      - `builder` 스테이지: Go 소스 코드를 컴파일하여 단일 실행 바이너리 생성.
      - 최종 스테이지: `alpine`이나 `scratch` 같은 초경량 이미지에 컴파일된 바이너리 파일과 설정 파일만 복사하여 이미지 사이즈를 최소화한다.
  - **컴파일**: `CGO_ENABLED=0 GOOS=linux go build ...` 와 같이 정적 바이너리(Static Binary)로 컴파일하여 어떤 Linux 환경에서도 의존성 없이 실행되도록 한다.