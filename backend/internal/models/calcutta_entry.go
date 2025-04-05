package models

// CalcuttaEntry represents a user's entry into a Calcutta
type CalcuttaEntry struct {
	ID         string `json:"id"`
	UserID     string `json:"userId"`
	CalcuttaID string `json:"calcuttaId"`
}
