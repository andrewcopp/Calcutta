package dtos

import (
	"strings"
	"time"

	"github.com/andrewcopp/Calcutta/backend/pkg/models"
)

type LoginRequest struct {
	Email string `json:"email"`
}

func (r *LoginRequest) Validate() error {
	if strings.TrimSpace(r.Email) == "" {
		return ErrFieldRequired("email")
	}
	return nil
}

type SignupRequest struct {
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

func (r *SignupRequest) Validate() error {
	if strings.TrimSpace(r.Email) == "" {
		return ErrFieldRequired("email")
	}
	if strings.TrimSpace(r.FirstName) == "" {
		return ErrFieldRequired("firstName")
	}
	if strings.TrimSpace(r.LastName) == "" {
		return ErrFieldRequired("lastName")
	}
	return nil
}

type UserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Created   time.Time `json:"created"`
	Updated   time.Time `json:"updated"`
}

func NewUserResponse(u *models.User) *UserResponse {
	return &UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Created:   u.Created,
		Updated:   u.Updated,
	}
}
