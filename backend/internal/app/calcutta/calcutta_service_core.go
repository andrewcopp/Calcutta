package calcutta

import (
	"github.com/andrewcopp/Calcutta/backend/internal/ports"
)

type Ports struct {
	CalcuttaReader  ports.CalcuttaReader
	CalcuttaWriter  ports.CalcuttaWriter
	EntryReader     ports.EntryReader
	EntryWriter     ports.EntryWriter
	PayoutReader    ports.PayoutReader
	PortfolioReader ports.PortfolioReader
	RoundReader     ports.RoundReader
	RoundWriter     ports.RoundWriter
	TeamReader      ports.TournamentTeamReader
}

// CalcuttaService handles business logic for Calcutta auctions
type Service struct {
	ports Ports
}

// NewCalcuttaService creates a new CalcuttaService
func New(ports Ports) *Service {
	return &Service{ports: ports}
}
