package models

// PortfolioStanding holds point-in-time computed values for a portfolio.
// Unlike Portfolio (who), this represents how they're doing (when).
type PortfolioStanding struct {
	PortfolioID        string
	TotalReturns       float64
	FinishPosition     int
	IsTied             bool
	PayoutCents        int
	InTheMoney         bool
	ExpectedValue      *float64
	ProjectedFavorites *float64
}
