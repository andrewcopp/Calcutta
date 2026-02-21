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

type CreateCalcuttaRequest struct {
	Name         string             `json:"name"`
	TournamentID string             `json:"tournamentId"`
	MinTeams     int                `json:"minTeams"`
	MaxTeams     int                `json:"maxTeams"`
	MaxBidPoints int                `json:"maxBidPoints"`
	ScoringRules []ScoringRuleInput `json:"scoringRules"`
}

func (r *CreateCalcuttaRequest) Validate() error {
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

func (r *CreateCalcuttaRequest) ToModel() *models.Calcutta {
	return &models.Calcutta{
		Name:         r.Name,
		TournamentID: r.TournamentID,
		MinTeams:     r.MinTeams,
		MaxTeams:     r.MaxTeams,
		MaxBidPoints: r.MaxBidPoints,
	}
}

func (r *CreateCalcuttaRequest) ToScoringRules() []*models.CalcuttaRound {
	rounds := make([]*models.CalcuttaRound, len(r.ScoringRules))
	for i, rule := range r.ScoringRules {
		rounds[i] = &models.CalcuttaRound{
			Round:  rule.WinIndex,
			Points: rule.PointsAwarded,
		}
	}
	return rounds
}

type CalcuttaAbilities struct {
	CanEditSettings     bool `json:"canEditSettings"`
	CanInviteUsers      bool `json:"canInviteUsers"`
	CanEditEntries      bool `json:"canEditEntries"`
	CanManageCoManagers bool `json:"canManageCoManagers"`
}

type CalcuttaResponse struct {
	ID           string             `json:"id"`
	Name         string             `json:"name"`
	TournamentID string             `json:"tournamentId"`
	OwnerID      string             `json:"ownerId"`
	MinTeams     int                `json:"minTeams"`
	MaxTeams     int                `json:"maxTeams"`
	MaxBidPoints int                `json:"maxBidPoints"`
	Visibility   string             `json:"visibility"`
	CreatedAt    time.Time          `json:"createdAt"`
	UpdatedAt    time.Time          `json:"updatedAt"`
	Abilities    *CalcuttaAbilities `json:"abilities,omitempty"`
}

func NewCalcuttaResponse(c *models.Calcutta) *CalcuttaResponse {
	return &CalcuttaResponse{
		ID:           c.ID,
		Name:         c.Name,
		TournamentID: c.TournamentID,
		OwnerID:      c.OwnerID,
		MinTeams:     c.MinTeams,
		MaxTeams:     c.MaxTeams,
		MaxBidPoints: c.MaxBidPoints,
		Visibility:   c.Visibility,
		CreatedAt:    c.CreatedAt,
		UpdatedAt:    c.UpdatedAt,
	}
}

type UpdateCalcuttaRequest struct {
	Name     *string `json:"name,omitempty"`
	MinTeams *int    `json:"minTeams,omitempty"`
	MaxTeams *int    `json:"maxTeams,omitempty"`
	MaxBidPoints *int `json:"maxBidPoints,omitempty"`
}

func (r *UpdateCalcuttaRequest) Validate() error {
	if r.Name == nil && r.MinTeams == nil && r.MaxTeams == nil && r.MaxBidPoints == nil {
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
	if r.MaxBidPoints != nil && *r.MaxBidPoints <= 0 {
		return ErrFieldInvalid("maxBidPoints", "must be greater than 0")
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
	Calcutta    *CalcuttaResponse     `json:"calcutta"`
	Invitations []*InvitationResponse `json:"invitations"`
}

func NewCalcuttaListResponse(calcuttas []*models.Calcutta) []*CalcuttaResponse {
	if calcuttas == nil {
		return []*CalcuttaResponse{}
	}

	responses := make([]*CalcuttaResponse, len(calcuttas))
	for i, c := range calcuttas {
		responses[i] = NewCalcuttaResponse(c)
	}
	return responses
}
