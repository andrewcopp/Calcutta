package models

import "time"

// Competition represents a named competition (e.g. "NCAA Tournament")
type Competition struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Season represents a year/season
type Season struct {
	ID   string `json:"id"`
	Year int    `json:"year"`
}

// Tournament represents a basketball tournament in the real world
type Tournament struct {
	ID                   string     `json:"id"`
	Name                 string     `json:"name"`
	Rounds               int        `json:"rounds"` // Total number of rounds in the tournament
	FinalFourTopLeft     string     `json:"finalFourTopLeft"`
	FinalFourBottomLeft  string     `json:"finalFourBottomLeft"`
	FinalFourTopRight    string     `json:"finalFourTopRight"`
	FinalFourBottomRight string     `json:"finalFourBottomRight"`
	StartingAt           *time.Time `json:"startingAt,omitempty"`
	CreatedAt            time.Time  `json:"createdAt"`
	UpdatedAt            time.Time  `json:"updatedAt"`
	DeletedAt            *time.Time `json:"deletedAt,omitempty"`
}

func (t *Tournament) HasStarted(now time.Time) bool {
	if t == nil || t.StartingAt == nil {
		return false
	}
	return !now.Before(*t.StartingAt)
}

const (
	TournamentEditDeniedReasonTournamentMissing = "tournament_missing"
	TournamentEditDeniedReasonTournamentStarted = "tournament_started"
)

func (t *Tournament) CanEditBids(now time.Time, isAdmin bool) (bool, string) {
	if t == nil {
		return false, TournamentEditDeniedReasonTournamentMissing
	}
	if isAdmin {
		return true, ""
	}
	if !t.HasStarted(now) {
		return true, ""
	}
	return false, TournamentEditDeniedReasonTournamentStarted
}
