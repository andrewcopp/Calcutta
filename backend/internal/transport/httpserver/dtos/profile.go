package dtos

import "time"

type UserProfileResponse struct {
	ID          string    `json:"id"`
	Email       *string   `json:"email,omitempty"`
	FirstName   string    `json:"firstName"`
	LastName    string    `json:"lastName"`
	Status      string    `json:"status"`
	Roles       []string  `json:"roles"`
	Permissions []string  `json:"permissions"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type RoleGrant struct {
	Key       string  `json:"key"`
	ScopeType string  `json:"scopeType"`
	ScopeID   *string `json:"scopeId,omitempty"`
	ScopeName *string `json:"scopeName,omitempty"`
}

type AdminUserDetailResponse struct {
	ID          string      `json:"id"`
	Email       *string     `json:"email,omitempty"`
	FirstName   string      `json:"firstName"`
	LastName    string      `json:"lastName"`
	Status      string      `json:"status"`
	Roles       []RoleGrant `json:"roles"`
	Permissions []string    `json:"permissions"`
	CreatedAt   time.Time   `json:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt"`
}
