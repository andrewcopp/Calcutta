package models

// EntryStanding holds point-in-time computed values for an entry.
// Unlike CalcuttaEntry (who), this represents how they're doing (when).
type EntryStanding struct {
	EntryID        string
	TotalPoints    float64
	FinishPosition int
	IsTied         bool
	PayoutCents    int
	InTheMoney     bool
	ProjectedEV    *float64
}
