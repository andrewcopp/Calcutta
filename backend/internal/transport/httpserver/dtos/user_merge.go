package dtos

import (
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

type StubUserResponse struct {
	ID        string  `json:"id"`
	FirstName string  `json:"firstName"`
	LastName  string  `json:"lastName"`
	Email     *string `json:"email,omitempty"`
	Status    string  `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

func NewStubUserResponse(u *models.User) StubUserResponse {
	return StubUserResponse{
		ID:        u.ID,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Email:     u.Email,
		Status:    u.Status,
		CreatedAt: u.CreatedAt,
	}
}

type StubUsersListResponse struct {
	Items []StubUserResponse `json:"items"`
}

type MergeCandidateResponse struct {
	ID        string  `json:"id"`
	FirstName string  `json:"firstName"`
	LastName  string  `json:"lastName"`
	Email     *string `json:"email,omitempty"`
	Status    string  `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

func NewMergeCandidateResponse(u *models.User) MergeCandidateResponse {
	return MergeCandidateResponse{
		ID:        u.ID,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Email:     u.Email,
		Status:    u.Status,
		CreatedAt: u.CreatedAt,
	}
}

type MergeCandidatesListResponse struct {
	Items []MergeCandidateResponse `json:"items"`
}

type MergeUsersRequest struct {
	SourceUserID string `json:"sourceUserId"`
	TargetUserID string `json:"targetUserId"`
}

type UserMergeResponse struct {
	ID               string    `json:"id"`
	SourceUserID     string    `json:"sourceUserId"`
	TargetUserID     string    `json:"targetUserId"`
	MergedBy         string    `json:"mergedBy"`
	EntriesMoved     int       `json:"entriesMoved"`
	InvitationsMoved int       `json:"invitationsMoved"`
	GrantsMoved      int       `json:"grantsMoved"`
	CreatedAt        time.Time `json:"createdAt"`
}

func NewUserMergeResponse(m *models.UserMerge) UserMergeResponse {
	return UserMergeResponse{
		ID:               m.ID,
		SourceUserID:     m.SourceUserID,
		TargetUserID:     m.TargetUserID,
		MergedBy:         m.MergedBy,
		EntriesMoved:     m.EntriesMoved,
		InvitationsMoved: m.InvitationsMoved,
		GrantsMoved:      m.GrantsMoved,
		CreatedAt:        m.CreatedAt,
	}
}

type MergeHistoryResponse struct {
	Items []UserMergeResponse `json:"items"`
}
