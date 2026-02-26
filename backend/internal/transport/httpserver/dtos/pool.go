package dtos

import (
	"fmt"
	"strings"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

type ScoringRuleInput struct {
	WinIndex      int `json:"winIndex"`
	PointsAwarded int `json:"pointsAwarded"`
}

type CreatePoolRequest struct {
	Name                 string             `json:"name"`
	TournamentID         string             `json:"tournamentId"`
	MinTeams             int                `json:"minTeams"`
	MaxTeams             int                `json:"maxTeams"`
	MaxInvestmentCredits int                `json:"maxInvestmentCredits"`
	ScoringRules         []ScoringRuleInput `json:"scoringRules"`
}

func (r *CreatePoolRequest) Validate() error {
	if r.Name == "" {
		return ErrFieldRequired("name")
	}
	if r.TournamentID == "" {
		return ErrFieldRequired("tournamentId")
	}
	if len(r.ScoringRules) == 0 {
		return ErrFieldRequired("scoringRules")
	}
	seen := make(map[int]bool, len(r.ScoringRules))
	for _, rule := range r.ScoringRules {
		if rule.WinIndex < 1 {
			return ErrFieldInvalid("scoringRules", fmt.Sprintf("winIndex must be >= 1, got %d", rule.WinIndex))
		}
		if rule.PointsAwarded < 0 {
			return ErrFieldInvalid("scoringRules", fmt.Sprintf("pointsAwarded must be >= 0, got %d", rule.PointsAwarded))
		}
		if seen[rule.WinIndex] {
			return ErrFieldInvalid("scoringRules", fmt.Sprintf("duplicate winIndex %d", rule.WinIndex))
		}
		seen[rule.WinIndex] = true
	}
	return nil
}

func (r *CreatePoolRequest) ToModel() *models.Pool {
	return &models.Pool{
		Name:                 r.Name,
		TournamentID:         r.TournamentID,
		MinTeams:             r.MinTeams,
		MaxTeams:             r.MaxTeams,
		MaxInvestmentCredits: r.MaxInvestmentCredits,
	}
}

func (r *CreatePoolRequest) ToScoringRules() []*models.ScoringRule {
	rules := make([]*models.ScoringRule, len(r.ScoringRules))
	for i, input := range r.ScoringRules {
		rules[i] = &models.ScoringRule{
			WinIndex:      input.WinIndex,
			PointsAwarded: input.PointsAwarded,
		}
	}
	return rules
}

type PoolAbilities struct {
	CanEditSettings     bool `json:"canEditSettings"`
	CanInviteUsers      bool `json:"canInviteUsers"`
	CanEditPortfolios   bool `json:"canEditPortfolios"`
	CanManageCoManagers bool `json:"canManageCoManagers"`
}

type PoolResponse struct {
	ID                   string         `json:"id"`
	Name                 string         `json:"name"`
	TournamentID         string         `json:"tournamentId"`
	OwnerID              string         `json:"ownerId"`
	MinTeams             int            `json:"minTeams"`
	MaxTeams             int            `json:"maxTeams"`
	MaxInvestmentCredits int            `json:"maxInvestmentCredits"`
	BudgetCredits        int            `json:"budgetCredits"`
	Visibility           string         `json:"visibility"`
	CreatedAt            time.Time      `json:"createdAt"`
	UpdatedAt            time.Time      `json:"updatedAt"`
	Abilities            *PoolAbilities `json:"abilities,omitempty"`
}

func NewPoolResponse(p *models.Pool) *PoolResponse {
	return &PoolResponse{
		ID:                   p.ID,
		Name:                 p.Name,
		TournamentID:         p.TournamentID,
		OwnerID:              p.OwnerID,
		MinTeams:             p.MinTeams,
		MaxTeams:             p.MaxTeams,
		MaxInvestmentCredits: p.MaxInvestmentCredits,
		BudgetCredits:        p.BudgetCredits,
		Visibility:           p.Visibility,
		CreatedAt:            p.CreatedAt,
		UpdatedAt:            p.UpdatedAt,
	}
}

type UpdatePoolRequest struct {
	Name                 *string `json:"name,omitempty"`
	MinTeams             *int    `json:"minTeams,omitempty"`
	MaxTeams             *int    `json:"maxTeams,omitempty"`
	MaxInvestmentCredits *int    `json:"maxInvestmentCredits,omitempty"`
}

func (r *UpdatePoolRequest) Validate() error {
	if r.Name == nil && r.MinTeams == nil && r.MaxTeams == nil && r.MaxInvestmentCredits == nil {
		return ErrFieldInvalid("body", "at least one field must be provided")
	}
	if r.Name != nil && strings.TrimSpace(*r.Name) == "" {
		return ErrFieldInvalid("name", "cannot be empty")
	}
	if r.MinTeams != nil && *r.MinTeams <= 0 {
		return ErrFieldInvalid("minTeams", "must be greater than 0")
	}
	if r.MaxTeams != nil && *r.MaxTeams <= 0 {
		return ErrFieldInvalid("maxTeams", "must be greater than 0")
	}
	if r.MaxInvestmentCredits != nil && *r.MaxInvestmentCredits <= 0 {
		return ErrFieldInvalid("maxInvestmentCredits", "must be greater than 0")
	}
	return nil
}

type ReinviteRequest struct {
	Name         string `json:"name"`
	TournamentID string `json:"tournamentId"`
}

func (r *ReinviteRequest) Validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return ErrFieldRequired("name")
	}
	if strings.TrimSpace(r.TournamentID) == "" {
		return ErrFieldRequired("tournamentId")
	}
	return nil
}

type ReinviteResponse struct {
	Pool        *PoolResponse         `json:"pool"`
	Invitations []*InvitationResponse `json:"invitations"`
}

func NewPoolListResponse(pools []*models.Pool) []*PoolResponse {
	if pools == nil {
		return []*PoolResponse{}
	}

	responses := make([]*PoolResponse, len(pools))
	for i, p := range pools {
		responses[i] = NewPoolResponse(p)
	}
	return responses
}
