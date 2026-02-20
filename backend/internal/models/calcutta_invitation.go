package models

import "time"

type CalcuttaInvitation struct {
	ID         string     `json:"id"`
	CalcuttaID string     `json:"calcuttaId"`
	UserID     string     `json:"userId"`
	InvitedBy  string     `json:"invitedBy"`
	Status     string     `json:"status"`
	RevokedAt  *time.Time `json:"revokedAt,omitempty"`
	Created    time.Time  `json:"created"`
	Updated    time.Time  `json:"updated"`
	Deleted    *time.Time `json:"deleted,omitempty"`
}
