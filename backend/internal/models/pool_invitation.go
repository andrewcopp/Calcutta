package models

import "time"

type PoolInvitation struct {
	ID        string     `json:"id"`
	PoolID    string     `json:"poolId"`
	UserID    string     `json:"userId"`
	InvitedBy string     `json:"invitedBy"`
	Status    string     `json:"status"`
	RevokedAt *time.Time `json:"revokedAt,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt,omitempty"`
}
