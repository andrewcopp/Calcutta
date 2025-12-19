package services

import (
	"github.com/andrewcopp/Calcutta/backend/internal/ports"
)

type CalcuttaServicePorts struct {
	CalcuttaReader  ports.CalcuttaReader
	CalcuttaWriter  ports.CalcuttaWriter
	EntryReader     ports.EntryReader
	PortfolioReader ports.PortfolioReader
	PortfolioWriter ports.PortfolioWriter
	RoundWriter     ports.RoundWriter
	TeamReader      ports.TournamentTeamReader
}

// CalcuttaService handles business logic for Calcutta auctions
type CalcuttaService struct {
	ports CalcuttaServicePorts
}

// NewCalcuttaService creates a new CalcuttaService
func NewCalcuttaService(ports CalcuttaServicePorts) *CalcuttaService {
	return &CalcuttaService{ports: ports}
}
