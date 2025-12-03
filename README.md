# Shell-Talk

Go로 작성된 간단한 클라이언트-서버 기반 채팅 애플리케이션입니다. 이 프로젝트는 실시간 메시지 교환을 위한 서버와 상호 작용하는 터미널 기반 클라이언트를 제공합니다.

## ✨ 주요 기능

-   실시간 채팅 기능
-   채팅방 생성 및 참여
-   데이터베이스 지원 (PostgreSQL, MongoDB)
-   Docker Compose를 이용한 간편한 개발 환경 구성

## 📂 디렉토리 구조

```
/
├── shell-talk-client/  # Go로 작성된 터미널 클라이언트
├── shell-talk-server/  # Go로 작성된 채팅 서버
├── docker-compose.yml  # 개발 환경을 위한 Docker Compose 설정
└── ...
```

## ⚙️ 사전 준비사항

-   Go (v1.21 이상 권장)
-   Docker 및 Docker Compose

## 🚀 시작하기

### 1. 프로젝트 클론

```bash
git clone <repository-url>
cd shell-talk
```

### 2. 데이터베이스 실행

프로젝트 루트 디렉토리에서 아래 명령어를 실행하여 Docker Compose로 데이터베이스(PostgreSQL, MongoDB)를 백그라운드에서 실행합니다.

```bash
docker-compose up -d postgres mongo
```

`docker-compose ps` 명령어로 서비스가 정상적으로 실행 중인지 확인할 수 있습니다.

### 3. 의존성 코드 생성 (Wire)

서버는 최초 실행 전 `wire`를 사용하여 의존성 주입 관련 코드를 생성해야 합니다. 이 과정이 없으면 `undefined: InitializeApp` 오류가 발생합니다.

먼저 `wire` 도구를 설치합니다.

```bash
go install github.com/google/wire/cmd/wire@latest
```

그 다음, 서버의 `cmd/server` 디렉토리로 이동하여 `wire` 명령어를 실행해 코드를 생성합니다.

```bash
cd shell-talk-server/cmd/server
wire
cd ../../..
```

### 4. 서버 실행

새로운 터미널을 열고, `shell-talk-server` 디렉토리로 이동하여 서버를 실행합니다. 서버는 실행 시 데이터베이스 연결을 위한 환경 변수가 필요합니다.

```bash
cd shell-talk-server

# PostgreSQL 사용 시
export POSTGRES_URL=postgres://user:password@localhost:5432/shelltalk?sslmode=disable
export MONGO_URL=

# MongoDB 사용 시
export MONGO_URL=mongodb://user:password@localhost:27017
export POSTGRES_URL=

# 서버 실행
go run ./cmd/server
```

> **참고**: `docker-compose.yml` 파일에 서버 서비스가 주석 처리되어 있습니다. 주석을 해제하고 빌드 설정을 완료하면 `docker-compose up --build` 명령어로 서버까지 한 번에 실행할 수 있습니다.

### 5. 클라이언트 실행

또 다른 새 터미널을 열고, `shell-talk-client` 디렉토리로 이동하여 클라이언트를 실행합니다.

```bash
cd shell-talk-client
go run ./cmd/client/main.go
```

클라이언트가 실행되면 서버에 연결하여 채팅을 시작할 수 있습니다.

## 🌿 브랜치 전략

이 프로젝트는 `simple-gitflow-branch-strategy.md` 파일에 기술된 GitFlow 기반의 간단한 브랜치 전략을 따릅니다.
