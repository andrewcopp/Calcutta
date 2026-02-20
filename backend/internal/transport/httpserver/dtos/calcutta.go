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

type CalcuttaResponse struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	TournamentID    string     `json:"tournamentId"`
	OwnerID         string     `json:"ownerId"`
	MinTeams        int        `json:"minTeams"`
	MaxTeams        int        `json:"maxTeams"`
	MaxBid          int        `json:"maxBid"`
	BiddingOpen     bool       `json:"biddingOpen"`
	BiddingLockedAt *time.Time `json:"biddingLockedAt,omitempty"`
	Visibility      string     `json:"visibility"`
	Created         time.Time  `json:"created"`
	Updated         time.Time  `json:"updated"`
}

func NewCalcuttaResponse(c *models.Calcutta) *CalcuttaResponse {
	return &CalcuttaResponse{
		ID:              c.ID,
		Name:            c.Name,
		TournamentID:    c.TournamentID,
		OwnerID:         c.OwnerID,
		MinTeams:        c.MinTeams,
		MaxTeams:        c.MaxTeams,
		MaxBid:          c.MaxBid,
		BiddingOpen:     c.BiddingOpen,
		BiddingLockedAt: c.BiddingLockedAt,
		Visibility:      c.Visibility,
		Created:         c.Created,
		Updated:         c.Updated,
	}
}

type UpdateCalcuttaRequest struct {
	Name        *string `json:"name,omitempty"`
	MinTeams    *int    `json:"minTeams,omitempty"`
	MaxTeams    *int    `json:"maxTeams,omitempty"`
	MaxBid      *int    `json:"maxBid,omitempty"`
	BiddingOpen *bool   `json:"biddingOpen,omitempty"`
}

func (r *UpdateCalcuttaRequest) Validate() error {
	if r.Name == nil && r.MinTeams == nil && r.MaxTeams == nil && r.MaxBid == nil && r.BiddingOpen == nil {
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
	if r.MaxBid != nil && *r.MaxBid <= 0 {
		return ErrFieldInvalid("maxBid", "must be greater than 0")
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
