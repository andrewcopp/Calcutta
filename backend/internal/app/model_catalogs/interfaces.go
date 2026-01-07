package model_catalogs

import (
	"context"

	tsim "github.com/andrewcopp/Calcutta/backend/internal/app/tournament_simulation"
	"github.com/andrewcopp/Calcutta/backend/pkg/models"
)

type TeamAdvancementProbs struct {
	TeamID     string
	ProbPI     float64
	ReachR64   float64
	ReachR32   float64
	ReachS16   float64
	ReachE8    float64
	ReachFF    float64
	ReachChamp float64
	WinChamp   float64
}

type AdvancementModelInput struct {
	TournamentID string
	Teams        []*models.TournamentTeam
	MatchupProbs map[tsim.MatchupKey]float64
}

type AdvancementArtifactPayload struct {
	ByTeam map[string]TeamAdvancementProbs
}

type AdvancementModelInterface interface {
	Descriptor() ModelDescriptor
	Compute(ctx context.Context, in AdvancementModelInput) (*AdvancementArtifactPayload, error)
}

type MarketShareTeamFeatures struct {
	TeamID   string
	Seed     int
	Region   string
	School   string
	Features map[string]float64
	Meta     map[string]string
}

type MarketShareModelInput struct {
	CalcuttaID   string
	TournamentID string
	Teams        []MarketShareTeamFeatures
}

type MarketShareArtifactPayload struct {
	PredictedShareByTeam map[string]float64
}

type MarketShareModelInterface interface {
	Descriptor() ModelDescriptor
	Compute(ctx context.Context, in MarketShareModelInput) (*MarketShareArtifactPayload, error)
}

type EntryOptimizerInput struct {
	CalcuttaID string

	BudgetPoints int
	MinTeams     int
	MaxTeams     int
	MinBidPoints int
	MaxBidPoints int

	ExpectedPointsByTeam map[string]float64
	MarketPointsByTeam   map[string]float64
}

type EntryOptimizerArtifactPayload struct {
	BidsByTeam map[string]int
}

type EntryOptimizerInterface interface {
	Descriptor() ModelDescriptor
	Compute(ctx context.Context, in EntryOptimizerInput) (*EntryOptimizerArtifactPayload, error)
}

type SimulationModelInput struct {
	TournamentID string
	Teams        []*models.TournamentTeam
	MatchupProbs map[tsim.MatchupKey]float64
	NSims        int
	Seed         int
}

type SimulationTeamResult struct {
	TeamID     string
	Wins       int
	Byes       int
	Eliminated bool
}

type SimulationArtifactPayload struct {
	Simulations map[int][]SimulationTeamResult
}

type SimulationModelInterface interface {
	Descriptor() ModelDescriptor
	Compute(ctx context.Context, in SimulationModelInput) (*SimulationArtifactPayload, error)
}
