package models

import "time"

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
	Created              time.Time  `json:"created"`
	Updated              time.Time  `json:"updated"`
	Deleted              *time.Time `json:"deleted,omitempty"`
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

func (t *Tournament) CanEditCalcutta(now time.Time, isAdmin bool) (bool, string) {
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

func (t *Tournament) CanEditBids(now time.Time, isAdmin bool) (bool, string) {
	return t.CanEditCalcutta(now, isAdmin)
}

func (t *Tournament) CanEditEntries(now time.Time, isAdmin bool) (bool, string) {
	return t.CanEditCalcutta(now, isAdmin)
}

// TournamentState represents the current state of a tournament
type TournamentState string

const (
	TournamentStateFuture     TournamentState = "future"
	TournamentStateInProgress TournamentState = "in_progress"
	TournamentStateCompleted  TournamentState = "completed"
)

// GetTournamentState determines the current state of a tournament based on its games
// This is a helper function that would be used by a service layer
func GetTournamentState(games []TournamentGame) TournamentState {
	if len(games) == 0 {
		return TournamentStateFuture
	}

	hasFutureGames := false
	hasInProgressGames := false
	hasCompletedGames := false

	for _, game := range games {
		status := game.GetGameStatus()

		switch status {
		case "future":
			hasFutureGames = true
		case "in_progress":
			hasInProgressGames = true
		case "completed":
			hasCompletedGames = true
		}
	}

	// If any games are actively in progress, the tournament is in progress
	if hasInProgressGames {
		return TournamentStateInProgress
	}

	// If we have both completed and future games, we're between rounds (still in progress)
	if hasCompletedGames && hasFutureGames {
		return TournamentStateInProgress
	}

	// If we only have future games, the tournament hasn't started
	if hasFutureGames && !hasCompletedGames {
		return TournamentStateFuture
	}

	// If we only have completed games, the tournament is done
	if hasCompletedGames && !hasFutureGames {
		return TournamentStateCompleted
	}

	// Default to in-progress if we can't determine the state
	return TournamentStateInProgress
}
