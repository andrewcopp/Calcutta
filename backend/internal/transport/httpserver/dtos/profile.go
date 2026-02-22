package dtos

import "time"

type UserProfileResponse struct {
	ID          string    `json:"id"`
	Email       *string   `json:"email,omitempty"`
	FirstName   string    `json:"firstName"`
	LastName    string    `json:"lastName"`
	Status      string    `json:"status"`
	Labels      []string  `json:"labels"`
	Permissions []string  `json:"permissions"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
