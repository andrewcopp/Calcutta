package dtos

type InvitePreviewResponse struct {
	FirstName            string  `json:"firstName"`
	CalcuttaName         string  `json:"calcuttaName"`
	CommissionerName     string  `json:"commissionerName"`
	TournamentStartingAt *string `json:"tournamentStartingAt,omitempty"`
}
