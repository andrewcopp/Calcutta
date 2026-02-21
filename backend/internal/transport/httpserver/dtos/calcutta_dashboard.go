package dtos

import "time"

type CalcuttaDashboardResponse struct {
	Calcutta              *CalcuttaResponse        `json:"calcutta"`
	TournamentStartingAt  *time.Time               `json:"tournamentStartingAt,omitempty"`
	BiddingOpen           bool                     `json:"biddingOpen"`
	TotalEntries          int                      `json:"totalEntries"`
	CurrentUserEntry      *EntryResponse           `json:"currentUserEntry,omitempty"`
	Abilities             *CalcuttaAbilities        `json:"abilities,omitempty"`
	Entries               []*EntryResponse          `json:"entries"`
	EntryTeams            []*EntryTeamResponse      `json:"entryTeams"`
	Portfolios            []*PortfolioResponse      `json:"portfolios"`
	PortfolioTeams        []*PortfolioTeamResponse  `json:"portfolioTeams"`
	Schools               []*SchoolResponse         `json:"schools"`
	TournamentTeams       []*TournamentTeamResponse `json:"tournamentTeams"`
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
