package postgres

import (
	"database/sql"
	"shell-talk-server/internal/domain"

	"github.com/google/uuid"
)

// RoomRepository handles database operations for rooms.
type RoomRepository struct {
	DB *sql.DB
}

// NewRoomRepository creates a new RoomRepository.
func NewRoomRepository(db *sql.DB) *RoomRepository {
	return &RoomRepository{DB: db}
}

// CreateRoom inserts a new room into the database.
func (r *RoomRepository) CreateRoom(room *domain.Room) error {
	query := `INSERT INTO rooms (id, name, owner_id, password_hash, created_at) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.DB.Exec(query, room.ID, room.Name, room.OwnerID, room.PasswordHash, room.CreatedAt)
	return err
}

// ListRooms retrieves all rooms from the database.
func (r *RoomRepository) ListRooms() ([]*domain.Room, error) {
	query := `SELECT id, name, owner_id, password_hash, created_at FROM rooms`
	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []*domain.Room
	for rows.Next() {
		room := &domain.Room{}
		if err := rows.Scan(&room.ID, &room.Name, &room.OwnerID, &room.PasswordHash, &room.CreatedAt); err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}
	return rooms, nil
}

// GetRoomByName retrieves a room by its unique name.
func (r *RoomRepository) GetRoomByName(name string) (*domain.Room, error) {
	room := &domain.Room{}
	query := `SELECT id, name, owner_id, password_hash, created_at FROM rooms WHERE name = $1`
	err := r.DB.QueryRow(query, name).Scan(&room.ID, &room.Name, &room.OwnerID, &room.PasswordHash, &room.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, err
	}
	return room, nil
}

// AddUserToRoom adds a user to a room's membership list.
func (r *RoomRepository) AddUserToRoom(roomID, userID uuid.UUID) error {
	query := `INSERT INTO room_members (room_id, user_id) VALUES ($1, $2) ON CONFLICT (room_id, user_id) DO NOTHING`
	_, err := r.DB.Exec(query, roomID, userID)
	return err
}

// RemoveUserFromRoom removes a user from a room's membership list.
func (r *RoomRepository) RemoveUserFromRoom(roomID, userID uuid.UUID) error {
	query := `DELETE FROM room_members WHERE room_id = $1 AND user_id = $2`
	_, err := r.DB.Exec(query, roomID, userID)
	return err
}

// IsRoomMember checks if a user is a member of a specific room.
func (r *RoomRepository) IsRoomMember(roomID, userID uuid.UUID) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM room_members WHERE room_id = $1 AND user_id = $2)`
	err := r.DB.QueryRow(query, roomID, userID).Scan(&exists)
	return exists, err
}

// GetRoomMembers retrieves a list of nicknames for users in a room.
func (r *RoomRepository) GetRoomMembers(roomID uuid.UUID) ([]string, error) {
	query := `
		SELECT u.nickname 
		FROM users u 
		JOIN room_members rm ON u.id = rm.user_id 
		WHERE rm.room_id = $1
	`
	rows, err := r.DB.Query(query, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []string
	for rows.Next() {
		var nickname string
		if err := rows.Scan(&nickname); err != nil {
			return nil, err
		}
		members = append(members, nickname)
	}
	return members, nil
}

// GetRoomMemberIDs retrieves a list of user IDs for a room.
func (r *RoomRepository) GetRoomMemberIDs(roomID uuid.UUID) ([]uuid.UUID, error) {
	query := `SELECT user_id FROM room_members WHERE room_id = $1`
	rows, err := r.DB.Query(query, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memberIDs []uuid.UUID
	for rows.Next() {
		var memberID uuid.UUID
		if err := rows.Scan(&memberID); err != nil {
			return nil, err
		}
		memberIDs = append(memberIDs, memberID)
	}
	return memberIDs, nil
}
