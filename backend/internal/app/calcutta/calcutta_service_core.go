package calcutta

import (
	"github.com/andrewcopp/Calcutta/backend/internal/ports"
)

type Ports struct {
	CalcuttaReader    ports.CalcuttaReader
	CalcuttaWriter    ports.CalcuttaWriter
	EntryReader       ports.EntryReader
	EntryWriter       ports.EntryWriter
	PayoutReader      ports.PayoutReader
	PayoutWriter      ports.PayoutWriter
	PortfolioReader   ports.PortfolioReader
	RoundReader       ports.RoundReader
	RoundWriter       ports.RoundWriter
	TeamReader        ports.TournamentTeamReader
	InvitationReader  ports.CalcuttaInvitationReader
	InvitationWriter  ports.CalcuttaInvitationWriter
}

// CalcuttaService handles business logic for Calcutta auctions
type Service struct {
	ports Ports
}

// NewCalcuttaService creates a new CalcuttaService
func New(ports Ports) *Service {
	return &Service{ports: ports}
}
