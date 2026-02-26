package dtos

import (
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

type OwnershipSummaryResponse struct {
	ID             string    `json:"id"`
	PortfolioID    string    `json:"portfolioId"`
	MaximumReturns float64   `json:"maximumReturns"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

func NewOwnershipSummaryResponse(s *models.OwnershipSummary) *OwnershipSummaryResponse {
	return &OwnershipSummaryResponse{
		ID:             s.ID,
		PortfolioID:    s.PortfolioID,
		MaximumReturns: s.MaximumReturns,
		CreatedAt:      s.CreatedAt,
		UpdatedAt:      s.UpdatedAt,
	}
}

func NewOwnershipSummaryListResponse(summaries []*models.OwnershipSummary) []*OwnershipSummaryResponse {
	if summaries == nil {
		return []*OwnershipSummaryResponse{}
	}

	responses := make([]*OwnershipSummaryResponse, len(summaries))
	for i, s := range summaries {
		responses[i] = NewOwnershipSummaryResponse(s)
	}
	return responses
}

type OwnershipDetailResponse struct {
	ID                  string                  `json:"id"`
	PortfolioID         string                  `json:"portfolioId"`
	TeamID              string                  `json:"teamId"`
	OwnershipPercentage float64                 `json:"ownershipPercentage"`
	ExpectedReturns     float64                 `json:"expectedReturns"`
	ActualReturns       float64                 `json:"actualReturns"`
	CreatedAt           time.Time               `json:"createdAt"`
	UpdatedAt           time.Time               `json:"updatedAt"`
	Team                *TournamentTeamResponse `json:"team,omitempty"`
}

func NewOwnershipDetailResponse(d *models.OwnershipDetail) *OwnershipDetailResponse {
	resp := &OwnershipDetailResponse{
		ID:                  d.ID,
		PortfolioID:         d.PortfolioID,
		TeamID:              d.TeamID,
		OwnershipPercentage: d.OwnershipPercentage,
		ExpectedReturns:     d.ExpectedReturns,
		ActualReturns:       d.ActualReturns,
		CreatedAt:           d.CreatedAt,
		UpdatedAt:           d.UpdatedAt,
	}
	if d.Team != nil {
		resp.Team = NewTournamentTeamResponse(d.Team, d.Team.School)
	}
	return resp
}

func NewOwnershipDetailListResponse(details []*models.OwnershipDetail) []*OwnershipDetailResponse {
	if details == nil {
		return []*OwnershipDetailResponse{}
	}

	responses := make([]*OwnershipDetailResponse, len(details))
	for i, d := range details {
		responses[i] = NewOwnershipDetailResponse(d)
	}
	return responses
}
