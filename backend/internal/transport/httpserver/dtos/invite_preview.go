package dtos

type InvitePreviewResponse struct {
	FirstName            string  `json:"firstName"`
	PoolName             string  `json:"poolName"`
	CommissionerName     string  `json:"commissionerName"`
	TournamentStartingAt *string `json:"tournamentStartingAt,omitempty"`
}
