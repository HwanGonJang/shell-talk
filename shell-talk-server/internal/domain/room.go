package domain

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Room represents a chat room in the system.
type Room struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	OwnerID      uuid.UUID `json:"owner_id"`
	PasswordHash string    `json:"-"` // Do not expose password hash
	CreatedAt    time.Time `json:"created_at"`
}

// NewRoom creates a new room.
func NewRoom(name, password string, ownerID uuid.UUID) (*Room, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return &Room{
		ID:           uuid.New(),
		Name:         name,
		OwnerID:      ownerID,
		PasswordHash: string(hashedPassword),
		CreatedAt:    time.Now(),
	}, nil
}

// CheckPassword compares a plaintext password with the room's hashed password.
func (r *Room) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(r.PasswordHash), []byte(password))
	return err == nil
}
