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

type LabelGrant struct {
	Key       string  `json:"key"`
	ScopeType string  `json:"scopeType"`
	ScopeID   *string `json:"scopeId,omitempty"`
	ScopeName *string `json:"scopeName,omitempty"`
}

type AdminUserDetailResponse struct {
	ID          string       `json:"id"`
	Email       *string      `json:"email,omitempty"`
	FirstName   string       `json:"firstName"`
	LastName    string       `json:"lastName"`
	Status      string       `json:"status"`
	Labels      []LabelGrant `json:"labels"`
	Permissions []string     `json:"permissions"`
	CreatedAt   time.Time    `json:"createdAt"`
	UpdatedAt   time.Time    `json:"updatedAt"`
}
