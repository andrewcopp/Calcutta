package dtos

import (
	"time"

	"github.com/andrewcopp/Calcutta/backend/pkg/models"
)

type EntryResponse struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	UserID         *string   `json:"userId,omitempty"`
	CalcuttaID     string    `json:"calcuttaId"`
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

type EntryTeamResponse struct {
	ID      string                  `json:"id"`
	EntryID string                  `json:"entryId"`
	TeamID  string                  `json:"teamId"`
	Bid     int                     `json:"bid"`
	Team    *TournamentTeamResponse `json:"team,omitempty"`
}

func NewEntryTeamResponse(et *models.CalcuttaEntryTeam) *EntryTeamResponse {
	resp := &EntryTeamResponse{
		ID:      et.ID,
		EntryID: et.EntryID,
		TeamID:  et.TeamID,
		Bid:     et.Bid,
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
