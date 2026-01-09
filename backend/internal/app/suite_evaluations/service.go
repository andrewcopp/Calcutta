package suite_evaluations

import (
	"context"
	"strings"
	"time"

	appcalcutta "github.com/andrewcopp/Calcutta/backend/internal/app/calcutta"
	"github.com/andrewcopp/Calcutta/backend/internal/app/simulation_game_outcomes"
	"github.com/andrewcopp/Calcutta/backend/pkg/models"
)

type EvaluationListItem struct {
	ID                        string
	SimulationBatchID         *string
	CohortID                  string
	CohortName                string
	OptimizerKey              string
	NSims                     int
	Seed                      int
	OurRank                   *int
	OurMeanNormalizedPayout   *float64
	OurMedianNormalizedPayout *float64
	OurPTop1                  *float64
	OurPInMoney               *float64
	TotalSimulations          *int
	CalcuttaID                string
	SimulatedCalcuttaID       *string
	GameOutcomeRunID          *string
	MarketShareRunID          *string
	StrategyGenerationRunID   *string
	CalcuttaEvaluationRunID   *string
	RealizedFinishPosition    *int
	RealizedIsTied            *bool
	RealizedInTheMoney        *bool
	RealizedPayoutCents       *int
	RealizedTotalPoints       *float64
	StartingStateKey          string
	ExcludedEntryName         *string
	Status                    string
	ClaimedAt                 *time.Time
	ClaimedBy                 *string
	ErrorMessage              *string
	CreatedAt                 time.Time
	UpdatedAt                 time.Time
}

type SnapshotEntryTeam struct {
	TeamID    string
	School    string
	Seed      int
	Region    string
	BidPoints int
}

type SnapshotEntry struct {
	SnapshotEntryID string
	DisplayName     string
	IsSynthetic     bool
	Teams           []SnapshotEntryTeam
}

type PortfolioBid struct {
	TeamID      string
	SchoolName  string
	Seed        int
	Region      string
	BidPoints   int
	ExpectedROI float64
}

type OurStrategyPerformance struct {
	Rank                   int
	EntryName              string
	MeanNormalizedPayout   float64
	MedianNormalizedPayout float64
	PTop1                  float64
	PInMoney               float64
	TotalSimulations       int
}

type EntryPerformance struct {
	Rank                 int
	EntryName            string
	SnapshotEntryID      *string
	MeanNormalizedPayout float64
	PTop1                float64
	PInMoney             float64
}

type Finish struct {
	FinishPosition int
	IsTied         bool
	InTheMoney     bool
	PayoutCents    int
	TotalPoints    float64
}

type CreateSimulationParams struct {
	SimulationRunBatchID *string
	CohortID             string
	CalcuttaID           string
	SimulatedCalcuttaID  *string
	GameOutcomeRunID     *string
	GameOutcomeSpec      *simulation_game_outcomes.Spec
	MarketShareRunID     *string
	OptimizerKey         string
	NSims                *int
	Seed                 *int
	StartingStateKey     string
	ExcludedEntryName    *string
}

type CreateSimulationResult struct {
	ID     string
	Status string
}

type Repo interface {
	ListEvaluations(ctx context.Context, calcuttaID, cohortID, simulationBatchID *string, limit, offset int) ([]EvaluationListItem, error)
	GetEvaluation(ctx context.Context, id string) (*EvaluationListItem, error)

	GetSnapshotEntry(ctx context.Context, evalID, snapshotEntryID string) (*SnapshotEntry, error)
	ListPortfolioBids(ctx context.Context, strategyGenerationRunID string) ([]PortfolioBid, error)
	GetOurStrategyPerformance(ctx context.Context, calcuttaEvaluationRunID, evalID string) (*OurStrategyPerformance, error)
	ListEntryPerformance(ctx context.Context, calcuttaEvaluationRunID string) ([]EntryPerformance, error)

	GetSimulationBatchConfig(ctx context.Context, simulationBatchID string) (*SimulationBatchConfig, error)
	GetCohortOptimizerKey(ctx context.Context, cohortID string) (string, error)
	GetTournamentIDForCalcutta(ctx context.Context, calcuttaID string) (string, error)
	GetTournamentIDForSimulatedCalcutta(ctx context.Context, simulatedCalcuttaID string) (string, error)
	GetLatestGameOutcomeRunID(ctx context.Context, tournamentID, algorithmID string) (string, error)
	GetLatestMarketShareRunID(ctx context.Context, calcuttaID, algorithmID string) (string, error)
	UpsertSyntheticCalcutta(ctx context.Context, cohortID, calcuttaID string) (string, *string, *string, error)
	EnsureSyntheticSnapshot(ctx context.Context, syntheticCalcuttaID, calcuttaID string, excludedEntryName *string) error
	CreateSimulationRun(ctx context.Context, p CreateSimulationParams, syntheticCalcuttaID string) (*CreateSimulationResult, error)

	LoadPayouts(ctx context.Context, calcuttaID string) ([]*models.CalcuttaPayout, error)
	LoadTeamPoints(ctx context.Context, calcuttaID string) (map[string]float64, error)
	LoadEntryBids(ctx context.Context, calcuttaID string) ([]EntryBidRow, error)
	LoadStrategyBids(ctx context.Context, strategyGenerationRunID string) (map[string]float64, error)
}

type SimulationBatchConfig struct {
	CohortID           string
	OptimizerKey       *string
	NSims              *int
	Seed               *int
	StartingStateKey   string
	ExcludedEntryName  *string
	GameOutcomesAlgID  string
	MarketShareAlgID   string
	CohortOptimizerKey string
	CohortNSims        int
	CohortSeed         int
}

type EntryBidRow struct {
	EntryID   string
	Name      string
	CreatedAt time.Time
	TeamID    string
	BidPoints int
}

type Service struct {
	repo Repo
}

func New(repo Repo) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListEvaluations(ctx context.Context, calcuttaID, cohortID, simulationBatchID *string, limit, offset int) ([]EvaluationListItem, error) {
	return s.repo.ListEvaluations(ctx, calcuttaID, cohortID, simulationBatchID, limit, offset)
}

func (s *Service) GetEvaluation(ctx context.Context, id string) (*EvaluationListItem, error) {
	return s.repo.GetEvaluation(ctx, id)
}

func (s *Service) GetSnapshotEntry(ctx context.Context, evalID, snapshotEntryID string) (*SnapshotEntry, error) {
	return s.repo.GetSnapshotEntry(ctx, evalID, snapshotEntryID)
}

func (s *Service) ListPortfolioBids(ctx context.Context, strategyGenerationRunID string) ([]PortfolioBid, error) {
	return s.repo.ListPortfolioBids(ctx, strategyGenerationRunID)
}

func (s *Service) GetOurStrategyPerformance(ctx context.Context, calcuttaEvaluationRunID, evalID string) (*OurStrategyPerformance, error) {
	return s.repo.GetOurStrategyPerformance(ctx, calcuttaEvaluationRunID, evalID)
}

func (s *Service) ListEntryPerformance(ctx context.Context, calcuttaEvaluationRunID string) ([]EntryPerformance, error) {
	return s.repo.ListEntryPerformance(ctx, calcuttaEvaluationRunID)
}

func (s *Service) CreateEvaluation(ctx context.Context, p CreateSimulationParams) (*CreateSimulationResult, error) {
	if p.GameOutcomeSpec == nil {
		tmp := &simulation_game_outcomes.Spec{Kind: "kenpom", Sigma: 10.0}
		tmp.Normalize()
		p.GameOutcomeSpec = tmp
	}
	if err := p.GameOutcomeSpec.Validate(); err != nil {
		return nil, err
	}

	cfg := (*SimulationBatchConfig)(nil)
	if p.SimulationRunBatchID != nil && *p.SimulationRunBatchID != "" {
		loaded, err := s.repo.GetSimulationBatchConfig(ctx, *p.SimulationRunBatchID)
		if err != nil {
			return nil, err
		}
		cfg = loaded
		p.CohortID = cfg.CohortID
		if p.OptimizerKey == "" {
			if cfg.OptimizerKey != nil && *cfg.OptimizerKey != "" {
				p.OptimizerKey = *cfg.OptimizerKey
			} else {
				p.OptimizerKey = cfg.CohortOptimizerKey
			}
		}
		if p.NSims == nil {
			if cfg.NSims != nil {
				p.NSims = cfg.NSims
			} else if cfg.CohortNSims > 0 {
				v := cfg.CohortNSims
				p.NSims = &v
			}
		}
		if p.Seed == nil {
			if cfg.Seed != nil {
				p.Seed = cfg.Seed
			} else if cfg.CohortSeed != 0 {
				v := cfg.CohortSeed
				p.Seed = &v
			}
		}
		if p.StartingStateKey == "" {
			p.StartingStateKey = cfg.StartingStateKey
		}
		if p.ExcludedEntryName == nil {
			p.ExcludedEntryName = cfg.ExcludedEntryName
		}

	} else {
		if p.OptimizerKey == "" {
			k, _ := s.repo.GetCohortOptimizerKey(ctx, p.CohortID)
			p.OptimizerKey = k
		}
	}

	synthID := ""
	snapshotID := (*string)(nil)
	existingExcluded := (*string)(nil)
	if strings.TrimSpace(p.CalcuttaID) != "" {
		var err error
		synthID, snapshotID, existingExcluded, err = s.repo.UpsertSyntheticCalcutta(ctx, p.CohortID, p.CalcuttaID)
		if err != nil {
			return nil, err
		}
		_ = snapshotID
		if snapshotID == nil || strings.TrimSpace(*snapshotID) == "" {
			if err := s.repo.EnsureSyntheticSnapshot(ctx, synthID, p.CalcuttaID, existingExcluded); err != nil {
				return nil, err
			}
		}
	}
	return s.repo.CreateSimulationRun(ctx, p, synthID)
}

func (s *Service) ComputeHypotheticalFinishByEntryName(ctx context.Context, calcuttaID string, strategyGenerationRunID string) (map[string]*Finish, bool, error) {
	if calcuttaID == "" || strategyGenerationRunID == "" {
		return nil, false, nil
	}
	payouts, err := s.repo.LoadPayouts(ctx, calcuttaID)
	if err != nil {
		return nil, false, err
	}
	teamPoints, err := s.repo.LoadTeamPoints(ctx, calcuttaID)
	if err != nil {
		return nil, false, err
	}
	rows, err := s.repo.LoadEntryBids(ctx, calcuttaID)
	if err != nil {
		return nil, false, err
	}
	ourBids, err := s.repo.LoadStrategyBids(ctx, strategyGenerationRunID)
	if err != nil {
		return nil, false, err
	}
	if len(ourBids) == 0 {
		return nil, false, nil
	}

	entryByID := make(map[string]*models.CalcuttaEntry)
	entryBids := make(map[string]map[string]float64)
	existingTotalBid := make(map[string]float64)
	for _, r := range rows {
		if _, ok := entryByID[r.EntryID]; !ok {
			entryByID[r.EntryID] = &models.CalcuttaEntry{ID: r.EntryID, Name: r.Name, CalcuttaID: calcuttaID, Created: r.CreatedAt}
			entryBids[r.EntryID] = make(map[string]float64)
		}
		entryBids[r.EntryID][r.TeamID] += float64(r.BidPoints)
		existingTotalBid[r.TeamID] += float64(r.BidPoints)
	}

	entries := make([]*models.CalcuttaEntry, 0, len(entryByID)+1)
	for entryID, e := range entryByID {
		total := 0.0
		for teamID, bid := range entryBids[entryID] {
			pts, ok := teamPoints[teamID]
			if !ok {
				continue
			}
			den := existingTotalBid[teamID] + ourBids[teamID]
			if den <= 0 {
				continue
			}
			total += pts * (bid / den)
		}
		e.TotalPoints = total
		entries = append(entries, e)
	}

	ourID := "our_strategy"
	ourCreated := time.Now()
	ourTotal := 0.0
	for teamID, bid := range ourBids {
		pts, ok := teamPoints[teamID]
		if !ok {
			continue
		}
		den := existingTotalBid[teamID] + bid
		if den <= 0 {
			continue
		}
		ourTotal += pts * (bid / den)
	}
	entries = append(entries, &models.CalcuttaEntry{ID: ourID, Name: "Our Strategy", CalcuttaID: calcuttaID, TotalPoints: ourTotal, Created: ourCreated})

	_, results := appcalcutta.ComputeEntryPlacementsAndPayouts(entries, payouts)
	out := make(map[string]*Finish)
	for _, e := range entries {
		res, ok := results[e.ID]
		if !ok {
			continue
		}
		out[e.Name] = &Finish{FinishPosition: res.FinishPosition, IsTied: res.IsTied, InTheMoney: res.InTheMoney, PayoutCents: res.PayoutCents, TotalPoints: e.TotalPoints}
	}
	return out, true, nil
}
