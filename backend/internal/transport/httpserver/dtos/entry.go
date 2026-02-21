package dtos

import (
	"strings"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

type EntryResponse struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	UserID         *string   `json:"userId,omitempty"`
	CalcuttaID     string    `json:"calcuttaId"`
	Status         string    `json:"status"`
	TotalPoints    float64   `json:"totalPoints"`
	FinishPosition int       `json:"finishPosition"`
	InTheMoney     bool      `json:"inTheMoney"`
	PayoutCents    int       `json:"payoutCents"`
	IsTied         bool      `json:"isTied"`
	Created        time.Time `json:"created"`
	Updated        time.Time `json:"updated"`
}

func NewEntryResponse(e *models.CalcuttaEntry) *EntryResponse {
	return &EntryResponse{
		ID:             e.ID,
		Name:           e.Name,
		UserID:         e.UserID,
		CalcuttaID:     e.CalcuttaID,
		Status:         e.Status,
		TotalPoints:    e.TotalPoints,
		FinishPosition: e.FinishPosition,
		InTheMoney:     e.InTheMoney,
		PayoutCents:    e.PayoutCents,
		IsTied:         e.IsTied,
		Created:        e.Created,
		Updated:        e.Updated,
	}
}

func NewEntryListResponse(entries []*models.CalcuttaEntry) []*EntryResponse {
	if entries == nil {
		return []*EntryResponse{}
	}

	responses := make([]*EntryResponse, len(entries))
	for i, e := range entries {
		responses[i] = NewEntryResponse(e)
	}
	return responses
}

type CreateEntryRequest struct {
	Name   string  `json:"name"`
	UserID *string `json:"userId,omitempty"`
}

func (r *CreateEntryRequest) Validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return ErrFieldRequired("name")
	}
	return nil
}

type UpdateEntryTeamRequest struct {
	TeamID string `json:"teamId"`
	Bid    int    `json:"bid"`
}

type UpdateEntryRequest struct {
	Teams []*UpdateEntryTeamRequest `json:"teams"`
}

func (r *UpdateEntryRequest) Validate() error {
	if r == nil {
		return ErrFieldInvalid("body", "request body is required")
	}
	if len(r.Teams) == 0 {
		return ErrFieldInvalid("teams", "at least one team must be provided")
	}
	for _, t := range r.Teams {
		if t == nil {
			return ErrFieldInvalid("teams", "team cannot be null")
		}
		if strings.TrimSpace(t.TeamID) == "" {
			return ErrFieldRequired("teamId")
		}
		if t.Bid <= 0 {
			return ErrFieldInvalid("bid", "must be greater than 0")
		}
	}
	return nil
}

type EntryTeamResponse struct {
	ID      string                  `json:"id"`
	EntryID string                  `json:"entryId"`
	TeamID  string                  `json:"teamId"`
	BidPoints int                     `json:"bidPoints"`
	Team      *TournamentTeamResponse `json:"team,omitempty"`
}

func NewEntryTeamResponse(et *models.CalcuttaEntryTeam) *EntryTeamResponse {
	resp := &EntryTeamResponse{
		ID:      et.ID,
		EntryID: et.EntryID,
		TeamID:  et.TeamID,
		BidPoints: et.BidPoints,
	}
	if et.Team != nil {
		resp.Team = NewTournamentTeamResponse(et.Team, et.Team.School)
	}
	return resp
}

func NewEntryTeamListResponse(teams []*models.CalcuttaEntryTeam) []*EntryTeamResponse {
	if teams == nil {
		return []*EntryTeamResponse{}
	}

	responses := make([]*EntryTeamResponse, len(teams))
	for i, t := range teams {
		responses[i] = NewEntryTeamResponse(t)
	}
	return responses
}
