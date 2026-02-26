package dtos

import (
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

type CalcuttaDashboardResponse struct {
	Calcutta             *CalcuttaResponse         `json:"calcutta"`
	TournamentStartingAt *time.Time                `json:"tournamentStartingAt,omitempty"`
	BiddingOpen          bool                      `json:"biddingOpen"`
	TotalEntries         int                       `json:"totalEntries"`
	CurrentUserEntry     *EntryResponse            `json:"currentUserEntry,omitempty"`
	Abilities            *CalcuttaAbilities        `json:"abilities,omitempty"`
	Entries              []*EntryResponse          `json:"entries"`
	EntryTeams           []*EntryTeamResponse      `json:"entryTeams"`
	Portfolios           []*PortfolioResponse      `json:"portfolios"`
	PortfolioTeams       []*PortfolioTeamResponse  `json:"portfolioTeams"`
	Schools              []*SchoolResponse         `json:"schools"`
	TournamentTeams      []*TournamentTeamResponse `json:"tournamentTeams"`
	RoundStandings       []*RoundStandingGroup        `json:"roundStandings"`
	FinalFourOutcomes    []*FinalFourOutcomeResponse  `json:"finalFourOutcomes,omitempty"`
}

type RoundStandingGroup struct {
	Round   int                   `json:"round"`
	Entries []*RoundStandingEntry `json:"entries"`
}

type RoundStandingEntry struct {
	EntryID            string   `json:"entryId"`
	TotalPoints        float64  `json:"totalPoints"`
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

type CalcuttaWithRankingResponse struct {
	*CalcuttaResponse
	HasEntry             bool                     `json:"hasEntry"`
	TournamentStartingAt *time.Time               `json:"tournamentStartingAt,omitempty"`
	Ranking              *CalcuttaRankingResponse  `json:"ranking,omitempty"`
}

type CalcuttaRankingResponse struct {
	Rank         int     `json:"rank"`
	TotalEntries int     `json:"totalEntries"`
	Points       float64 `json:"points"`
}
