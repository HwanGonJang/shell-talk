package domain

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// User represents a user account in the system.
type User struct {
	ID           uuid.UUID `json:"id"`
	Nickname     string    `json:"nickname"`
	PasswordHash string    `json:"-"` // Do not expose password hash
	CreatedAt    time.Time `json:"created_at"`
}

// NewUser creates a new user with a hashed password.
func NewUser(nickname, password string) (*User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return &User{
		ID:           uuid.New(),
		Nickname:     nickname,
		PasswordHash: string(hashedPassword),
		CreatedAt:    time.Now(),
	}, nil
}

// CheckPassword compares a plaintext password with the user's hashed password.
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}
