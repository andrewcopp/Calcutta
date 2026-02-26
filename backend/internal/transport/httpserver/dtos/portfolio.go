package dtos

import (
	"strings"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

type PortfolioResponse struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	UserID         *string   `json:"userId,omitempty"`
	PoolID         string    `json:"poolId"`
	Status         string    `json:"status"`
	TotalReturns   float64   `json:"totalReturns"`
	FinishPosition int       `json:"finishPosition"`
	InTheMoney     bool      `json:"inTheMoney"`
	PayoutCents    int       `json:"payoutCents"`
	IsTied         bool      `json:"isTied"`
	ExpectedValue      *float64  `json:"expectedValue,omitempty"`
	ProjectedFavorites *float64  `json:"projectedFavorites,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

func NewPortfolioResponse(p *models.Portfolio, s *models.PortfolioStanding) *PortfolioResponse {
	resp := &PortfolioResponse{
		ID:        p.ID,
		Name:      p.Name,
		UserID:    p.UserID,
		PoolID:    p.PoolID,
		Status:    p.Status,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
	if s != nil {
		resp.TotalReturns = s.TotalReturns
		resp.FinishPosition = s.FinishPosition
		resp.InTheMoney = s.InTheMoney
		resp.PayoutCents = s.PayoutCents
		resp.IsTied = s.IsTied
		resp.ExpectedValue = s.ExpectedValue
		resp.ProjectedFavorites = s.ProjectedFavorites
	}
	return resp
}

func NewPortfolioListResponse(portfolios []*models.Portfolio, standingsByID map[string]*models.PortfolioStanding) []*PortfolioResponse {
	if portfolios == nil {
		return []*PortfolioResponse{}
	}

	responses := make([]*PortfolioResponse, len(portfolios))
	for i, p := range portfolios {
		responses[i] = NewPortfolioResponse(p, standingsByID[p.ID])
	}
	return responses
}

type CreatePortfolioRequest struct {
	Name   string  `json:"name"`
	UserID *string `json:"userId,omitempty"`
}

func (r *CreatePortfolioRequest) Validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return ErrFieldRequired("name")
	}
	return nil
}

type UpdateInvestmentRequest struct {
	TeamID  string `json:"teamId"`
	Credits int    `json:"credits"`
}

type UpdatePortfolioRequest struct {
	Teams []*UpdateInvestmentRequest `json:"teams"`
}

func (r *UpdatePortfolioRequest) Validate() error {
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
		if t.Credits <= 0 {
			return ErrFieldInvalid("credits", "must be greater than 0")
		}
	}
	return nil
}

type InvestmentResponse struct {
	ID        string                  `json:"id"`
	PortfolioID string               `json:"portfolioId"`
	TeamID    string                  `json:"teamId"`
	Credits   int                     `json:"credits"`
	CreatedAt time.Time               `json:"createdAt"`
	UpdatedAt time.Time               `json:"updatedAt"`
	Team      *TournamentTeamResponse `json:"team,omitempty"`
}

func NewInvestmentResponse(inv *models.Investment) *InvestmentResponse {
	resp := &InvestmentResponse{
		ID:          inv.ID,
		PortfolioID: inv.PortfolioID,
		TeamID:      inv.TeamID,
		Credits:     inv.Credits,
		CreatedAt:   inv.CreatedAt,
		UpdatedAt:   inv.UpdatedAt,
	}
	if inv.Team != nil {
		resp.Team = NewTournamentTeamResponse(inv.Team, inv.Team.School)
	}
	return resp
}

func NewInvestmentListResponse(investments []*models.Investment) []*InvestmentResponse {
	if investments == nil {
		return []*InvestmentResponse{}
	}

	responses := make([]*InvestmentResponse, len(investments))
	for i, inv := range investments {
		responses[i] = NewInvestmentResponse(inv)
	}
	return responses
}
