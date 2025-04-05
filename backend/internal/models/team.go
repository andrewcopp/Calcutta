package models

// Team represents a college basketball team in a tournament
type Team struct {
	ID           string `json:"id"`
	SchoolID     string `json:"schoolId"`
	TournamentID string `json:"tournamentId"`
	Byes         int    `json:"byes"` // Number of byes the team received (0 = no byes, 1 = first round bye, etc.)
	Wins         int    `json:"wins"` // Number of wins in the tournament
}
