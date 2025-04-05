package models

// CalcuttaPortfolioTeam represents a team investment within a user's Calcutta portfolio
type CalcuttaPortfolioTeam struct {
	ID                  string  `json:"id"`
	PortfolioID         string  `json:"portfolioId"`
	TeamID              string  `json:"teamId"`
	OwnershipPercentage float64 `json:"ownershipPercentage"`
	PointsEarned        int     `json:"pointsEarned"`
}
