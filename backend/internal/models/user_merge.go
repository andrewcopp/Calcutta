package models

import "time"

type UserMerge struct {
	ID               string    `json:"id"`
	SourceUserID     string    `json:"sourceUserId"`
	TargetUserID     string    `json:"targetUserId"`
	MergedBy         string    `json:"mergedBy"`
	EntriesMoved     int       `json:"entriesMoved"`
	InvitationsMoved int       `json:"invitationsMoved"`
	GrantsMoved      int       `json:"grantsMoved"`
	CreatedAt        time.Time `json:"createdAt"`
}
