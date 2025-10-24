package service

import (
	"errors"
	"shell-talk-server/internal/domain"
)

// UserService provides user-related services.
type UserService struct {
	userRepo IUserRepository
}

// NewUserService creates a new UserService.
func NewUserService(userRepo IUserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

// Register creates a new user account.
func (s *UserService) Register(nickname, password string) (*domain.User, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.GetUserByNickname(nickname)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("nickname is already taken")
	}

	// Create new user domain object (handles password hashing)
	newUser, err := domain.NewUser(nickname, password)
	if err != nil {
		return nil, err
	}

	// Persist user
	if err := s.userRepo.CreateUser(newUser); err != nil {
		return nil, err
	}

	return newUser, nil
}

// Login authenticates a user.
func (s *UserService) Login(nickname, password string) (*domain.User, error) {
	// Find user by nickname
	user, err := s.userRepo.GetUserByNickname(nickname)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("invalid credentials")
	}

	// Check password
	if !user.CheckPassword(password) {
		return nil, errors.New("invalid credentials")
	}

	return user, nil
}

// GetUserByNickname retrieves a user by their nickname.
func (s *UserService) GetUserByNickname(nickname string) (*domain.User, error) {
	return s.userRepo.GetUserByNickname(nickname)
}
