package models

// Entry represents a user's entry into a Calcutta
type Entry struct {
	ID         string `json:"id"`
	UserID     string `json:"userId"`
	CalcuttaID string `json:"calcuttaId"`
}
