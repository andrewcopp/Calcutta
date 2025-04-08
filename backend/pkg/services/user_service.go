package services

import (
	"context"
	"fmt"
	"time"

	"github.com/andrewcopp/Calcutta/backend/pkg/models"
	"github.com/google/uuid"
)

type UserService struct {
	userRepo *UserRepository
}

func NewUserService(userRepo *UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

func (s *UserService) Login(ctx context.Context, email string) (*models.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

func (s *UserService) Signup(ctx context.Context, email, firstName, lastName string) (*models.User, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	if existingUser != nil {
		return nil, fmt.Errorf("user already exists")
	}

	// Create new user
	user := &models.User{
		ID:        uuid.New().String(),
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		Created:   time.Now(),
		Updated:   time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}
