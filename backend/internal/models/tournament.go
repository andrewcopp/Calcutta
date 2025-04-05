package models

// Tournament represents a basketball tournament in the real world
type Tournament struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Rounds int    `json:"rounds"` // Total number of rounds in the tournament
}
