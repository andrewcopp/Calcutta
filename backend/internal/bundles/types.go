package bundles

import "time"

type SchoolsBundle struct {
	Version     int           `json:"version"`
	GeneratedAt time.Time     `json:"generated_at"`
	Schools     []SchoolEntry `json:"schools"`
}

type SchoolEntry struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type TournamentBundle struct {
	Version     int              `json:"version"`
	GeneratedAt time.Time        `json:"generated_at"`
	Tournament  TournamentRecord `json:"tournament"`
	Teams       []TeamRecord     `json:"teams"`
}

type TournamentRecord struct {
	ImportKey            string     `json:"import_key"`
	Name                 string     `json:"name"`
	Rounds               int        `json:"rounds"`
	StartingAt           *time.Time `json:"starting_at,omitempty"`
	FinalFourTopLeft     string     `json:"final_four_top_left"`
	FinalFourBottomLeft  string     `json:"final_four_bottom_left"`
	FinalFourTopRight    string     `json:"final_four_top_right"`
	FinalFourBottomRight string     `json:"final_four_bottom_right"`
}

type TeamRecord struct {
	SchoolSlug string        `json:"school_slug"`
	SchoolName string        `json:"school_name"`
	Seed       int           `json:"seed"`
	Region     string        `json:"region"`
	Byes       int           `json:"byes"`
	Wins       int           `json:"wins"`
	IsEliminated bool        `json:"is_eliminated"`
	KenPom     *KenPomRecord `json:"kenpom,omitempty"`
}

type KenPomRecord struct {
	NetRTG float64 `json:"net_rtg"`
	ORTG   float64 `json:"o_rtg"`
	DRTG   float64 `json:"d_rtg"`
	AdjT   float64 `json:"adj_t"`
}

type PoolBundle struct {
	Version     int              `json:"version"`
	GeneratedAt time.Time        `json:"generated_at"`
	Tournament  TournamentRef    `json:"tournament"`
	Pool        PoolRecord       `json:"pool"`
	Rounds      []RoundRecord    `json:"rounds"`
	Payouts     []PayoutRecord   `json:"payouts"`
	Portfolios  []PortfolioRecord `json:"portfolios"`
	Investments []InvestmentRecord `json:"investments"`
}

type TournamentRef struct {
	ImportKey string `json:"import_key"`
	Name      string `json:"name"`
}

type PoolRecord struct {
	Key   string   `json:"key"`
	Name  string   `json:"name"`
	Owner *UserRef `json:"owner,omitempty"`
}

type UserRef struct {
	Email     *string `json:"email,omitempty"`
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
}

type RoundRecord struct {
	Round  int `json:"round"`
	Points int `json:"points"`
}

type PayoutRecord struct {
	Position    int `json:"position"`
	AmountCents int `json:"amount_cents"`
}

type PortfolioRecord struct {
	Key       string  `json:"key"`
	Name      string  `json:"name"`
	UserName  *string `json:"user_name,omitempty"`
	UserEmail *string `json:"user_email,omitempty"`
}

type InvestmentRecord struct {
	PortfolioKey string `json:"portfolio_key"`
	SchoolSlug   string `json:"school_slug"`
	Credits      int    `json:"credits"`
}
