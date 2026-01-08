package recommended_entry_bids

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	dbadapter "github.com/andrewcopp/Calcutta/backend/internal/adapters/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

type GenerateParams struct {
	CalcuttaID            string
	RunKey                string
	Name                  string
	OptimizerKey          string
	MarketShareRunID      *string
	SimulatedTournamentID *string
	MinBidPoints          int
	MaxBidPoints          int
	MinTeams              int
	MaxTeams              int
	BudgetPoints          int
}

type GenerateFromPredictionsParams struct {
	CalcuttaID           string
	RunKey               string
	Name                 string
	OptimizerKey         string
	GameOutcomeRunID     string
	MarketShareRunID     string
	ExcludedEntryName    string
	StartingStateKey     *string
	ExpectedPointsByTeam []ExpectedTeam
	PredictedShareByTeam map[string]float64
	MinBidPoints         int
	MaxBidPoints         int
	MinTeams             int
	MaxTeams             int
	BudgetPoints         int
}

type GenerateResult struct {
	StrategyGenerationRunID string
	RunKey                  string
	NTeams                  int
	TotalBidPoints          int
	SimulatedTournamentID   string
}

type calcuttaContext struct {
	CalcuttaID            string
	CoreTournamentID      string
	BudgetPoints          int
	MinTeams              int
	MaxTeams              int
	MaxBidPoints          int
	NumEntries            int
	SimulatedTournamentID string
}

type ExpectedTeam struct {
	TeamID         string
	ExpectedPoints float64
}

type marketShare struct {
	TeamID         string
	PredictedShare float64
}

type persistedStrategyGenerationParams struct {
	MarketShareRunID  string `json:"market_share_run_id"`
	GameOutcomeRunID  string `json:"game_outcome_run_id"`
	ExcludedEntryName any    `json:"excluded_entry_name"`
	BudgetPoints      int    `json:"budget_points"`
	MinTeams          int    `json:"min_teams"`
	MaxTeams          int    `json:"max_teams"`
	MinBidPoints      int    `json:"min_bid_points"`
	MaxBidPoints      int    `json:"max_bid_points"`
}

func (s *Service) GenerateAndWriteToExistingStrategyGenerationRun(ctx context.Context, strategyGenerationRunID string) (*GenerateResult, error) {
	if strings.TrimSpace(strategyGenerationRunID) == "" {
		return nil, errors.New("strategyGenerationRunID is required")
	}

	var (
		runKeyText      *string
		runKeyUUIDText  *string
		name            *string
		optimizerKey    *string
		purpose         *string
		returnsModelKey *string
		calcuttaID      string
		simulatedID     *string
		marketShareID   *string
		gameOutcomeID   *string
		excludedName    *string
		startingKey     *string
		paramsJSON      []byte
	)
	if err := s.pool.QueryRow(ctx, `
		SELECT
			run_key,
			COALESCE(run_key_uuid::text, ''::text) AS run_key_uuid,
			name,
			optimizer_key,
			purpose,
			returns_model_key,
			calcutta_id::text,
			simulated_tournament_id::text,
			market_share_run_id::text,
			game_outcome_run_id::text,
			excluded_entry_name,
			starting_state_key,
			params_json
		FROM derived.strategy_generation_runs
		WHERE id = $1::uuid
			AND deleted_at IS NULL
		LIMIT 1
	`, strategyGenerationRunID).Scan(
		&runKeyText,
		&runKeyUUIDText,
		&name,
		&optimizerKey,
		&purpose,
		&returnsModelKey,
		&calcuttaID,
		&simulatedID,
		&marketShareID,
		&gameOutcomeID,
		&excludedName,
		&startingKey,
		&paramsJSON,
	); err != nil {
		return nil, err
	}

	pk := persistedStrategyGenerationParams{}
	if len(paramsJSON) > 0 {
		_ = json.Unmarshal(paramsJSON, &pk)
	}

	effRunKey := ""
	if runKeyText != nil && strings.TrimSpace(*runKeyText) != "" {
		effRunKey = strings.TrimSpace(*runKeyText)
	} else if runKeyUUIDText != nil && strings.TrimSpace(*runKeyUUIDText) != "" {
		effRunKey = strings.TrimSpace(*runKeyUUIDText)
	} else {
		effRunKey = uuid.NewString()
	}

	effName := ""
	if name != nil {
		effName = strings.TrimSpace(*name)
	}
	effOptimizerKey := ""
	if optimizerKey != nil {
		effOptimizerKey = strings.TrimSpace(*optimizerKey)
	}

	var marketShareRunID *string
	if marketShareID != nil && strings.TrimSpace(*marketShareID) != "" {
		v := strings.TrimSpace(*marketShareID)
		marketShareRunID = &v
	} else if strings.TrimSpace(pk.MarketShareRunID) != "" {
		v := strings.TrimSpace(pk.MarketShareRunID)
		marketShareRunID = &v
	}

	var gameOutcomeRunID *string
	if gameOutcomeID != nil && strings.TrimSpace(*gameOutcomeID) != "" {
		v := strings.TrimSpace(*gameOutcomeID)
		gameOutcomeRunID = &v
	} else if strings.TrimSpace(pk.GameOutcomeRunID) != "" {
		v := strings.TrimSpace(pk.GameOutcomeRunID)
		gameOutcomeRunID = &v
	}

	var excludedEntryName string
	if excludedName != nil {
		excludedEntryName = strings.TrimSpace(*excludedName)
	}
	if excludedEntryName == "" && pk.ExcludedEntryName != nil {
		switch v := pk.ExcludedEntryName.(type) {
		case string:
			excludedEntryName = strings.TrimSpace(v)
		case *string:
			if v != nil {
				excludedEntryName = strings.TrimSpace(*v)
			}
		}
	}

	var simTournamentID *string
	if simulatedID != nil && strings.TrimSpace(*simulatedID) != "" {
		v := strings.TrimSpace(*simulatedID)
		simTournamentID = &v
	}

	effPurpose := ""
	if purpose != nil {
		effPurpose = strings.TrimSpace(*purpose)
	}
	effReturnsModelKey := ""
	if returnsModelKey != nil {
		effReturnsModelKey = strings.TrimSpace(*returnsModelKey)
	}

	if effPurpose == "lab_entries_generation" && effReturnsModelKey == "pgo_dp" {
		if gameOutcomeRunID == nil || strings.TrimSpace(*gameOutcomeRunID) == "" {
			return nil, errors.New("game_outcome_run_id is required for pgo_dp lab entry generation")
		}

		cc, err := s.loadCalcuttaContext(ctx, calcuttaID)
		if err != nil {
			return nil, err
		}

		msSelectedID, marketByTeam, err := s.loadPredictedMarketShares(ctx, cc, marketShareRunID)
		if err != nil {
			return nil, err
		}
		if msSelectedID == nil || strings.TrimSpace(*msSelectedID) == "" {
			return nil, errors.New("market_share_run_id is required for pgo_dp lab entry generation")
		}

		repo := dbadapter.NewAnalyticsRepository(s.pool)
		_, _, returns, err := repo.GetCalcuttaPredictedReturns(ctx, calcuttaID, &strategyGenerationRunID, gameOutcomeRunID)
		if err != nil {
			return nil, err
		}
		if len(returns) == 0 {
			return nil, errors.New("no predicted returns found")
		}

		expected := make([]ExpectedTeam, 0, len(returns))
		for _, r := range returns {
			expected = append(expected, ExpectedTeam{TeamID: r.TeamID, ExpectedPoints: r.ExpectedValue})
		}

		return s.GenerateFromPredictionsAndWrite(ctx, GenerateFromPredictionsParams{
			CalcuttaID:           calcuttaID,
			RunKey:               effRunKey,
			Name:                 effName,
			OptimizerKey:         effOptimizerKey,
			GameOutcomeRunID:     strings.TrimSpace(*gameOutcomeRunID),
			MarketShareRunID:     strings.TrimSpace(*msSelectedID),
			ExcludedEntryName:    excludedEntryName,
			StartingStateKey:     startingKey,
			ExpectedPointsByTeam: expected,
			PredictedShareByTeam: marketByTeam,
			BudgetPoints:         pk.BudgetPoints,
			MinTeams:             pk.MinTeams,
			MaxTeams:             pk.MaxTeams,
			MinBidPoints:         pk.MinBidPoints,
			MaxBidPoints:         pk.MaxBidPoints,
		})
	}

	return s.GenerateAndWrite(ctx, GenerateParams{
		CalcuttaID:            calcuttaID,
		RunKey:                effRunKey,
		Name:                  effName,
		OptimizerKey:          effOptimizerKey,
		MarketShareRunID:      marketShareRunID,
		SimulatedTournamentID: simTournamentID,
		BudgetPoints:          pk.BudgetPoints,
		MinTeams:              pk.MinTeams,
		MaxTeams:              pk.MaxTeams,
		MinBidPoints:          pk.MinBidPoints,
		MaxBidPoints:          pk.MaxBidPoints,
	})
}

func (s *Service) GenerateAndWrite(ctx context.Context, p GenerateParams) (*GenerateResult, error) {
	if p.CalcuttaID == "" {
		return nil, errors.New("CalcuttaID is required")
	}

	cc, err := s.loadCalcuttaContext(ctx, p.CalcuttaID)
	if err != nil {
		return nil, err
	}

	if p.SimulatedTournamentID != nil && *p.SimulatedTournamentID != "" {
		cc.SimulatedTournamentID = *p.SimulatedTournamentID
	} else {
		simID, ok, err := s.getLatestSimulatedTournamentID(ctx, cc.CoreTournamentID)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, fmt.Errorf("no simulated_tournaments batch found for calcutta_id=%s", p.CalcuttaID)
		}
		cc.SimulatedTournamentID = simID
	}

	expected, _, err := s.loadExpectedPointsByTeam(ctx, cc)
	if err != nil {
		return nil, err
	}
	if len(expected) == 0 {
		return nil, errors.New("no expected points found")
	}

	selectedMarketShareRunID, marketByTeam, err := s.loadPredictedMarketShares(ctx, cc, p.MarketShareRunID)
	if err != nil {
		return nil, err
	}

	poolSize := float64(cc.NumEntries * cc.BudgetPoints)
	teams := make([]Team, 0, len(expected))
	for _, t := range expected {
		share, ok := marketByTeam[t.TeamID]
		if !ok {
			rid := ""
			if selectedMarketShareRunID != nil {
				rid = *selectedMarketShareRunID
			}
			return nil, fmt.Errorf("missing predicted_market_share for team_id=%s (calcutta_id=%s market_share_run_id=%s)", t.TeamID, cc.CalcuttaID, rid)
		}
		marketPoints := share * poolSize
		teams = append(teams, Team{ID: t.TeamID, ExpectedPoints: t.ExpectedPoints, MarketPoints: marketPoints})
	}

	minBid := p.MinBidPoints
	if minBid <= 0 {
		minBid = 1
	}
	maxBid := p.MaxBidPoints
	if maxBid <= 0 {
		maxBid = cc.MaxBidPoints
	}

	minTeams := p.MinTeams
	if minTeams <= 0 {
		minTeams = cc.MinTeams
	}
	maxTeams := p.MaxTeams
	if maxTeams <= 0 {
		maxTeams = cc.MaxTeams
	}

	budget := p.BudgetPoints
	if budget <= 0 {
		budget = cc.BudgetPoints
	}

	optimizerKey := p.OptimizerKey
	if optimizerKey == "" {
		optimizerKey = "minlp_v1"
	}
	if optimizerKey != "minlp_v1" {
		return nil, fmt.Errorf("unsupported optimizer %q (supported: minlp_v1)", optimizerKey)
	}

	alloc, err := AllocateBids(teams, AllocationParams{
		BudgetPoints: budget,
		MinTeams:     minTeams,
		MaxTeams:     maxTeams,
		MinBidPoints: minBid,
		MaxBidPoints: maxBid,
	})
	if err != nil {
		return nil, err
	}

	totalBid := 0
	for _, v := range alloc.Bids {
		totalBid += v
	}

	runKey := p.RunKey
	if runKey == "" {
		runKey = uuid.NewString()
	}
	name := p.Name
	if name == "" {
		name = optimizerKey
	}

	paramsJSON, _ := json.Marshal(map[string]any{
		"budget_points":    budget,
		"min_teams":        minTeams,
		"max_teams":        maxTeams,
		"min_bid_points":   minBid,
		"max_bid_points":   maxBid,
		"pool_size_points": poolSize,
	})

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	simID := cc.SimulatedTournamentID
	strategyRunID, err := upsertStrategyGenerationRun(ctx, tx, upsertStrategyGenerationRunParams{
		RunKey:                runKey,
		Name:                  name,
		SimulatedTournamentID: &simID,
		CalcuttaID:            cc.CalcuttaID,
		Purpose:               "go_recommended_entry_bids",
		ReturnsModelKey:       "legacy",
		InvestmentModelKey:    "predicted_market_share",
		OptimizerKey:          optimizerKey,
		MarketShareRunID:      selectedMarketShareRunID,
		ParamsJSON:            paramsJSON,
	})
	if err != nil {
		return nil, err
	}

	err = writeRecommendedEntryBids(ctx, tx, writeRecommendedEntryBidsParams{
		RunID:                   runKey,
		StrategyGenerationRunID: strategyRunID,
		Bids:                    alloc.Bids,
		ExpectedPointsByTeam:    toExpectedTeams(expected),
		MarketPointsByTeam:      mapTeamMarketPoints(teams),
	})
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &GenerateResult{
		StrategyGenerationRunID: strategyRunID,
		RunKey:                  runKey,
		NTeams:                  len(alloc.Bids),
		TotalBidPoints:          totalBid,
		SimulatedTournamentID:   cc.SimulatedTournamentID,
	}, nil
}

func (s *Service) GenerateFromPredictionsAndWrite(ctx context.Context, p GenerateFromPredictionsParams) (*GenerateResult, error) {
	if p.CalcuttaID == "" {
		return nil, errors.New("CalcuttaID is required")
	}
	if p.GameOutcomeRunID == "" {
		return nil, errors.New("GameOutcomeRunID is required")
	}
	if p.MarketShareRunID == "" {
		return nil, errors.New("MarketShareRunID is required")
	}
	if len(p.ExpectedPointsByTeam) == 0 {
		return nil, errors.New("ExpectedPointsByTeam is required")
	}
	if len(p.PredictedShareByTeam) == 0 {
		return nil, errors.New("PredictedShareByTeam is required")
	}

	cc, err := s.loadCalcuttaContext(ctx, p.CalcuttaID)
	if err != nil {
		return nil, err
	}

	minBid := p.MinBidPoints
	if minBid <= 0 {
		minBid = 1
	}
	maxBid := p.MaxBidPoints
	if maxBid <= 0 {
		maxBid = cc.MaxBidPoints
	}

	minTeams := p.MinTeams
	if minTeams <= 0 {
		minTeams = cc.MinTeams
	}
	maxTeams := p.MaxTeams
	if maxTeams <= 0 {
		maxTeams = cc.MaxTeams
	}

	budget := p.BudgetPoints
	if budget <= 0 {
		budget = cc.BudgetPoints
	}

	optimizerKey := p.OptimizerKey
	if optimizerKey == "" {
		optimizerKey = "minlp_v1"
	}
	if optimizerKey != "minlp_v1" {
		return nil, fmt.Errorf("unsupported optimizer %q (supported: minlp_v1)", optimizerKey)
	}

	poolSize := float64(cc.NumEntries * cc.BudgetPoints)
	teams := make([]Team, 0, len(p.ExpectedPointsByTeam))
	for _, t := range p.ExpectedPointsByTeam {
		share, ok := p.PredictedShareByTeam[t.TeamID]
		if !ok {
			return nil, fmt.Errorf("missing predicted_market_share for team_id=%s (calcutta_id=%s market_share_run_id=%s)", t.TeamID, cc.CalcuttaID, p.MarketShareRunID)
		}
		marketPoints := share * poolSize
		teams = append(teams, Team{ID: t.TeamID, ExpectedPoints: t.ExpectedPoints, MarketPoints: marketPoints})
	}

	alloc, err := AllocateBids(teams, AllocationParams{
		BudgetPoints: budget,
		MinTeams:     minTeams,
		MaxTeams:     maxTeams,
		MinBidPoints: minBid,
		MaxBidPoints: maxBid,
	})
	if err != nil {
		return nil, err
	}

	totalBid := 0
	for _, v := range alloc.Bids {
		totalBid += v
	}

	runKey := p.RunKey
	if runKey == "" {
		runKey = uuid.NewString()
	}
	name := p.Name
	if name == "" {
		name = optimizerKey
	}

	paramsJSON, _ := json.Marshal(map[string]any{
		"budget_points":       budget,
		"min_teams":           minTeams,
		"max_teams":           maxTeams,
		"min_bid_points":      minBid,
		"max_bid_points":      maxBid,
		"pool_size_points":    poolSize,
		"assumed_entries":     cc.NumEntries,
		"excluded_entry_name": p.ExcludedEntryName,
		"game_outcome_run_id": p.GameOutcomeRunID,
		"market_share_run_id": p.MarketShareRunID,
		"starting_state_key":  p.StartingStateKey,
	})

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var excludedEntryNameParam *string
	if strings.TrimSpace(p.ExcludedEntryName) != "" {
		excludedEntryNameParam = &p.ExcludedEntryName
	}

	strategyRunID, err := upsertStrategyGenerationRun(ctx, tx, upsertStrategyGenerationRunParams{
		RunKey:                runKey,
		Name:                  name,
		SimulatedTournamentID: nil,
		CalcuttaID:            cc.CalcuttaID,
		Purpose:               "lab_entries_generation",
		ReturnsModelKey:       "pgo_dp",
		InvestmentModelKey:    "predicted_market_share",
		OptimizerKey:          optimizerKey,
		MarketShareRunID:      &p.MarketShareRunID,
		GameOutcomeRunID:      &p.GameOutcomeRunID,
		ExcludedEntryName:     excludedEntryNameParam,
		StartingStateKey:      p.StartingStateKey,
		ParamsJSON:            paramsJSON,
	})
	if err != nil {
		return nil, err
	}

	err = writeRecommendedEntryBids(ctx, tx, writeRecommendedEntryBidsParams{
		RunID:                   runKey,
		StrategyGenerationRunID: strategyRunID,
		Bids:                    alloc.Bids,
		ExpectedPointsByTeam:    p.ExpectedPointsByTeam,
		MarketPointsByTeam:      mapTeamMarketPoints(teams),
	})
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &GenerateResult{
		StrategyGenerationRunID: strategyRunID,
		RunKey:                  runKey,
		NTeams:                  len(alloc.Bids),
		TotalBidPoints:          totalBid,
		SimulatedTournamentID:   "",
	}, nil
}

func (s *Service) loadCalcuttaContext(ctx context.Context, calcuttaID string) (*calcuttaContext, error) {
	query := `
		SELECT
			c.id,
			c.tournament_id,
			c.budget_points,
			c.min_teams,
			c.max_teams,
			c.max_bid,
			(
				SELECT COUNT(*)::int
				FROM core.entries ce
				WHERE ce.calcutta_id = c.id
					AND ce.deleted_at IS NULL
			)
		FROM core.calcuttas c
		JOIN core.tournaments t ON t.id = c.tournament_id AND t.deleted_at IS NULL
		WHERE c.id = $1::uuid
			AND c.deleted_at IS NULL
		LIMIT 1
	`

	cc := &calcuttaContext{}
	if err := s.pool.QueryRow(ctx, query, calcuttaID).Scan(
		&cc.CalcuttaID,
		&cc.CoreTournamentID,
		&cc.BudgetPoints,
		&cc.MinTeams,
		&cc.MaxTeams,
		&cc.MaxBidPoints,
		&cc.NumEntries,
	); err != nil {
		return nil, err
	}
	if cc.NumEntries <= 0 {
		cc.NumEntries = 47
	}
	if cc.BudgetPoints <= 0 {
		cc.BudgetPoints = 100
	}
	return cc, nil
}

func (s *Service) getLatestSimulatedTournamentID(ctx context.Context, coreTournamentID string) (string, bool, error) {
	var batchID string
	q := `
		SELECT b.id
		FROM derived.simulated_tournaments b
		WHERE b.tournament_id = $1
			AND b.deleted_at IS NULL
			AND EXISTS (
				SELECT 1
				FROM derived.simulated_teams st
				WHERE st.tournament_id = $1
					AND st.simulated_tournament_id = b.id
					AND st.deleted_at IS NULL
			)
		ORDER BY b.created_at DESC
		LIMIT 1
	`
	err := s.pool.QueryRow(ctx, q, coreTournamentID).Scan(&batchID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", false, nil
		}
		return "", false, err
	}
	return batchID, true, nil
}

func (s *Service) loadExpectedPointsByTeam(ctx context.Context, cc *calcuttaContext) ([]ExpectedTeam, float64, error) {
	q := `
		SELECT
			st.team_id,
			AVG(core.calcutta_points_for_progress($3::uuid, st.wins + 1, st.byes))::float AS expected_points
		FROM derived.simulated_teams st
		WHERE st.tournament_id = $1
			AND st.simulated_tournament_id = $2::uuid
			AND st.deleted_at IS NULL
		GROUP BY st.team_id
	`

	rows, err := s.pool.Query(ctx, q, cc.CoreTournamentID, cc.SimulatedTournamentID, cc.CalcuttaID)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]ExpectedTeam, 0)
	total := 0.0
	for rows.Next() {
		var teamID string
		var expected float64
		if err := rows.Scan(&teamID, &expected); err != nil {
			return nil, 0, err
		}
		out = append(out, ExpectedTeam{TeamID: teamID, ExpectedPoints: expected})
		total += expected
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

func (s *Service) loadPredictedMarketShares(ctx context.Context, cc *calcuttaContext, marketShareRunID *string) (*string, map[string]float64, error) {
	if cc == nil {
		return nil, nil, errors.New("calcutta context is required")
	}
	if cc.CalcuttaID == "" {
		return nil, nil, errors.New("calcutta_id is required")
	}

	selected := (*string)(nil)
	if marketShareRunID != nil && *marketShareRunID != "" {
		selected = marketShareRunID
	} else {
		var latestRunID string
		if err := s.pool.QueryRow(ctx, `
			SELECT id::text
			FROM derived.market_share_runs
			WHERE calcutta_id = $1::uuid
				AND deleted_at IS NULL
			ORDER BY created_at DESC
			LIMIT 1
		`, cc.CalcuttaID).Scan(&latestRunID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, nil, fmt.Errorf("no market_share_runs found for calcutta_id=%s", cc.CalcuttaID)
			}
			return nil, nil, err
		}
		selected = &latestRunID
	}

	q := `
		SELECT team_id, predicted_share
		FROM derived.predicted_market_share
		WHERE run_id = $1::uuid
			AND deleted_at IS NULL
	`
	rows, err := s.pool.Query(ctx, q, *selected)
	if err != nil {
		return selected, nil, err
	}
	defer rows.Close()

	out := make(map[string]float64)
	for rows.Next() {
		var teamID string
		var share float64
		if err := rows.Scan(&teamID, &share); err != nil {
			return selected, nil, err
		}
		out[teamID] = share
	}
	if err := rows.Err(); err != nil {
		return selected, nil, err
	}
	return selected, out, nil
}

type upsertStrategyGenerationRunParams struct {
	RunKey                string
	Name                  string
	SimulatedTournamentID *string
	CalcuttaID            string
	Purpose               string
	ReturnsModelKey       string
	InvestmentModelKey    string
	OptimizerKey          string
	MarketShareRunID      *string
	GameOutcomeRunID      *string
	ExcludedEntryName     *string
	StartingStateKey      *string
	ParamsJSON            []byte
}

func upsertStrategyGenerationRun(ctx context.Context, tx pgx.Tx, p upsertStrategyGenerationRunParams) (string, error) {
	var id string
	q := `
		INSERT INTO derived.strategy_generation_runs (
			run_key,
			name,
			simulated_tournament_id,
			calcutta_id,
			purpose,
			returns_model_key,
			investment_model_key,
			optimizer_key,
			market_share_run_id,
			game_outcome_run_id,
			excluded_entry_name,
			starting_state_key,
			params_json,
			git_sha,
			created_at,
			updated_at,
			deleted_at
		)
		VALUES ($1, $2, $3::uuid, $4::uuid, $5, $6, $7, $8, $9::uuid, $10::uuid, $11::text, $12::text, $13::jsonb, NULL, NOW(), NOW(), NULL)
		ON CONFLICT (run_key) DO UPDATE SET
			name = EXCLUDED.name,
			simulated_tournament_id = EXCLUDED.simulated_tournament_id,
			calcutta_id = EXCLUDED.calcutta_id,
			purpose = EXCLUDED.purpose,
			returns_model_key = EXCLUDED.returns_model_key,
			investment_model_key = EXCLUDED.investment_model_key,
			optimizer_key = EXCLUDED.optimizer_key,
			market_share_run_id = EXCLUDED.market_share_run_id,
			game_outcome_run_id = EXCLUDED.game_outcome_run_id,
			excluded_entry_name = EXCLUDED.excluded_entry_name,
			starting_state_key = EXCLUDED.starting_state_key,
			params_json = EXCLUDED.params_json,
			updated_at = NOW(),
			deleted_at = NULL
		RETURNING id
	`
	err := tx.QueryRow(ctx, q, p.RunKey, p.Name, p.SimulatedTournamentID, p.CalcuttaID, p.Purpose, p.ReturnsModelKey, p.InvestmentModelKey, p.OptimizerKey, p.MarketShareRunID, p.GameOutcomeRunID, p.ExcludedEntryName, p.StartingStateKey, p.ParamsJSON).Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

type writeRecommendedEntryBidsParams struct {
	RunID                   string
	StrategyGenerationRunID string
	Bids                    map[string]int
	ExpectedPointsByTeam    []ExpectedTeam
	MarketPointsByTeam      map[string]float64
}

func toExpectedTeams(in []ExpectedTeam) []ExpectedTeam { return in }

func mapTeamMarketPoints(teams []Team) map[string]float64 {
	out := make(map[string]float64, len(teams))
	for _, t := range teams {
		out[t.ID] = t.MarketPoints
	}
	return out
}

func writeRecommendedEntryBids(ctx context.Context, tx pgx.Tx, p writeRecommendedEntryBidsParams) error {
	_, err := tx.Exec(ctx, `
		DELETE FROM derived.recommended_entry_bids
		WHERE strategy_generation_run_id = $1::uuid
	`, p.StrategyGenerationRunID)
	if err != nil {
		return err
	}

	expByTeam := make(map[string]float64, len(p.ExpectedPointsByTeam))
	for _, t := range p.ExpectedPointsByTeam {
		expByTeam[t.TeamID] = t.ExpectedPoints
	}

	rows := make([][]any, 0, len(p.Bids))
	for teamID, bid := range p.Bids {
		if bid <= 0 {
			continue
		}
		expected := expByTeam[teamID]
		expectedROI := 0.0
		if bid > 0 {
			expectedROI = expected / float64(bid)
		}
		rows = append(rows, []any{p.RunID, p.StrategyGenerationRunID, teamID, bid, expectedROI, time.Now()})
	}

	copyFrom, err := tx.CopyFrom(
		ctx,
		pgx.Identifier{"derived", "recommended_entry_bids"},
		[]string{"run_id", "strategy_generation_run_id", "team_id", "bid_points", "expected_roi", "created_at"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return err
	}
	_ = copyFrom
	return nil
}
