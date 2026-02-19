package models

import "time"

// User represents a participant in the Calcutta system
type User struct {
	ID                 string     `json:"id"`
	Email              *string    `json:"email,omitempty"`
	FirstName          string     `json:"firstName"`
	LastName           string     `json:"lastName"`
	Status             string     `json:"status"`
	PasswordHash       *string    `json:"-"`
	ExternalProvider   *string    `json:"externalProvider,omitempty"`
	ExternalProviderID *string    `json:"externalProviderId,omitempty"`
	Created            time.Time  `json:"created"`
	Updated            time.Time  `json:"updated"`
	Deleted            *time.Time `json:"deleted,omitempty"`
}
