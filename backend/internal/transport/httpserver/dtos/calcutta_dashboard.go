package dtos

import (
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

type CalcuttaDashboardResponse struct {
	Calcutta             *CalcuttaResponse        `json:"calcutta"`
	TournamentStartingAt *time.Time               `json:"tournamentStartingAt,omitempty"`
	BiddingOpen          bool                     `json:"biddingOpen"`
	TotalEntries         int                      `json:"totalEntries"`
	CurrentUserEntry     *EntryResponse           `json:"currentUserEntry,omitempty"`
	Abilities            *CalcuttaAbilities       `json:"abilities,omitempty"`
	Entries              []*EntryResponse         `json:"entries"`
	EntryTeams           []*EntryTeamResponse     `json:"entryTeams"`
	Portfolios           []*PortfolioResponse     `json:"portfolios"`
	PortfolioTeams       []*PortfolioTeamResponse `json:"portfolioTeams"`
	Schools              []*SchoolResponse        `json:"schools"`
	TournamentTeams      []*TournamentTeamResponse `json:"tournamentTeams"`
	ScoringRules         []*ScoringRuleResponse   `json:"scoringRules"`
	Payouts              []*PayoutResponse        `json:"payouts"`
}

type ScoringRuleResponse struct {
	WinIndex      int `json:"winIndex"`
	PointsAwarded int `json:"pointsAwarded"`
}

func NewScoringRuleResponse(r *models.CalcuttaRound) *ScoringRuleResponse {
	return &ScoringRuleResponse{
		WinIndex:      r.Round,
		PointsAwarded: r.Points,
	}
}

func NewScoringRuleListResponse(rounds []*models.CalcuttaRound) []*ScoringRuleResponse {
	responses := make([]*ScoringRuleResponse, 0, len(rounds))
	for _, r := range rounds {
		responses = append(responses, NewScoringRuleResponse(r))
	}
	return responses
}

type PayoutResponse struct {
	Position    int `json:"position"`
	AmountCents int `json:"amountCents"`
}

func NewPayoutResponse(p *models.CalcuttaPayout) *PayoutResponse {
	return &PayoutResponse{
		Position:    p.Position,
		AmountCents: p.AmountCents,
	}
}

func NewPayoutListResponse(payouts []*models.CalcuttaPayout) []*PayoutResponse {
	responses := make([]*PayoutResponse, 0, len(payouts))
	for _, p := range payouts {
		responses = append(responses, NewPayoutResponse(p))
	}
	return responses
}

type CalcuttaWithRankingResponse struct {
	*CalcuttaResponse
	HasEntry              bool                    `json:"hasEntry"`
	TournamentStartingAt  *time.Time              `json:"tournamentStartingAt,omitempty"`
	Ranking               *CalcuttaRankingResponse `json:"ranking,omitempty"`
}

type CalcuttaRankingResponse struct {
	Rank         int     `json:"rank"`
	TotalEntries int     `json:"totalEntries"`
	Points       float64 `json:"points"`
}
