package dtos

import (
	"time"

	"github.com/andrewcopp/Calcutta/backend/pkg/models"
)

type EntryResponse struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	UserID     *string   `json:"userId,omitempty"`
	CalcuttaID string    `json:"calcuttaId"`
	Created    time.Time `json:"created"`
	Updated    time.Time `json:"updated"`
}

func NewEntryResponse(e *models.CalcuttaEntry) *EntryResponse {
	return &EntryResponse{
		ID:         e.ID,
		Name:       e.Name,
		UserID:     e.UserID,
		CalcuttaID: e.CalcuttaID,
		Created:    e.Created,
		Updated:    e.Updated,
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
	ID      string `json:"id"`
	EntryID string `json:"entryId"`
	TeamID  string `json:"teamId"`
	Bid     int    `json:"bid"`
}

func NewEntryTeamResponse(et *models.CalcuttaEntryTeam) *EntryTeamResponse {
	return &EntryTeamResponse{
		ID:      et.ID,
		EntryID: et.EntryID,
		TeamID:  et.TeamID,
		Bid:     et.Bid,
	}
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
