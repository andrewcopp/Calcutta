package httpserver

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/dtos"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
)

func (s *Server) getSuiteCalcuttaEvaluationSnapshotEntryHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	evalID := vars["id"]
	snapshotEntryID := vars["snapshotEntryId"]
	if evalID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "id is required", "id")
		return
	}
	if snapshotEntryID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "snapshotEntryId is required", "snapshotEntryId")
		return
	}

	ctx := r.Context()

	var calcuttaEvaluationRunID *string
	if err := s.pool.QueryRow(ctx, `
		SELECT calcutta_evaluation_run_id::text
		FROM derived.suite_calcutta_evaluations
		WHERE id = $1::uuid
			AND deleted_at IS NULL
		LIMIT 1
	`, evalID).Scan(&calcuttaEvaluationRunID); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	if calcuttaEvaluationRunID == nil || strings.TrimSpace(*calcuttaEvaluationRunID) == "" {
		writeError(w, r, http.StatusConflict, "invalid_state", "Evaluation has no calcutta_evaluation_run_id", "calcutta_evaluation_run_id")
		return
	}

	var snapshotID string
	if err := s.pool.QueryRow(ctx, `
		SELECT calcutta_snapshot_id::text
		FROM derived.calcutta_evaluation_runs
		WHERE id = $1::uuid
			AND deleted_at IS NULL
		LIMIT 1
	`, *calcuttaEvaluationRunID).Scan(&snapshotID); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	var displayName string
	var isSynthetic bool
	if err := s.pool.QueryRow(ctx, `
		SELECT display_name, is_synthetic
		FROM core.calcutta_snapshot_entries
		WHERE id = $1::uuid
			AND calcutta_snapshot_id = $2::uuid
			AND deleted_at IS NULL
		LIMIT 1
	`, snapshotEntryID, snapshotID).Scan(&displayName, &isSynthetic); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, r, http.StatusNotFound, "not_found", "Snapshot entry not found for this evaluation", "snapshotEntryId")
			return
		}
		writeErrorFromErr(w, r, err)
		return
	}

	rows, err := s.pool.Query(ctx, `
		SELECT
			t.id::text,
			s.name,
			COALESCE(t.seed, 0)::int,
			COALESCE(t.region, ''::text),
			cset.bid_points::int
		FROM core.calcutta_snapshot_entry_teams cset
		JOIN core.teams t ON t.id = cset.team_id AND t.deleted_at IS NULL
		JOIN core.schools s ON s.id = t.school_id AND s.deleted_at IS NULL
		WHERE cset.calcutta_snapshot_entry_id = $1::uuid
			AND cset.deleted_at IS NULL
		ORDER BY cset.bid_points DESC
	`, snapshotEntryID)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	defer rows.Close()

	teams := make([]suiteCalcuttaSnapshotEntryTeam, 0)
	for rows.Next() {
		var t suiteCalcuttaSnapshotEntryTeam
		if err := rows.Scan(&t.TeamID, &t.School, &t.Seed, &t.Region, &t.BidPoints); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
		teams = append(teams, t)
	}
	if err := rows.Err(); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, suiteCalcuttaSnapshotEntryResponse{
		SnapshotEntryID: snapshotEntryID,
		DisplayName:     displayName,
		IsSynthetic:     isSynthetic,
		Teams:           teams,
	})
}

func (s *Server) createSuiteCalcuttaEvaluationHandler(w http.ResponseWriter, r *http.Request) {
	var req dtos.CreateSuiteCalcuttaEvaluationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}
	if err := req.Validate(); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	ctx := r.Context()

	suiteExecutionID := ""
	if req.SuiteExecutionID != nil {
		suiteExecutionID = *req.SuiteExecutionID
	}

	// Resolve (or create) suite_id.
	suiteID := ""
	evalOptimizerKey := ""
	if suiteExecutionID != "" {
		// Inherit suite + config from suite execution.
		var execSuiteID string
		var execOptimizerKey *string
		var execNSims *int
		var execSeed *int
		var execStartingStateKey string
		var execExcludedEntryName *string
		var goAlgID string
		var msAlgID string
		var suiteOptimizerKey string
		var suiteNSims int
		var suiteSeed int
		if err := s.pool.QueryRow(ctx, `
			SELECT
				e.suite_id::text,
				e.optimizer_key,
				e.n_sims,
				e.seed,
				e.starting_state_key,
				e.excluded_entry_name,
				COALESCE(s.game_outcomes_algorithm_id::text, ''::text) AS game_outcomes_algorithm_id,
				COALESCE(s.market_share_algorithm_id::text, ''::text) AS market_share_algorithm_id,
				COALESCE(s.optimizer_key, ''::text) AS suite_optimizer_key,
				COALESCE(s.n_sims, 0)::int AS suite_n_sims,
				COALESCE(s.seed, 0)::int AS suite_seed
			FROM derived.suite_executions e
			JOIN derived.suites s ON s.id = e.suite_id AND s.deleted_at IS NULL
			WHERE e.id = $1::uuid
				AND e.deleted_at IS NULL
			LIMIT 1
		`, suiteExecutionID).Scan(
			&execSuiteID,
			&execOptimizerKey,
			&execNSims,
			&execSeed,
			&execStartingStateKey,
			&execExcludedEntryName,
			&goAlgID,
			&msAlgID,
			&suiteOptimizerKey,
			&suiteNSims,
			&suiteSeed,
		); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}

		suiteID = execSuiteID

		// Determine effective optimizer/nSims/seed using request overrides then execution then suite defaults.
		if req.OptimizerKey != nil {
			evalOptimizerKey = *req.OptimizerKey
		} else if execOptimizerKey != nil && *execOptimizerKey != "" {
			evalOptimizerKey = *execOptimizerKey
		} else {
			evalOptimizerKey = suiteOptimizerKey
		}

		if execNSims != nil {
			req.NSims = *execNSims
		} else if suiteNSims > 0 {
			req.NSims = suiteNSims
		}
		if execSeed != nil {
			req.Seed = *execSeed
		} else if suiteSeed != 0 {
			req.Seed = suiteSeed
		}
		if execStartingStateKey != "" {
			req.StartingStateKey = execStartingStateKey
		}
		if execExcludedEntryName != nil {
			req.ExcludedEntryName = execExcludedEntryName
		}

		// Resolve run IDs from suite algorithms if not explicitly provided.
		if req.GameOutcomeRunID == nil || req.MarketShareRunID == nil {
			var tournamentID string
			if err := s.pool.QueryRow(ctx, `
				SELECT tournament_id::text
				FROM core.calcuttas
				WHERE id = $1::uuid
					AND deleted_at IS NULL
				LIMIT 1
			`, req.CalcuttaID).Scan(&tournamentID); err != nil {
				writeErrorFromErr(w, r, err)
				return
			}

			if req.GameOutcomeRunID == nil {
				var resolved string
				_ = s.pool.QueryRow(ctx, `
					SELECT COALESCE((
						SELECT id::text
						FROM derived.game_outcome_runs
						WHERE tournament_id = $1::uuid
							AND algorithm_id = $2::uuid
							AND deleted_at IS NULL
						ORDER BY created_at DESC
						LIMIT 1
					), ''::text) AS id
				`, tournamentID, goAlgID).Scan(&resolved)
				if resolved == "" {
					writeError(w, r, http.StatusConflict, "missing_run", "Missing game-outcome run for suite execution", "gameOutcomeRunId")
					return
				}
				req.GameOutcomeRunID = &resolved
			}

			if req.MarketShareRunID == nil {
				var resolved string
				_ = s.pool.QueryRow(ctx, `
					SELECT COALESCE((
						SELECT id::text
						FROM derived.market_share_runs
						WHERE calcutta_id = $1::uuid
							AND algorithm_id = $2::uuid
							AND deleted_at IS NULL
						ORDER BY created_at DESC
						LIMIT 1
					), ''::text) AS id
				`, req.CalcuttaID, msAlgID).Scan(&resolved)
				if resolved == "" {
					writeError(w, r, http.StatusConflict, "missing_run", "Missing market-share run for suite execution", "marketShareRunId")
					return
				}
				req.MarketShareRunID = &resolved
			}
		}
	} else if req.SuiteID != nil {
		suiteID = *req.SuiteID
		if req.OptimizerKey != nil {
			evalOptimizerKey = *req.OptimizerKey
		} else {
			_ = s.pool.QueryRow(ctx, `
				SELECT COALESCE(optimizer_key, '')
				FROM derived.suites
				WHERE id = $1::uuid
					AND deleted_at IS NULL
				LIMIT 1
			`, suiteID).Scan(&evalOptimizerKey)
		}
	} else {
		// Resolve algorithm ids from runs.
		var goAlgID string
		if err := s.pool.QueryRow(ctx, `
			SELECT algorithm_id::text
			FROM derived.game_outcome_runs
			WHERE id = $1::uuid
				AND deleted_at IS NULL
		`, *req.GameOutcomeRunID).Scan(&goAlgID); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}

		var msAlgID string
		if err := s.pool.QueryRow(ctx, `
			SELECT algorithm_id::text
			FROM derived.market_share_runs
			WHERE id = $1::uuid
				AND deleted_at IS NULL
		`, *req.MarketShareRunID).Scan(&msAlgID); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}

		var insertedID string
		if err := s.pool.QueryRow(ctx, `
			INSERT INTO derived.suites (
				name,
				description,
				game_outcomes_algorithm_id,
				market_share_algorithm_id,
				optimizer_key,
				n_sims,
				seed,
				params_json
			)
			VALUES ($1, NULL, $2::uuid, $3::uuid, $4, $5, $6, '{}'::jsonb)
			ON CONFLICT (name) WHERE deleted_at IS NULL
			DO UPDATE SET
				game_outcomes_algorithm_id = EXCLUDED.game_outcomes_algorithm_id,
				market_share_algorithm_id = EXCLUDED.market_share_algorithm_id,
				optimizer_key = EXCLUDED.optimizer_key,
				n_sims = EXCLUDED.n_sims,
				seed = EXCLUDED.seed,
				updated_at = NOW(),
				deleted_at = NULL
			RETURNING id
		`, *req.SuiteName, goAlgID, msAlgID, *req.OptimizerKey, req.NSims, req.Seed).Scan(&insertedID); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
		suiteID = insertedID
		evalOptimizerKey = *req.OptimizerKey
	}

	var evalID string
	var status string
	var goRun any
	if req.GameOutcomeRunID != nil {
		goRun = *req.GameOutcomeRunID
	} else {
		goRun = nil
	}
	var msRun any
	if req.MarketShareRunID != nil {
		msRun = *req.MarketShareRunID
	} else {
		msRun = nil
	}
	var excluded any
	if req.ExcludedEntryName != nil {
		excluded = *req.ExcludedEntryName
	} else {
		excluded = nil
	}

	q := `
		INSERT INTO derived.suite_calcutta_evaluations (
			suite_execution_id,
			suite_id,
			calcutta_id,
			game_outcome_run_id,
			market_share_run_id,
			optimizer_key,
			n_sims,
			seed,
			starting_state_key,
			excluded_entry_name
		)
		VALUES ($1::uuid, $2::uuid, $3::uuid, $4::uuid, $5::uuid, $6, $7::int, $8::int, $9, $10::text)
		RETURNING id, status
	`
	var execID any
	if suiteExecutionID != "" {
		execID = suiteExecutionID
	} else {
		execID = nil
	}
	var effNSims any
	if req.NSims > 0 {
		effNSims = req.NSims
	} else {
		effNSims = nil
	}
	var effSeed any
	if req.Seed != 0 {
		effSeed = req.Seed
	} else {
		effSeed = nil
	}
	if err := s.pool.QueryRow(ctx, q, execID, suiteID, req.CalcuttaID, goRun, msRun, evalOptimizerKey, effNSims, effSeed, req.StartingStateKey, excluded).Scan(&evalID, &status); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, createSuiteCalcuttaEvaluationResponse{ID: evalID, Status: status})
}

func (s *Server) listSuiteCalcuttaEvaluationsHandler(w http.ResponseWriter, r *http.Request) {
	calcuttaID := r.URL.Query().Get("calcutta_id")
	suiteID := r.URL.Query().Get("suite_id")
	if strings.TrimSpace(suiteID) == "" {
		suiteID = r.URL.Query().Get("cohort_id")
	}
	suiteExecutionID := r.URL.Query().Get("suite_execution_id")
	if strings.TrimSpace(suiteExecutionID) == "" {
		suiteExecutionID = r.URL.Query().Get("simulation_run_batch_id")
	}
	limit := getLimit(r, 50)
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	offset := getOffset(r, 0)
	if offset < 0 {
		offset = 0
	}

	items, err := s.loadSuiteCalcuttaEvaluations(r.Context(), calcuttaID, suiteID, suiteExecutionID, limit, offset)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, suiteCalcuttaEvaluationListResponse{Items: items})
}

func (s *Server) getSuiteCalcuttaEvaluationHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "id is required", "id")
		return
	}

	it, err := s.loadSuiteCalcuttaEvaluationByID(r.Context(), id)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, it)
}

func (s *Server) getSuiteCalcuttaEvaluationResultHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "id is required", "id")
		return
	}

	ctx := r.Context()

	eval, err := s.loadSuiteCalcuttaEvaluationByID(ctx, id)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	if eval.StrategyGenerationRunID == nil || *eval.StrategyGenerationRunID == "" {
		writeError(w, r, http.StatusConflict, "invalid_state", "Evaluation has no generated entry yet", "strategy_generation_run_id")
		return
	}

	rows, err := s.pool.Query(ctx, `
		SELECT
			t.id::text as team_id,
			s.name as school_name,
			COALESCE(t.seed, 0)::int as seed,
			COALESCE(t.region, ''::text) as region,
			reb.bid_points::int as bid_points,
			COALESCE(reb.expected_roi, 0.0)::double precision as expected_roi
		FROM derived.recommended_entry_bids reb
		JOIN core.teams t ON t.id = reb.team_id AND t.deleted_at IS NULL
		JOIN core.schools s ON s.id = t.school_id AND s.deleted_at IS NULL
		WHERE reb.strategy_generation_run_id = $1::uuid
			AND reb.deleted_at IS NULL
		ORDER BY reb.bid_points DESC
	`, *eval.StrategyGenerationRunID)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	defer rows.Close()

	portfolio := make([]suiteCalcuttaEvaluationPortfolioBid, 0)
	for rows.Next() {
		var b suiteCalcuttaEvaluationPortfolioBid
		if err := rows.Scan(&b.TeamID, &b.SchoolName, &b.Seed, &b.Region, &b.BidPoints, &b.ExpectedROI); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
		portfolio = append(portfolio, b)
	}
	if err := rows.Err(); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	var our *suiteCalcuttaEvaluationOurStrategyPerformance
	if eval.CalcuttaEvaluationRunID != nil && *eval.CalcuttaEvaluationRunID != "" {
		var tmp suiteCalcuttaEvaluationOurStrategyPerformance
		err := s.pool.QueryRow(ctx, `
			WITH ranked AS (
				SELECT
					ROW_NUMBER() OVER (ORDER BY COALESCE(ep.mean_normalized_payout, 0.0) DESC)::int AS rank,
					ep.entry_name,
					COALESCE(ep.mean_normalized_payout, 0.0)::double precision AS mean_normalized_payout,
					COALESCE(ep.median_normalized_payout, 0.0)::double precision AS median_normalized_payout,
					COALESCE(ep.p_top1, 0.0)::double precision AS p_top1,
					COALESCE(ep.p_in_money, 0.0)::double precision AS p_in_money
				FROM derived.entry_performance ep
				WHERE ep.calcutta_evaluation_run_id = $1::uuid
					AND ep.deleted_at IS NULL
			)
			SELECT
				r.rank,
				r.entry_name,
				r.mean_normalized_payout,
				r.median_normalized_payout,
				r.p_top1,
				r.p_in_money,
				COALESCE((
					SELECT st.n_sims::int
					FROM derived.calcutta_evaluation_runs cer
					JOIN derived.simulated_tournaments st
						ON st.id = cer.simulated_tournament_id
						AND st.deleted_at IS NULL
					WHERE cer.id = $1::uuid
						AND cer.deleted_at IS NULL
					LIMIT 1
				), 0)::int as total_simulations
			FROM ranked r
			WHERE r.entry_name IN ('Our Strategy', 'our_strategy', 'Out Strategy')
			ORDER BY r.rank ASC
			LIMIT 1
		`, *eval.CalcuttaEvaluationRunID).Scan(
			&tmp.Rank,
			&tmp.EntryName,
			&tmp.MeanNormalizedPayout,
			&tmp.MedianNormalizedPayout,
			&tmp.PTop1,
			&tmp.PInMoney,
			&tmp.TotalSimulations,
		)
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				writeErrorFromErr(w, r, err)
				return
			}
		} else {
			our = &tmp
		}
	}

	entries := make([]suiteCalcuttaEvaluationEntryPerformance, 0)
	finishByName := map[string]*suiteCalcuttaEvalFinish{}
	if eval.CalcuttaID != "" && eval.StrategyGenerationRunID != nil && strings.TrimSpace(*eval.StrategyGenerationRunID) != "" {
		if m, ok, err := s.computeHypotheticalFinishByEntryNameForStrategyRun(ctx, eval.CalcuttaID, *eval.StrategyGenerationRunID); err != nil {
			writeErrorFromErr(w, r, err)
			return
		} else if ok {
			finishByName = m
		}
	}

	if eval.CalcuttaEvaluationRunID != nil && strings.TrimSpace(*eval.CalcuttaEvaluationRunID) != "" {
		rows, err := s.pool.Query(ctx, `
			WITH cer AS (
				SELECT calcutta_snapshot_id
				FROM derived.calcutta_evaluation_runs
				WHERE id = $1::uuid
					AND deleted_at IS NULL
				LIMIT 1
			),
			ranked AS (
				SELECT
					ROW_NUMBER() OVER (ORDER BY COALESCE(ep.mean_normalized_payout, 0.0) DESC)::int AS rank,
					ep.entry_name,
					COALESCE(ep.mean_normalized_payout, 0.0)::double precision AS mean_normalized_payout,
					COALESCE(ep.p_top1, 0.0)::double precision AS p_top1,
					COALESCE(ep.p_in_money, 0.0)::double precision AS p_in_money
				FROM derived.entry_performance ep
				WHERE ep.calcutta_evaluation_run_id = $1::uuid
					AND ep.deleted_at IS NULL
			)
			SELECT
				r.rank,
				r.entry_name,
				se.id::text as snapshot_entry_id,
				r.mean_normalized_payout,
				r.p_top1,
				r.p_in_money
			FROM ranked r
			LEFT JOIN core.calcutta_snapshot_entries se
				ON se.calcutta_snapshot_id = (SELECT calcutta_snapshot_id FROM cer)
				AND se.display_name = r.entry_name
				AND se.deleted_at IS NULL
			ORDER BY r.rank ASC
		`, *eval.CalcuttaEvaluationRunID)
		if err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var it suiteCalcuttaEvaluationEntryPerformance
			if err := rows.Scan(&it.Rank, &it.EntryName, &it.SnapshotEntryID, &it.MeanNormalizedPayout, &it.PTop1, &it.PInMoney); err != nil {
				writeErrorFromErr(w, r, err)
				return
			}
			if f := finishByName[it.EntryName]; f != nil {
				it.FinishPosition = &f.FinishPosition
				it.IsTied = &f.IsTied
				it.InTheMoney = &f.InTheMoney
				it.PayoutCents = &f.PayoutCents
				it.TotalPoints = &f.TotalPoints
			}
			entries = append(entries, it)
		}
		if err := rows.Err(); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
	}

	writeJSON(w, http.StatusOK, suiteCalcuttaEvaluationResultResponse{
		Evaluation:  *eval,
		Portfolio:   portfolio,
		OurStrategy: our,
		Entries:     entries,
	})
}
