package models

import "time"

// Default constraints for new pools.
const (
	DefaultMinTeams              = 3
	DefaultMaxTeams              = 10
	DefaultMaxInvestmentCredits  = 50
	DefaultBudgetCredits         = 100
)

// ApplyDefaults fills in zero-value constraint fields with sensible defaults.
func (p *Pool) ApplyDefaults() {
	if p.MinTeams == 0 {
		p.MinTeams = DefaultMinTeams
	}
	if p.MaxTeams == 0 {
		p.MaxTeams = DefaultMaxTeams
	}
	if p.MaxInvestmentCredits == 0 {
		p.MaxInvestmentCredits = DefaultMaxInvestmentCredits
	}
	if p.BudgetCredits == 0 {
		p.BudgetCredits = DefaultBudgetCredits
	}
}

// Pool represents an investment pool for a tournament
type Pool struct {
	ID                   string     `json:"id"`
	TournamentID         string     `json:"tournamentId"`
	OwnerID              string     `json:"ownerId"`
	CreatedBy            string     `json:"createdBy"`
	Name                 string     `json:"name"`
	MinTeams             int        `json:"minTeams"`
	MaxTeams             int        `json:"maxTeams"`
	MaxInvestmentCredits int        `json:"maxInvestmentCredits"`
	BudgetCredits        int        `json:"budgetCredits"`
	EntryFeeCents        int        `json:"entryFeeCents"`
	Visibility           string     `json:"visibility"`
	CreatedAt            time.Time  `json:"createdAt"`
	UpdatedAt            time.Time  `json:"updatedAt"`
	DeletedAt            *time.Time `json:"deletedAt,omitempty"`
}
