package simulation

// MatchupKey uniquely identifies a game matchup for probability lookup.
type MatchupKey struct {
	GameID  string
	Team1ID string
	Team2ID string
}

// TeamSimulationResult holds the outcome for a single team in a single simulation.
type TeamSimulationResult struct {
	SimID      int
	TeamID     string
	Wins       int
	Byes       int
	IsEliminated bool
}

// Options configures simulation execution.
type Options struct {
	Workers int
}

// ProbabilityProvider returns the probability that team1 beats team2 in a given game.
type ProbabilityProvider interface {
	Prob(gameID string, team1ID string, team2ID string) float64
}
