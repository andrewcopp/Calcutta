package models

import "time"

type CalcuttaInvitation struct {
	ID         string     `json:"id"`
	CalcuttaID string     `json:"calcuttaId"`
	UserID     string     `json:"userId"`
	InvitedBy  string     `json:"invitedBy"`
	Status     string     `json:"status"`
	RevokedAt  *time.Time `json:"revokedAt,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
	DeletedAt  *time.Time `json:"deletedAt,omitempty"`
}
