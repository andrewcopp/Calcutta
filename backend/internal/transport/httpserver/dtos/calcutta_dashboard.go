package dtos

import "time"

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
	RoundStandings       []*RoundStandingGroup     `json:"roundStandings"`
}

type RoundStandingGroup struct {
	Round   int                   `json:"round"`
	Entries []*RoundStandingEntry `json:"entries"`
}

type RoundStandingEntry struct {
	EntryID        string  `json:"entryId"`
	TotalPoints    float64 `json:"totalPoints"`
	FinishPosition int     `json:"finishPosition"`
	IsTied         bool    `json:"isTied"`
	PayoutCents    int     `json:"payoutCents"`
	InTheMoney     bool    `json:"inTheMoney"`
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
