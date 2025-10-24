package service

import (
	"context"
	"shell-talk-server/internal/domain"

	"github.com/google/uuid"
)

// --- Service Interfaces ---

// IUserService defines the interface for user-related business logic.
type IUserService interface {
	Register(nickname, password string) (*domain.User, error)
	Login(nickname, password string) (*domain.User, error)
	GetUserByNickname(nickname string) (*domain.User, error)
}

// IRoomService defines the interface for room-related business logic.
type IRoomService interface {
	CreateRoom(name, password string, owner *domain.User) (*domain.Room, error)
	JoinRoom(name, password string, user *domain.User) (*domain.Room, error)
	LeaveRoom(name string, user *domain.User) (*domain.Room, error)
	ListRooms() ([]*domain.Room, error)
	GetRoomByName(name string) (*domain.Room, error)
	GetRoomMembers(name string) ([]string, error)
	IsRoomMember(name string, user *domain.User) (bool, error)
	GetRoomMemberIDs(name string) ([]uuid.UUID, error)
}

// --- Repository Interfaces ---

// IUserRepository defines the interface for user persistence.
type IUserRepository interface {
	CreateUser(user *domain.User) error
	GetUserByNickname(nickname string) (*domain.User, error)
	GetUserByID(id uuid.UUID) (*domain.User, error)
}

// IRoomRepository defines the interface for room persistence.
type IRoomRepository interface {
	CreateRoom(room *domain.Room) error
	GetRoomByName(name string) (*domain.Room, error)
	AddUserToRoom(roomID, userID uuid.UUID) error
	GetRoomMembers(roomID uuid.UUID) ([]string, error)
	ListRooms() ([]*domain.Room, error)
	RemoveUserFromRoom(roomID, userID uuid.UUID) error
	IsRoomMember(roomID, userID uuid.UUID) (bool, error)
	GetRoomMemberIDs(roomID uuid.UUID) ([]uuid.UUID, error)
}

// IMessageRepository defines the interface for message persistence.
type IMessageRepository interface {
	SaveMessage(ctx context.Context, message *domain.ChatMessage) error
	GetMessagesByConversationID(ctx context.Context, conversationID string, limit int64) ([]*domain.ChatMessage, error)
}
