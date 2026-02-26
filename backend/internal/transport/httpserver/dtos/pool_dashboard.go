package dtos

import (
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

type PoolDashboardResponse struct {
	Pool                 *PoolResponse               `json:"pool"`
	TournamentStartingAt *time.Time                  `json:"tournamentStartingAt,omitempty"`
	InvestingOpen        bool                        `json:"investingOpen"`
	TotalPortfolios      int                         `json:"totalPortfolios"`
	CurrentUserPortfolio *PortfolioResponse           `json:"currentUserPortfolio,omitempty"`
	Abilities            *PoolAbilities               `json:"abilities,omitempty"`
	Portfolios           []*PortfolioResponse         `json:"portfolios"`
	Investments          []*InvestmentResponse        `json:"investments"`
	OwnershipSummaries   []*OwnershipSummaryResponse  `json:"ownershipSummaries"`
	OwnershipDetails     []*OwnershipDetailResponse   `json:"ownershipDetails"`
	Schools              []*SchoolResponse            `json:"schools"`
	TournamentTeams      []*TournamentTeamResponse    `json:"tournamentTeams"`
	RoundStandings       []*RoundStandingGroup        `json:"roundStandings"`
	FinalFourOutcomes    []*FinalFourOutcomeResponse  `json:"finalFourOutcomes,omitempty"`
}

type RoundStandingGroup struct {
	Round   int                   `json:"round"`
	Entries []*RoundStandingEntry `json:"entries"`
}

type RoundStandingEntry struct {
	PortfolioID        string   `json:"portfolioId"`
	TotalReturns       float64  `json:"totalReturns"`
	FinishPosition     int      `json:"finishPosition"`
	IsTied             bool     `json:"isTied"`
	PayoutCents        int      `json:"payoutCents"`
	InTheMoney         bool     `json:"inTheMoney"`
	ExpectedValue      *float64 `json:"expectedValue,omitempty"`
	ProjectedFavorites *float64 `json:"projectedFavorites,omitempty"`
}

type FinalFourTeam struct {
	TeamID   string `json:"teamId"`
	SchoolID string `json:"schoolId"`
	Seed     int    `json:"seed"`
	Region   string `json:"region"`
}

type FinalFourOutcomeResponse struct {
	Semifinal1Winner *FinalFourTeam        `json:"semifinal1Winner"`
	Semifinal2Winner *FinalFourTeam        `json:"semifinal2Winner"`
	Champion         *FinalFourTeam        `json:"champion"`
	RunnerUp         *FinalFourTeam        `json:"runnerUp"`
	Entries          []*RoundStandingEntry `json:"entries"`
}

func NewFinalFourTeam(bt *models.BracketTeam) *FinalFourTeam {
	if bt == nil {
		return nil
	}
	return &FinalFourTeam{
		TeamID:   bt.TeamID,
		SchoolID: bt.SchoolID,
		Seed:     bt.Seed,
		Region:   bt.Region,
	}
}

type PoolWithRankingResponse struct {
	*PoolResponse
	HasPortfolio         bool                    `json:"hasPortfolio"`
	TournamentStartingAt *time.Time              `json:"tournamentStartingAt,omitempty"`
	Ranking              *PoolRankingResponse     `json:"ranking,omitempty"`
}

type PoolRankingResponse struct {
	Rank            int     `json:"rank"`
	TotalPortfolios int     `json:"totalPortfolios"`
	Returns         float64 `json:"returns"`
}
