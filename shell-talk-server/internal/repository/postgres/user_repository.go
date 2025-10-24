package postgres

import (
	"database/sql"
	"shell-talk-server/internal/domain"

	"github.com/google/uuid"
)

// UserRepository handles database operations for users.
type UserRepository struct {
	DB *sql.DB
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{DB: db}
}

// CreateUser inserts a new user into the database.
func (r *UserRepository) CreateUser(user *domain.User) error {
	query := `INSERT INTO users (id, nickname, password_hash, created_at) VALUES ($1, $2, $3, $4)`
	_, err := r.DB.Exec(query, user.ID, user.Nickname, user.PasswordHash, user.CreatedAt)
	return err
}

// GetUserByNickname retrieves a user by their nickname.
func (r *UserRepository) GetUserByNickname(nickname string) (*domain.User, error) {
	user := &domain.User{}
	query := `SELECT id, nickname, password_hash, created_at FROM users WHERE nickname = $1`
	err := r.DB.QueryRow(query, nickname).Scan(&user.ID, &user.Nickname, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No user found is not an application error
		}
		return nil, err
	}
	return user, nil
}

// GetUserByID retrieves a user by their ID.
func (r *UserRepository) GetUserByID(id uuid.UUID) (*domain.User, error) {
	user := &domain.User{}
	query := `SELECT id, nickname, password_hash, created_at FROM users WHERE id = $1`
	err := r.DB.QueryRow(query, id).Scan(&user.ID, &user.Nickname, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}
