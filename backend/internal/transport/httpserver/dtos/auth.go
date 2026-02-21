package dtos

import (
	"strings"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r *LoginRequest) Validate() error {
	if strings.TrimSpace(r.Email) == "" {
		return ErrFieldRequired("email")
	}
	if strings.TrimSpace(r.Password) == "" {
		return ErrFieldRequired("password")
	}
	return nil
}

type SignupRequest struct {
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Password  string `json:"password"`
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
	if err := ValidatePassword(r.Password); err != nil {
		return err
	}
	return nil
}

type AuthResponse struct {
	User        *UserResponse `json:"user"`
	AccessToken string        `json:"accessToken"`
}

type UserResponse struct {
	ID        string    `json:"id"`
	Email     *string   `json:"email,omitempty"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func NewUserResponse(u *models.User) *UserResponse {
	return &UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Status:    u.Status,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
