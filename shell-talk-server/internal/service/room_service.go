package service

import (
	"errors"
	"shell-talk-server/internal/domain"

	"github.com/google/uuid"
)

// RoomService provides room-related services.
type RoomService struct {
	roomRepo IRoomRepository
}

// NewRoomService creates a new RoomService.
func NewRoomService(roomRepo IRoomRepository) *RoomService {
	return &RoomService{roomRepo: roomRepo}
}

// CreateRoom creates a new chat room.
func (s *RoomService) CreateRoom(name, password string, owner *domain.User) (*domain.Room, error) {
	// Check if room name is already taken
	existing, err := s.roomRepo.GetRoomByName(name)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("room name is already taken")
	}

	// Create new room domain object
	newRoom, err := domain.NewRoom(name, password, owner.ID)
	if err != nil {
		return nil, err
	}

	// Persist room
	if err := s.roomRepo.CreateRoom(newRoom); err != nil {
		return nil, err
	}

	// Add owner as the first member
	if err := s.roomRepo.AddUserToRoom(newRoom.ID, owner.ID); err != nil {
		return nil, err
	}

	return newRoom, nil
}

// GetUserRooms retrieves all rooms for a given user.
func (s *RoomService) GetUserRooms(userID uuid.UUID) ([]*domain.Room, error) {
	return s.roomRepo.GetUserRooms(userID)
}

// GetRoomByName retrieves a room by its unique name.
func (s *RoomService) GetRoomByName(name string) (*domain.Room, error) {
	return s.roomRepo.GetRoomByName(name)
}

// JoinRoom allows a user to join an existing room.
func (s *RoomService) JoinRoom(name, password string, user *domain.User) (*domain.Room, error) {
	room, err := s.roomRepo.GetRoomByName(name)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return nil, errors.New("room not found")
	}

	if !room.CheckPassword(password) {
		return nil, errors.New("invalid password")
	}

	if err := s.roomRepo.AddUserToRoom(room.ID, user.ID); err != nil {
		return nil, err
	}

	return room, nil
}

// LeaveRoom allows a user to leave a room.
func (s *RoomService) LeaveRoom(name string, user *domain.User) (*domain.Room, error) {
	room, err := s.roomRepo.GetRoomByName(name)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return nil, errors.New("room not found")
	}

	if err := s.roomRepo.RemoveUserFromRoom(room.ID, user.ID); err != nil {
		return nil, err
	}

	return room, nil
}

// ListRooms returns all rooms.
func (s *RoomService) ListRooms() ([]*domain.Room, error) {
	return s.roomRepo.ListRooms()
}

// GetRoomMembers fetches the list of members for a given room name.
func (s *RoomService) GetRoomMembers(name string) ([]string, error) {
	room, err := s.roomRepo.GetRoomByName(name)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return nil, errors.New("room not found")
	}

	return s.roomRepo.GetRoomMembers(room.ID)
}

// GetRoomMemberIDs fetches the list of member UUIDs for a given room name.
func (s *RoomService) GetRoomMemberIDs(name string) ([]uuid.UUID, error) {
	room, err := s.roomRepo.GetRoomByName(name)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return nil, errors.New("room not found")
	}

	return s.roomRepo.GetRoomMemberIDs(room.ID)
}

// IsRoomMember checks if a user is part of a room.
func (s *RoomService) IsRoomMember(name string, user *domain.User) (bool, error) {
	room, err := s.roomRepo.GetRoomByName(name)
	if err != nil {
		return false, err
	}
	if room == nil {
		return false, errors.New("room not found")
	}

	return s.roomRepo.IsRoomMember(room.ID, user.ID)
}
