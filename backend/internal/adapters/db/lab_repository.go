package db

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/andrewcopp/Calcutta/backend/internal/app/apperrors"
	"github.com/andrewcopp/Calcutta/backend/internal/app/lab"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// LabRepository provides database access for lab.* tables.
type LabRepository struct {
	pool *pgxpool.Pool
}

// NewLabRepository creates a new lab repository.
func NewLabRepository(pool *pgxpool.Pool) *LabRepository {
	return &LabRepository{pool: pool}
}

// ListInvestmentModels returns investment models matching the filter.
func (r *LabRepository) ListInvestmentModels(filter lab.ListModelsFilter, page lab.Pagination) ([]lab.InvestmentModel, error) {
	ctx := context.Background()

	query := `
		SELECT
			im.id::text,
			im.name,
			im.kind,
			im.params_json::text,
			im.notes,
			im.created_at,
			im.updated_at,
			(SELECT COUNT(*) FROM lab.entries e WHERE e.investment_model_id = im.id AND e.deleted_at IS NULL)::int AS n_entries,
			(SELECT COUNT(*) FROM lab.evaluations ev JOIN lab.entries e ON e.id = ev.entry_id WHERE e.investment_model_id = im.id AND ev.deleted_at IS NULL AND e.deleted_at IS NULL)::int AS n_evaluations
		FROM lab.investment_models im
		WHERE im.deleted_at IS NULL
	`
	args := []any{}
	argIdx := 1

	if filter.Kind != nil && *filter.Kind != "" {
		query += ` AND im.kind = $` + labItoa(argIdx)
		args = append(args, *filter.Kind)
		argIdx++
	}

	query += ` ORDER BY im.created_at DESC`
	query += ` LIMIT $` + labItoa(argIdx) + ` OFFSET $` + labItoa(argIdx+1)
	args = append(args, page.Limit, page.Offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]lab.InvestmentModel, 0)
	for rows.Next() {
		var m lab.InvestmentModel
		var paramsStr string
		if err := rows.Scan(&m.ID, &m.Name, &m.Kind, &paramsStr, &m.Notes, &m.CreatedAt, &m.UpdatedAt, &m.NEntries, &m.NEvaluations); err != nil {
			return nil, err
		}
		m.ParamsJSON = json.RawMessage(paramsStr)
		out = append(out, m)
	}
	return out, rows.Err()
}

// GetInvestmentModel returns a single investment model by ID.
func (r *LabRepository) GetInvestmentModel(id string) (*lab.InvestmentModel, error) {
	ctx := context.Background()

	query := `
		SELECT
			im.id::text,
			im.name,
			im.kind,
			im.params_json::text,
			im.notes,
			im.created_at,
			im.updated_at,
			(SELECT COUNT(*) FROM lab.entries e WHERE e.investment_model_id = im.id AND e.deleted_at IS NULL)::int AS n_entries,
			(SELECT COUNT(*) FROM lab.evaluations ev JOIN lab.entries e ON e.id = ev.entry_id WHERE e.investment_model_id = im.id AND ev.deleted_at IS NULL AND e.deleted_at IS NULL)::int AS n_evaluations
		FROM lab.investment_models im
		WHERE im.id = $1::uuid AND im.deleted_at IS NULL
	`

	var m lab.InvestmentModel
	var paramsStr string
	err := r.pool.QueryRow(ctx, query, id).Scan(&m.ID, &m.Name, &m.Kind, &paramsStr, &m.Notes, &m.CreatedAt, &m.UpdatedAt, &m.NEntries, &m.NEvaluations)
	if err == pgx.ErrNoRows {
		return nil, &apperrors.NotFoundError{Resource: "investment_model", ID: id}
	}
	if err != nil {
		return nil, err
	}
	m.ParamsJSON = json.RawMessage(paramsStr)
	return &m, nil
}

// GetModelLeaderboard returns the model leaderboard.
func (r *LabRepository) GetModelLeaderboard() ([]lab.LeaderboardEntry, error) {
	ctx := context.Background()

	query := `
		SELECT
			investment_model_id::text,
			model_name,
			model_kind,
			n_entries::int,
			n_entries_with_predictions::int,
			n_evaluations::int,
			n_calcuttas_with_entries::int,
			n_calcuttas_with_evaluations::int,
			avg_mean_payout,
			avg_median_payout,
			avg_p_top1,
			avg_p_in_money,
			first_eval_at,
			last_eval_at
		FROM lab.model_leaderboard
		ORDER BY avg_mean_payout DESC NULLS LAST
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]lab.LeaderboardEntry, 0)
	for rows.Next() {
		var e lab.LeaderboardEntry
		if err := rows.Scan(
			&e.InvestmentModelID,
			&e.ModelName,
			&e.ModelKind,
			&e.NEntries,
			&e.NEntriesWithPredictions,
			&e.NEvaluations,
			&e.NCalcuttasWithEntries,
			&e.NCalcuttasWithEvaluations,
			&e.AvgMeanPayout,
			&e.AvgMedianPayout,
			&e.AvgPTop1,
			&e.AvgPInMoney,
			&e.FirstEvalAt,
			&e.LastEvalAt,
		); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// ListEntries returns entries matching the filter.
func (r *LabRepository) ListEntries(filter lab.ListEntriesFilter, page lab.Pagination) ([]lab.EntryDetail, error) {
	ctx := context.Background()

	query := `
		SELECT
			e.id::text,
			e.investment_model_id::text,
			e.calcutta_id::text,
			e.game_outcome_kind,
			e.game_outcome_params_json::text,
			e.optimizer_kind,
			e.optimizer_params_json::text,
			e.starting_state_key,
			e.bids_json::text,
			e.created_at,
			e.updated_at,
			im.name AS model_name,
			im.kind AS model_kind,
			c.name AS calcutta_name,
			(SELECT COUNT(*) FROM lab.evaluations ev WHERE ev.entry_id = e.id AND ev.deleted_at IS NULL)::int AS n_evaluations
		FROM lab.entries e
		JOIN lab.investment_models im ON im.id = e.investment_model_id
		JOIN core.calcuttas c ON c.id = e.calcutta_id
		WHERE e.deleted_at IS NULL
	`
	args := []any{}
	argIdx := 1

	if filter.InvestmentModelID != nil && *filter.InvestmentModelID != "" {
		query += ` AND e.investment_model_id = $` + labItoa(argIdx) + `::uuid`
		args = append(args, *filter.InvestmentModelID)
		argIdx++
	}
	if filter.CalcuttaID != nil && *filter.CalcuttaID != "" {
		query += ` AND e.calcutta_id = $` + labItoa(argIdx) + `::uuid`
		args = append(args, *filter.CalcuttaID)
		argIdx++
	}
	if filter.StartingStateKey != nil && *filter.StartingStateKey != "" {
		query += ` AND e.starting_state_key = $` + labItoa(argIdx)
		args = append(args, *filter.StartingStateKey)
		argIdx++
	}

	query += ` ORDER BY e.created_at DESC`
	query += ` LIMIT $` + labItoa(argIdx) + ` OFFSET $` + labItoa(argIdx+1)
	args = append(args, page.Limit, page.Offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]lab.EntryDetail, 0)
	for rows.Next() {
		var e lab.EntryDetail
		var gameOutcomeParamsStr, optimizerParamsStr, bidsStr string
		if err := rows.Scan(
			&e.ID,
			&e.InvestmentModelID,
			&e.CalcuttaID,
			&e.GameOutcomeKind,
			&gameOutcomeParamsStr,
			&e.OptimizerKind,
			&optimizerParamsStr,
			&e.StartingStateKey,
			&bidsStr,
			&e.CreatedAt,
			&e.UpdatedAt,
			&e.ModelName,
			&e.ModelKind,
			&e.CalcuttaName,
			&e.NEvaluations,
		); err != nil {
			return nil, err
		}
		e.GameOutcomeParamsJSON = json.RawMessage(gameOutcomeParamsStr)
		e.OptimizerParamsJSON = json.RawMessage(optimizerParamsStr)
		e.BidsJSON = json.RawMessage(bidsStr)
		out = append(out, e)
	}
	return out, rows.Err()
}

// GetEntry returns a single entry by ID with full details.
func (r *LabRepository) GetEntry(id string) (*lab.EntryDetail, error) {
	ctx := context.Background()

	query := `
		SELECT
			e.id::text,
			e.investment_model_id::text,
			e.calcutta_id::text,
			e.game_outcome_kind,
			e.game_outcome_params_json::text,
			e.optimizer_kind,
			e.optimizer_params_json::text,
			e.starting_state_key,
			e.bids_json::text,
			e.created_at,
			e.updated_at,
			im.name AS model_name,
			im.kind AS model_kind,
			c.name AS calcutta_name,
			(SELECT COUNT(*) FROM lab.evaluations ev WHERE ev.entry_id = e.id AND ev.deleted_at IS NULL)::int AS n_evaluations
		FROM lab.entries e
		JOIN lab.investment_models im ON im.id = e.investment_model_id
		JOIN core.calcuttas c ON c.id = e.calcutta_id
		WHERE e.id = $1::uuid AND e.deleted_at IS NULL
	`

	var e lab.EntryDetail
	var gameOutcomeParamsStr, optimizerParamsStr, bidsStr string
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&e.ID,
		&e.InvestmentModelID,
		&e.CalcuttaID,
		&e.GameOutcomeKind,
		&gameOutcomeParamsStr,
		&e.OptimizerKind,
		&optimizerParamsStr,
		&e.StartingStateKey,
		&bidsStr,
		&e.CreatedAt,
		&e.UpdatedAt,
		&e.ModelName,
		&e.ModelKind,
		&e.CalcuttaName,
		&e.NEvaluations,
	)
	if err == pgx.ErrNoRows {
		return nil, &apperrors.NotFoundError{Resource: "entry", ID: id}
	}
	if err != nil {
		return nil, err
	}
	e.GameOutcomeParamsJSON = json.RawMessage(gameOutcomeParamsStr)
	e.OptimizerParamsJSON = json.RawMessage(optimizerParamsStr)
	e.BidsJSON = json.RawMessage(bidsStr)
	return &e, nil
}

// ListEvaluations returns evaluations matching the filter.
func (r *LabRepository) ListEvaluations(filter lab.ListEvaluationsFilter, page lab.Pagination) ([]lab.EvaluationDetail, error) {
	ctx := context.Background()

	query := `
		SELECT
			ev.id::text,
			ev.entry_id::text,
			ev.n_sims,
			ev.seed,
			ev.mean_normalized_payout,
			ev.median_normalized_payout,
			ev.p_top1,
			ev.p_in_money,
			ev.our_rank,
			ev.simulated_calcutta_id::text,
			ev.created_at,
			ev.updated_at,
			im.name AS model_name,
			im.kind AS model_kind,
			c.id::text AS calcutta_id,
			c.name AS calcutta_name,
			e.starting_state_key
		FROM lab.evaluations ev
		JOIN lab.entries e ON e.id = ev.entry_id
		JOIN lab.investment_models im ON im.id = e.investment_model_id
		JOIN core.calcuttas c ON c.id = e.calcutta_id
		WHERE ev.deleted_at IS NULL AND e.deleted_at IS NULL
	`
	args := []any{}
	argIdx := 1

	if filter.EntryID != nil && *filter.EntryID != "" {
		query += ` AND ev.entry_id = $` + labItoa(argIdx) + `::uuid`
		args = append(args, *filter.EntryID)
		argIdx++
	}
	if filter.InvestmentModelID != nil && *filter.InvestmentModelID != "" {
		query += ` AND e.investment_model_id = $` + labItoa(argIdx) + `::uuid`
		args = append(args, *filter.InvestmentModelID)
		argIdx++
	}
	if filter.CalcuttaID != nil && *filter.CalcuttaID != "" {
		query += ` AND e.calcutta_id = $` + labItoa(argIdx) + `::uuid`
		args = append(args, *filter.CalcuttaID)
		argIdx++
	}

	query += ` ORDER BY ev.mean_normalized_payout DESC NULLS LAST, ev.created_at DESC`
	query += ` LIMIT $` + labItoa(argIdx) + ` OFFSET $` + labItoa(argIdx+1)
	args = append(args, page.Limit, page.Offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]lab.EvaluationDetail, 0)
	for rows.Next() {
		var ev lab.EvaluationDetail
		if err := rows.Scan(
			&ev.ID,
			&ev.EntryID,
			&ev.NSims,
			&ev.Seed,
			&ev.MeanNormalizedPayout,
			&ev.MedianNormalizedPayout,
			&ev.PTop1,
			&ev.PInMoney,
			&ev.OurRank,
			&ev.SimulatedCalcuttaID,
			&ev.CreatedAt,
			&ev.UpdatedAt,
			&ev.ModelName,
			&ev.ModelKind,
			&ev.CalcuttaID,
			&ev.CalcuttaName,
			&ev.StartingStateKey,
		); err != nil {
			return nil, err
		}
		out = append(out, ev)
	}
	return out, rows.Err()
}

// GetEvaluation returns a single evaluation by ID with full details.
func (r *LabRepository) GetEvaluation(id string) (*lab.EvaluationDetail, error) {
	ctx := context.Background()

	query := `
		SELECT
			ev.id::text,
			ev.entry_id::text,
			ev.n_sims,
			ev.seed,
			ev.mean_normalized_payout,
			ev.median_normalized_payout,
			ev.p_top1,
			ev.p_in_money,
			ev.our_rank,
			ev.simulated_calcutta_id::text,
			ev.created_at,
			ev.updated_at,
			im.name AS model_name,
			im.kind AS model_kind,
			c.id::text AS calcutta_id,
			c.name AS calcutta_name,
			e.starting_state_key
		FROM lab.evaluations ev
		JOIN lab.entries e ON e.id = ev.entry_id
		JOIN lab.investment_models im ON im.id = e.investment_model_id
		JOIN core.calcuttas c ON c.id = e.calcutta_id
		WHERE ev.id = $1::uuid AND ev.deleted_at IS NULL AND e.deleted_at IS NULL
	`

	var ev lab.EvaluationDetail
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&ev.ID,
		&ev.EntryID,
		&ev.NSims,
		&ev.Seed,
		&ev.MeanNormalizedPayout,
		&ev.MedianNormalizedPayout,
		&ev.PTop1,
		&ev.PInMoney,
		&ev.OurRank,
		&ev.SimulatedCalcuttaID,
		&ev.CreatedAt,
		&ev.UpdatedAt,
		&ev.ModelName,
		&ev.ModelKind,
		&ev.CalcuttaID,
		&ev.CalcuttaName,
		&ev.StartingStateKey,
	)
	if err == pgx.ErrNoRows {
		return nil, &apperrors.NotFoundError{Resource: "evaluation", ID: id}
	}
	if err != nil {
		return nil, err
	}
	return &ev, nil
}

// GetEntryEnriched returns a single entry with enriched bid data (team names, seeds, naive allocation).
func (r *LabRepository) GetEntryEnriched(id string) (*lab.EntryDetailEnriched, error) {
	ctx := context.Background()

	// First get the basic entry details
	query := `
		SELECT
			e.id::text,
			e.investment_model_id::text,
			e.calcutta_id::text,
			e.game_outcome_kind,
			e.game_outcome_params_json::text,
			e.optimizer_kind,
			e.optimizer_params_json::text,
			e.starting_state_key,
			e.predictions_json::text,
			e.bids_json::text,
			e.created_at,
			e.updated_at,
			im.name AS model_name,
			im.kind AS model_kind,
			c.name AS calcutta_name,
			c.tournament_id::text,
			(SELECT COUNT(*) FROM lab.evaluations ev WHERE ev.entry_id = e.id AND ev.deleted_at IS NULL)::int AS n_evaluations
		FROM lab.entries e
		JOIN lab.investment_models im ON im.id = e.investment_model_id
		JOIN core.calcuttas c ON c.id = e.calcutta_id
		WHERE e.id = $1::uuid AND e.deleted_at IS NULL
	`

	var result lab.EntryDetailEnriched
	var (
		tournamentID                                     string
		gameOutcomeParamsStr, optimizerParamsStr, bidsStr string
		predictionsStr                                   *string
	)

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&result.ID, &result.InvestmentModelID, &result.CalcuttaID,
		&result.GameOutcomeKind, &gameOutcomeParamsStr, &result.OptimizerKind, &optimizerParamsStr,
		&result.StartingStateKey, &predictionsStr, &bidsStr, &result.CreatedAt, &result.UpdatedAt,
		&result.ModelName, &result.ModelKind, &result.CalcuttaName, &tournamentID, &result.NEvaluations,
	)
	if err == pgx.ErrNoRows {
		return nil, &apperrors.NotFoundError{Resource: "entry", ID: id}
	}
	if err != nil {
		return nil, err
	}

	// Parse the raw predictions (if present)
	var rawPredictions []lab.Prediction
	result.HasPredictions = predictionsStr != nil && *predictionsStr != ""
	if result.HasPredictions {
		if err := json.Unmarshal([]byte(*predictionsStr), &rawPredictions); err != nil {
			return nil, err
		}
	}

	// Parse the raw bids
	var rawBids []lab.EntryBid
	if err := json.Unmarshal([]byte(bidsStr), &rawBids); err != nil {
		return nil, err
	}

	// Get team info for all teams in this tournament
	teamQuery := `
		SELECT t.id::text, s.name, t.seed, t.region
		FROM core.teams t
		JOIN core.schools s ON s.id = t.school_id AND s.deleted_at IS NULL
		WHERE t.tournament_id = $1::uuid AND t.deleted_at IS NULL
	`
	teamRows, err := r.pool.Query(ctx, teamQuery, tournamentID)
	if err != nil {
		return nil, err
	}
	defer teamRows.Close()

	type teamInfo struct {
		Name   string
		Seed   int
		Region string
	}
	teamMap := make(map[string]teamInfo)
	for teamRows.Next() {
		var tid, name, region string
		var seed int
		if err := teamRows.Scan(&tid, &name, &seed, &region); err != nil {
			return nil, err
		}
		teamMap[tid] = teamInfo{Name: name, Seed: seed, Region: region}
	}
	if err := teamRows.Err(); err != nil {
		return nil, err
	}

	// Get total pool budget (number of entries × budget per entry)
	// This is needed for predictions because predicted_market_share is a fraction of the TOTAL pool
	var totalPoolBudget int
	poolBudgetQuery := `
		SELECT c.budget_points * COUNT(e.id)::int
		FROM core.calcuttas c
		LEFT JOIN core.entries e ON e.calcutta_id = c.id AND e.deleted_at IS NULL
		WHERE c.id = $1::uuid AND c.deleted_at IS NULL
		GROUP BY c.budget_points
	`
	if err := r.pool.QueryRow(ctx, poolBudgetQuery, result.CalcuttaID).Scan(&totalPoolBudget); err != nil || totalPoolBudget <= 0 {
		// Fallback: reasonable default
		totalPoolBudget = 4200
	}

	// Calculate total budget from our bids (for naive allocation comparison)
	totalBudget := 0
	for _, b := range rawBids {
		totalBudget += b.BidPoints
	}

	// Build teamExpectedPoints from predictions (single source of truth from Python)
	// This ensures consistency between predictions tab and bids tab
	teamExpectedPoints := make(map[string]float64)
	if result.HasPredictions {
		for _, p := range rawPredictions {
			teamExpectedPoints[p.TeamID] = p.ExpectedPoints
		}
	} else {
		// Fallback: seed-based estimates (only used when no predictions exist)
		seedExpectedPoints := map[int]float64{
			1: 320, 2: 240, 3: 160, 4: 120, 5: 80, 6: 64, 7: 48, 8: 40,
			9: 32, 10: 24, 11: 20, 12: 16, 13: 8, 14: 4, 15: 2, 16: 1,
		}
		for tid, ti := range teamMap {
			teamExpectedPoints[tid] = seedExpectedPoints[ti.Seed]
		}
	}

	// Calculate total expected points for normalization
	totalExpectedPoints := 0.0
	for _, ev := range teamExpectedPoints {
		totalExpectedPoints += ev
	}

	// Build a map of bid points and expected ROI by team ID for quick lookup
	bidPointsByTeam := make(map[string]int)
	expectedROIByTeam := make(map[string]*float64)
	for _, b := range rawBids {
		bidPointsByTeam[b.TeamID] = b.BidPoints
		expectedROIByTeam[b.TeamID] = b.ExpectedROI
	}

	// Build enriched bids for ALL teams (not just those with bids > 0)
	enrichedBids := make([]lab.EnrichedBid, 0, len(teamMap))
	for tid, ti := range teamMap {
		bidPoints := bidPointsByTeam[tid] // 0 if not in map

		// Naive allocation: team's expected points / total expected points * budget
		naiveShare := teamExpectedPoints[tid] / totalExpectedPoints
		naivePoints := int(naiveShare * float64(totalBudget))

		// Edge: (naive - bid) / naive * 100 (positive = undervalued opportunity)
		edgePercent := 0.0
		if naivePoints > 0 {
			edgePercent = float64(naivePoints-bidPoints) / float64(naivePoints) * 100
		}

		enrichedBids = append(enrichedBids, lab.EnrichedBid{
			TeamID:      tid,
			SchoolName:  ti.Name,
			Seed:        ti.Seed,
			Region:      ti.Region,
			BidPoints:   bidPoints,
			NaivePoints: naivePoints,
			EdgePercent: edgePercent,
			ExpectedROI: expectedROIByTeam[tid],
		})
	}

	// Build enriched predictions (if present)
	if result.HasPredictions {
		enrichedPredictions := make([]lab.EnrichedPrediction, 0, len(rawPredictions))
		for _, p := range rawPredictions {
			ti, ok := teamMap[p.TeamID]
			if !ok {
				continue
			}

			// Convert market share to bid points (using total POOL budget, not our single entry budget)
			predictedBidPoints := int(p.PredictedMarketShare * float64(totalPoolBudget))

			// Expected ROI = expected points / predicted market bid
			expectedROI := 0.0
			if predictedBidPoints > 0 {
				expectedROI = p.ExpectedPoints / float64(predictedBidPoints)
			}

			// Naive allocation for comparison (also uses total pool budget for predictions)
			naiveShare := teamExpectedPoints[p.TeamID] / totalExpectedPoints
			naivePoints := int(naiveShare * float64(totalPoolBudget))

			// Edge: (naive - predicted) / naive * 100
			edgePercent := 0.0
			if naivePoints > 0 {
				edgePercent = float64(naivePoints-predictedBidPoints) / float64(naivePoints) * 100
			}

			enrichedPredictions = append(enrichedPredictions, lab.EnrichedPrediction{
				TeamID:               p.TeamID,
				SchoolName:           ti.Name,
				Seed:                 ti.Seed,
				Region:               ti.Region,
				PredictedMarketShare: p.PredictedMarketShare,
				PredictedBidPoints:   predictedBidPoints,
				ExpectedPoints:       p.ExpectedPoints,
				ExpectedROI:          expectedROI,
				NaivePoints:          naivePoints,
				EdgePercent:          edgePercent,
			})
		}
		result.Predictions = enrichedPredictions
	}

	// Set the remaining fields
	result.GameOutcomeParamsJSON = json.RawMessage(gameOutcomeParamsStr)
	result.OptimizerParamsJSON = json.RawMessage(optimizerParamsStr)
	result.Bids = enrichedBids

	return &result, nil
}

// GetEntryEnrichedByModelAndCalcutta returns an enriched entry for a model/calcutta pair.
// Defaults to starting_state_key='post_first_four' if not specified.
func (r *LabRepository) GetEntryEnrichedByModelAndCalcutta(modelName, calcuttaID, startingStateKey string) (*lab.EntryDetailEnriched, error) {
	ctx := context.Background()

	if startingStateKey == "" {
		startingStateKey = "post_first_four"
	}

	// Find the entry ID by model name, calcutta ID, and starting state key
	var entryID string
	err := r.pool.QueryRow(ctx, `
		SELECT e.id::text
		FROM lab.entries e
		JOIN lab.investment_models im ON im.id = e.investment_model_id AND im.deleted_at IS NULL
		WHERE im.name = $1
			AND e.calcutta_id = $2::uuid
			AND e.starting_state_key = $3
			AND e.deleted_at IS NULL
		ORDER BY e.created_at DESC
		LIMIT 1
	`, modelName, calcuttaID, startingStateKey).Scan(&entryID)
	if err == pgx.ErrNoRows {
		return nil, &apperrors.NotFoundError{Resource: "entry", ID: modelName + "/" + calcuttaID}
	}
	if err != nil {
		return nil, err
	}

	// Delegate to GetEntryEnriched
	return r.GetEntryEnriched(entryID)
}

// GetEvaluationEntryResults returns per-entry results for an evaluation.
func (r *LabRepository) GetEvaluationEntryResults(evaluationID string) ([]lab.EvaluationEntryResult, error) {
	ctx := context.Background()

	query := `
		SELECT
			entry_name,
			mean_normalized_payout,
			p_top1,
			p_in_money,
			rank
		FROM lab.evaluation_entry_results
		WHERE evaluation_id = $1::uuid
		ORDER BY rank ASC
	`

	rows, err := r.pool.Query(ctx, query, evaluationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]lab.EvaluationEntryResult, 0)
	for rows.Next() {
		var e lab.EvaluationEntryResult
		if err := rows.Scan(
			&e.EntryName,
			&e.MeanNormalizedPayout,
			&e.PTop1,
			&e.PInMoney,
			&e.Rank,
		); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// GetEvaluationEntryProfile returns detailed profile for an entry in an evaluation.
func (r *LabRepository) GetEvaluationEntryProfile(evaluationID, entryName string) (*lab.EvaluationEntryProfile, error) {
	ctx := context.Background()

	// First, get the entry result from lab.evaluation_entry_results
	var profile lab.EvaluationEntryProfile
	err := r.pool.QueryRow(ctx, `
		SELECT
			entry_name,
			mean_normalized_payout,
			p_top1,
			p_in_money,
			rank
		FROM lab.evaluation_entry_results
		WHERE evaluation_id = $1::uuid
			AND entry_name = $2
	`, evaluationID, entryName).Scan(
		&profile.EntryName,
		&profile.MeanNormalizedPayout,
		&profile.PTop1,
		&profile.PInMoney,
		&profile.Rank,
	)
	if err == pgx.ErrNoRows {
		return nil, &apperrors.NotFoundError{Resource: "evaluation_entry", ID: evaluationID + "/" + entryName}
	}
	if err != nil {
		return nil, err
	}

	// Get the simulated_calcutta_id from the evaluation
	var simulatedCalcuttaID *string
	err = r.pool.QueryRow(ctx, `
		SELECT simulated_calcutta_id
		FROM lab.evaluations
		WHERE id = $1::uuid
			AND deleted_at IS NULL
	`, evaluationID).Scan(&simulatedCalcuttaID)
	if err != nil {
		return nil, err
	}

	if simulatedCalcuttaID == nil {
		// No simulated calcutta - return profile without bids
		profile.Bids = []lab.EvaluationEntryBid{}
		return &profile, nil
	}

	// Get bids from derived.simulated_entries + derived.simulated_entry_teams
	// Also compute ownership percentage
	query := `
		WITH entry_bids AS (
			SELECT
				set.team_id,
				set.bid_points
			FROM derived.simulated_entries se
			JOIN derived.simulated_entry_teams set ON set.simulated_entry_id = se.id AND set.deleted_at IS NULL
			WHERE se.simulated_calcutta_id = $1::uuid
				AND se.display_name = $2
				AND se.deleted_at IS NULL
		),
		total_pool AS (
			SELECT
				set.team_id,
				SUM(set.bid_points)::float AS total_bid
			FROM derived.simulated_entries se
			JOIN derived.simulated_entry_teams set ON set.simulated_entry_id = se.id AND set.deleted_at IS NULL
			WHERE se.simulated_calcutta_id = $1::uuid
				AND se.deleted_at IS NULL
			GROUP BY set.team_id
		)
		SELECT
			eb.team_id,
			s.name AS school_name,
			t.seed,
			t.region,
			eb.bid_points,
			COALESCE(eb.bid_points::float / NULLIF(tp.total_bid, 0), 0) AS ownership
		FROM entry_bids eb
		JOIN core.teams t ON t.id = eb.team_id AND t.deleted_at IS NULL
		JOIN core.schools s ON s.id = t.school_id AND s.deleted_at IS NULL
		LEFT JOIN total_pool tp ON tp.team_id = eb.team_id
		WHERE eb.bid_points > 0
		ORDER BY t.seed ASC, s.name ASC
	`

	rows, err := r.pool.Query(ctx, query, *simulatedCalcuttaID, entryName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	profile.Bids = make([]lab.EvaluationEntryBid, 0)
	profile.TotalBidPoints = 0
	for rows.Next() {
		var bid lab.EvaluationEntryBid
		if err := rows.Scan(
			&bid.TeamID,
			&bid.SchoolName,
			&bid.Seed,
			&bid.Region,
			&bid.BidPoints,
			&bid.Ownership,
		); err != nil {
			return nil, err
		}
		profile.TotalBidPoints += bid.BidPoints
		profile.Bids = append(profile.Bids, bid)
	}

	return &profile, rows.Err()
}

// labItoa converts int to string for building parameterized queries.
func labItoa(i int) string {
	return strconv.Itoa(i)
}
