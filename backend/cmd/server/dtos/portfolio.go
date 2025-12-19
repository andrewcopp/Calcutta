package dtos

import (
	"time"

	"github.com/andrewcopp/Calcutta/backend/pkg/models"
)

type PortfolioResponse struct {
	ID            string    `json:"id"`
	EntryID       string    `json:"entryId"`
	MaximumPoints float64   `json:"maximumPoints"`
	Created       time.Time `json:"created"`
	Updated       time.Time `json:"updated"`
}

func NewPortfolioResponse(p *models.CalcuttaPortfolio) *PortfolioResponse {
	return &PortfolioResponse{
		ID:            p.ID,
		EntryID:       p.EntryID,
		MaximumPoints: p.MaximumPoints,
		Created:       p.Created,
		Updated:       p.Updated,
	}
}

func NewPortfolioListResponse(portfolios []*models.CalcuttaPortfolio) []*PortfolioResponse {
	if portfolios == nil {
		return []*PortfolioResponse{}
	}

	responses := make([]*PortfolioResponse, len(portfolios))
	for i, p := range portfolios {
		responses[i] = NewPortfolioResponse(p)
	}
	return responses
}

type PortfolioTeamResponse struct {
	ID                  string    `json:"id"`
	PortfolioID         string    `json:"portfolioId"`
	TeamID              string    `json:"teamId"`
	OwnershipPercentage float64   `json:"ownershipPercentage"`
	ExpectedPoints      float64   `json:"expectedPoints"`
	PredictedPoints     float64   `json:"predictedPoints"`
	ActualPoints        float64   `json:"actualPoints"`
	Updated             time.Time `json:"updated"`
}

func NewPortfolioTeamResponse(pt *models.CalcuttaPortfolioTeam) *PortfolioTeamResponse {
	return &PortfolioTeamResponse{
		ID:                  pt.ID,
		PortfolioID:         pt.PortfolioID,
		TeamID:              pt.TeamID,
		OwnershipPercentage: pt.OwnershipPercentage,
		ExpectedPoints:      pt.ExpectedPoints,
		PredictedPoints:     pt.PredictedPoints,
		ActualPoints:        pt.ActualPoints,
		Updated:             pt.Updated,
	}
}

func NewPortfolioTeamListResponse(teams []*models.CalcuttaPortfolioTeam) []*PortfolioTeamResponse {
	if teams == nil {
		return []*PortfolioTeamResponse{}
	}

	responses := make([]*PortfolioTeamResponse, len(teams))
	for i, t := range teams {
		responses[i] = NewPortfolioTeamResponse(t)
	}
	return responses
}

type UpdatePortfolioTeamScoresRequest struct {
	ExpectedPoints  float64 `json:"expectedPoints"`
	PredictedPoints float64 `json:"predictedPoints"`
}

func (r *UpdatePortfolioTeamScoresRequest) Validate() error {
	if r.ExpectedPoints < 0 {
		return ErrFieldInvalid("expectedPoints", "must be >= 0")
	}
	if r.PredictedPoints < 0 {
		return ErrFieldInvalid("predictedPoints", "must be >= 0")
	}
	return nil
}

type UpdatePortfolioMaximumScoreRequest struct {
	MaximumPoints float64 `json:"maximumPoints"`
}

func (r *UpdatePortfolioMaximumScoreRequest) Validate() error {
	if r.MaximumPoints < 0 {
		return ErrFieldInvalid("maximumPoints", "must be >= 0")
	}
	return nil
}
