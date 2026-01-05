package db

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/adapters/db/sqlc"
	"github.com/andrewcopp/Calcutta/backend/internal/ports"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func asFloat64(v any) (float64, error) {
	switch t := v.(type) {
	case nil:
		return 0, nil
	case float64:
		return t, nil
	case float32:
		return float64(t), nil
	case int:
		return float64(t), nil
	case int32:
		return float64(t), nil
	case int64:
		return float64(t), nil
	case uint:
		return float64(t), nil
	case uint32:
		return float64(t), nil
	case uint64:
		return float64(t), nil
	case string:
		f, err := strconv.ParseFloat(t, 64)
		if err != nil {
			return 0, err
		}
		return f, nil
	default:
		return 0, fmt.Errorf("unexpected numeric type %T", v)
	}
}

type AnalyticsRepository struct {
	pool *pgxpool.Pool
	q    *sqlc.Queries
}

func NewAnalyticsRepository(pool *pgxpool.Pool) *AnalyticsRepository {
	return &AnalyticsRepository{pool: pool, q: sqlc.New(pool)}
}

func (r *AnalyticsRepository) resolveStrategyGenerationRunID(ctx context.Context, calcuttaID string, strategyGenerationRunID *string) (*string, error) {
	if strategyGenerationRunID != nil && *strategyGenerationRunID != "" {
		runID := *strategyGenerationRunID
		return &runID, nil
	}

	latestID, err := r.q.GetLatestStrategyGenerationRunIDByCoreCalcuttaID(ctx, calcuttaID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &latestID, nil
}

func (r *AnalyticsRepository) GetSeedAnalytics(ctx context.Context) ([]ports.SeedAnalyticsData, float64, float64, error) {
	rows, err := r.q.GetSeedAnalytics(ctx)
	if err != nil {
		return nil, 0, 0, err
	}

	out := make([]ports.SeedAnalyticsData, 0, len(rows))
	var totalPoints float64
	var totalInvestment float64
	for _, row := range rows {
		out = append(out, ports.SeedAnalyticsData{
			Seed:            int(row.Seed),
			TotalPoints:     row.TotalPoints,
			TotalInvestment: row.TotalInvestment,
			TeamCount:       int(row.TeamCount),
		})
		totalPoints += row.TotalPoints
		totalInvestment += row.TotalInvestment
	}

	return out, totalPoints, totalInvestment, nil
}

func (r *AnalyticsRepository) GetRegionAnalytics(ctx context.Context) ([]ports.RegionAnalyticsData, float64, float64, error) {
	rows, err := r.q.GetRegionAnalytics(ctx)
	if err != nil {
		return nil, 0, 0, err
	}

	out := make([]ports.RegionAnalyticsData, 0, len(rows))
	var totalPoints float64
	var totalInvestment float64
	for _, row := range rows {
		out = append(out, ports.RegionAnalyticsData{
			Region:          row.Region,
			TotalPoints:     row.TotalPoints,
			TotalInvestment: row.TotalInvestment,
			TeamCount:       int(row.TeamCount),
		})
		totalPoints += row.TotalPoints
		totalInvestment += row.TotalInvestment
	}

	return out, totalPoints, totalInvestment, nil
}

func (r *AnalyticsRepository) GetTeamAnalytics(ctx context.Context) ([]ports.TeamAnalyticsData, error) {
	rows, err := r.q.GetTeamAnalytics(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]ports.TeamAnalyticsData, 0, len(rows))
	for _, row := range rows {
		out = append(out, ports.TeamAnalyticsData{
			SchoolID:        row.SchoolID,
			SchoolName:      row.SchoolName,
			TotalPoints:     row.TotalPoints,
			TotalInvestment: row.TotalInvestment,
			Appearances:     int(row.Appearances),
			TotalSeed:       int(row.TotalSeed),
		})
	}

	return out, nil
}

func (r *AnalyticsRepository) ListAlgorithms(ctx context.Context, kind *string) ([]ports.Algorithm, error) {
	var rows pgx.Rows
	var err error
	if kind != nil && *kind != "" {
		rows, err = r.pool.Query(ctx, `
			SELECT id::text, kind, name, description, params_json::bytea, created_at
			FROM derived.algorithms
			WHERE kind = $1::text
				AND deleted_at IS NULL
			ORDER BY created_at DESC
		`, *kind)
	} else {
		rows, err = r.pool.Query(ctx, `
			SELECT id::text, kind, name, description, params_json::bytea, created_at
			FROM derived.algorithms
			WHERE deleted_at IS NULL
			ORDER BY created_at DESC
		`)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]ports.Algorithm, 0)
	for rows.Next() {
		var id, k, name string
		var desc *string
		var params []byte
		var createdAt time.Time
		if err := rows.Scan(&id, &k, &name, &desc, &params, &createdAt); err != nil {
			return nil, err
		}
		out = append(out, ports.Algorithm{
			ID:          id,
			Kind:        k,
			Name:        name,
			Description: desc,
			ParamsJSON:  params,
			CreatedAt:   createdAt,
		})
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return out, nil
}

func (r *AnalyticsRepository) ListGameOutcomeRunsByTournamentID(ctx context.Context, tournamentID string) ([]ports.GameOutcomeRun, error) {
	if tournamentID == "" {
		return nil, errors.New("tournamentID is required")
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id::text, algorithm_id::text, tournament_id::text, params_json::bytea, git_sha, created_at
		FROM derived.game_outcome_runs
		WHERE tournament_id = $1::uuid
			AND deleted_at IS NULL
		ORDER BY created_at DESC
	`, tournamentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]ports.GameOutcomeRun, 0)
	for rows.Next() {
		var run ports.GameOutcomeRun
		if err := rows.Scan(&run.ID, &run.AlgorithmID, &run.TournamentID, &run.ParamsJSON, &run.GitSHA, &run.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, run)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return out, nil
}

func (r *AnalyticsRepository) ListMarketShareRunsByCalcuttaID(ctx context.Context, calcuttaID string) ([]ports.MarketShareRun, error) {
	if calcuttaID == "" {
		return nil, errors.New("calcuttaID is required")
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id::text, algorithm_id::text, calcutta_id::text, params_json::bytea, git_sha, created_at
		FROM derived.market_share_runs
		WHERE calcutta_id = $1::uuid
			AND deleted_at IS NULL
		ORDER BY created_at DESC
	`, calcuttaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]ports.MarketShareRun, 0)
	for rows.Next() {
		var run ports.MarketShareRun
		if err := rows.Scan(&run.ID, &run.AlgorithmID, &run.CalcuttaID, &run.ParamsJSON, &run.GitSHA, &run.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, run)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return out, nil
}

func (r *AnalyticsRepository) GetLatestPredictionRunsForCalcutta(ctx context.Context, calcuttaID string) (*ports.LatestPredictionRuns, error) {
	if calcuttaID == "" {
		return nil, errors.New("calcuttaID is required")
	}

	var tournamentID string
	var gameOutcomeRunID string
	var marketShareRunID string
	if err := r.pool.QueryRow(ctx, `
		SELECT
			c.tournament_id::text AS tournament_id,
			COALESCE((
				SELECT gor.id::text
				FROM derived.game_outcome_runs gor
				WHERE gor.tournament_id = c.tournament_id
					AND gor.deleted_at IS NULL
				ORDER BY gor.created_at DESC
				LIMIT 1
			), ''::text) AS game_outcome_run_id,
			COALESCE((
				SELECT msr.id::text
				FROM derived.market_share_runs msr
				WHERE msr.calcutta_id = c.id
					AND msr.deleted_at IS NULL
				ORDER BY msr.created_at DESC
				LIMIT 1
			), ''::text) AS market_share_run_id
		FROM core.calcuttas c
		WHERE c.id = $1::uuid
			AND c.deleted_at IS NULL
		LIMIT 1
	`, calcuttaID).Scan(&tournamentID, &gameOutcomeRunID, &marketShareRunID); err != nil {
		return nil, err
	}

	var goPtr *string
	if gameOutcomeRunID != "" {
		v := gameOutcomeRunID
		goPtr = &v
	}
	var msPtr *string
	if marketShareRunID != "" {
		v := marketShareRunID
		msPtr = &v
	}

	return &ports.LatestPredictionRuns{
		TournamentID:     tournamentID,
		GameOutcomeRunID: goPtr,
		MarketShareRunID: msPtr,
	}, nil
}

func (r *AnalyticsRepository) GetCalcuttaPredictedInvestment(ctx context.Context, calcuttaID string, strategyGenerationRunID *string, marketShareRunID *string) (*string, *string, []ports.CalcuttaPredictedInvestmentData, error) {
	runIDPtr, err := r.resolveStrategyGenerationRunID(ctx, calcuttaID, strategyGenerationRunID)
	if err != nil {
		return nil, nil, nil, err
	}

	marketShareSelectedID, out, err := computeCalcuttaPredictedInvestmentFromPGO(ctx, r.pool, calcuttaID, marketShareRunID)
	if err != nil {
		return nil, nil, nil, err
	}

	return runIDPtr, marketShareSelectedID, out, nil
}

func (r *AnalyticsRepository) GetCalcuttaPredictedReturns(ctx context.Context, calcuttaID string, strategyGenerationRunID *string, gameOutcomeRunID *string) (*string, *string, []ports.CalcuttaPredictedReturnsData, error) {
	runIDPtr, err := r.resolveStrategyGenerationRunID(ctx, calcuttaID, strategyGenerationRunID)
	if err != nil {
		return nil, nil, nil, err
	}

	gameOutcomeSelectedID, out, err := computeCalcuttaPredictedReturnsFromPGO(ctx, r.pool, calcuttaID, gameOutcomeRunID)
	if err != nil {
		return nil, nil, nil, err
	}

	return runIDPtr, gameOutcomeSelectedID, out, nil
}

func (r *AnalyticsRepository) GetTournamentPredictedAdvancement(ctx context.Context, tournamentID string, gameOutcomeRunID *string) (*string, []ports.TournamentPredictedAdvancementData, error) {
	selectedID, out, err := computeTournamentPredictedAdvancementFromPGO(ctx, r.pool, tournamentID, gameOutcomeRunID)
	if err != nil {
		return nil, nil, err
	}
	return selectedID, out, nil
}

func (r *AnalyticsRepository) GetCalcuttaPredictedMarketShare(ctx context.Context, calcuttaID string, marketShareRunID *string, gameOutcomeRunID *string) (*string, *string, []ports.CalcuttaPredictedMarketShareData, error) {
	marketShareSelectedID, gameOutcomeSelectedID, out, err := computeCalcuttaPredictedMarketShareFromPGO(ctx, r.pool, calcuttaID, marketShareRunID, gameOutcomeRunID)
	if err != nil {
		return nil, nil, nil, err
	}
	return marketShareSelectedID, gameOutcomeSelectedID, out, nil
}

func (r *AnalyticsRepository) GetCalcuttaSimulatedEntry(ctx context.Context, calcuttaID string, strategyGenerationRunID *string) (*string, []ports.CalcuttaSimulatedEntryData, error) {
	runIDPtr, err := r.resolveStrategyGenerationRunID(ctx, calcuttaID, strategyGenerationRunID)
	if err != nil {
		return nil, nil, err
	}
	if runIDPtr != nil {
		runID := *runIDPtr
		rows, err := r.q.GetCalcuttaSimulatedEntryByStrategyGenerationRunID(ctx, sqlc.GetCalcuttaSimulatedEntryByStrategyGenerationRunIDParams{
			StrategyGenerationRunID: runID,
			CalcuttaID:              calcuttaID,
		})
		if err != nil {
			return nil, nil, err
		}

		out := make([]ports.CalcuttaSimulatedEntryData, 0, len(rows))
		for _, row := range rows {
			out = append(out, ports.CalcuttaSimulatedEntryData{
				TeamID:         row.TeamID,
				SchoolName:     row.SchoolName,
				Seed:           int(row.Seed),
				Region:         row.Region,
				ExpectedPoints: row.ExpectedPoints,
				ExpectedMarket: row.ExpectedMarket,
				OurBid:         row.OurBid,
			})
		}

		return runIDPtr, out, nil
	}

	// Legacy fallback.
	rows, err := r.q.GetCalcuttaSimulatedEntry(ctx, calcuttaID)
	if err != nil {
		return nil, nil, err
	}

	out := make([]ports.CalcuttaSimulatedEntryData, 0, len(rows))
	for _, row := range rows {
		out = append(out, ports.CalcuttaSimulatedEntryData{
			TeamID:         row.TeamID,
			SchoolName:     row.SchoolName,
			Seed:           int(row.Seed),
			Region:         row.Region,
			ExpectedPoints: row.ExpectedPoints,
			ExpectedMarket: row.ExpectedMarket,
			OurBid:         row.OurBid,
		})
	}

	return nil, out, nil
}

func (r *AnalyticsRepository) GetSeedVarianceAnalytics(ctx context.Context) ([]ports.SeedVarianceData, error) {
	rows, err := r.q.GetSeedVarianceAnalytics(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]ports.SeedVarianceData, 0, len(rows))
	for _, row := range rows {
		out = append(out, ports.SeedVarianceData{
			Seed:             int(row.Seed),
			InvestmentStdDev: row.InvestmentStddev,
			PointsStdDev:     row.PointsStddev,
			InvestmentMean:   row.InvestmentMean,
			PointsMean:       row.PointsMean,
			TeamCount:        int(row.TeamCount),
		})
	}
	return out, nil
}

func (r *AnalyticsRepository) GetSeedInvestmentPoints(ctx context.Context) ([]ports.SeedInvestmentPointData, error) {
	rows, err := r.q.GetSeedInvestmentPoints(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]ports.SeedInvestmentPointData, 0, len(rows))
	for _, row := range rows {
		out = append(out, ports.SeedInvestmentPointData{
			Seed:             int(row.Seed),
			TournamentName:   row.TournamentName,
			TournamentYear:   int(row.TournamentYear),
			CalcuttaID:       row.CalcuttaID,
			TeamID:           row.TeamID,
			SchoolName:       row.SchoolName,
			TotalBid:         row.TotalBid,
			CalcuttaTotalBid: row.CalcuttaTotalBid,
			NormalizedBid:    row.NormalizedBid,
		})
	}

	return out, nil
}

func (r *AnalyticsRepository) GetBestInvestments(ctx context.Context, limit int) ([]ports.BestInvestmentData, error) {
	rows, err := r.q.GetBestInvestments(ctx, int32(limit))
	if err != nil {
		return nil, err
	}

	out := make([]ports.BestInvestmentData, 0, len(rows))
	for _, row := range rows {
		out = append(out, ports.BestInvestmentData{
			TournamentName:   row.TournamentName,
			TournamentYear:   int(row.TournamentYear),
			CalcuttaID:       row.CalcuttaID,
			TeamID:           row.TeamID,
			SchoolName:       row.SchoolName,
			Seed:             int(row.Seed),
			Region:           row.Region,
			TeamPoints:       row.TeamPoints,
			TotalBid:         row.TotalBid,
			CalcuttaTotalBid: row.CalcuttaTotalBid,
			CalcuttaTotalPts: row.CalcuttaTotalPoints,
			InvestmentShare:  row.InvestmentShare,
			PointsShare:      row.PointsShare,
			RawROI:           row.RawRoi,
			NormalizedROI:    row.NormalizedRoi,
		})
	}

	return out, nil
}

func (r *AnalyticsRepository) GetBestInvestmentBids(ctx context.Context, limit int) ([]ports.InvestmentLeaderboardData, error) {
	rows, err := r.q.GetBestInvestmentBids(ctx, int32(limit))
	if err != nil {
		return nil, err
	}

	out := make([]ports.InvestmentLeaderboardData, 0, len(rows))
	for _, row := range rows {
		out = append(out, ports.InvestmentLeaderboardData{
			TournamentName:      row.TournamentName,
			TournamentYear:      int(row.TournamentYear),
			CalcuttaID:          row.CalcuttaID,
			EntryID:             row.EntryID,
			EntryName:           row.EntryName,
			TeamID:              row.TeamID,
			SchoolName:          row.SchoolName,
			Seed:                int(row.Seed),
			Investment:          row.Investment,
			OwnershipPercentage: row.OwnershipPercentage,
			RawReturns:          row.RawReturns,
			NormalizedReturns:   row.NormalizedReturns,
		})
	}

	return out, nil
}

func (r *AnalyticsRepository) GetBestEntries(ctx context.Context, limit int) ([]ports.EntryLeaderboardData, error) {
	rows, err := r.q.GetBestEntries(ctx, int32(limit))
	if err != nil {
		return nil, err
	}

	out := make([]ports.EntryLeaderboardData, 0, len(rows))
	for _, row := range rows {
		out = append(out, ports.EntryLeaderboardData{
			TournamentName:    row.TournamentName,
			TournamentYear:    int(row.TournamentYear),
			CalcuttaID:        row.CalcuttaID,
			EntryID:           row.EntryID,
			EntryName:         row.EntryName,
			TotalReturns:      row.TotalReturns,
			TotalParticipants: int(row.TotalParticipants),
			AverageReturns:    row.AverageReturns,
			NormalizedReturns: row.NormalizedReturns,
		})
	}

	return out, nil
}

func (r *AnalyticsRepository) GetBestCareers(ctx context.Context, limit int) ([]ports.CareerLeaderboardData, error) {
	rows, err := r.q.GetBestCareers(ctx, int32(limit))
	if err != nil {
		return nil, err
	}

	out := make([]ports.CareerLeaderboardData, 0, len(rows))
	for _, row := range rows {
		out = append(out, ports.CareerLeaderboardData{
			EntryName:              row.EntryName,
			Years:                  int(row.Years),
			BestFinish:             int(row.BestFinish),
			Wins:                   int(row.Wins),
			Podiums:                int(row.Podiums),
			InTheMoneys:            int(row.InTheMoneys),
			Top10s:                 int(row.Top10s),
			CareerEarningsCents:    int(row.CareerEarningsCents),
			ActiveInLatestCalcutta: row.ActiveInLatestCalcutta,
		})
	}

	return out, nil
}
