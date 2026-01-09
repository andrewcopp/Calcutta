package httpserver

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/dtos"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
)

func (s *Server) listCohortSimulationsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cohortID := strings.TrimSpace(vars["cohortId"])
	if cohortID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "cohortId is required", "cohortId")
		return
	}
	if _, err := uuid.Parse(cohortID); err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "cohortId must be a valid UUID", "cohortId")
		return
	}

	q := r.URL.Query()
	q.Set("cohort_id", cohortID)
	r.URL.RawQuery = q.Encode()
	s.listSuiteCalcuttaEvaluationsHandler(w, r)
}

func (s *Server) getCohortSimulationHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cohortID := strings.TrimSpace(vars["cohortId"])
	id := strings.TrimSpace(vars["id"])
	if cohortID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "cohortId is required", "cohortId")
		return
	}
	if _, err := uuid.Parse(cohortID); err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "cohortId must be a valid UUID", "cohortId")
		return
	}
	if id == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "id is required", "id")
		return
	}

	it, err := s.loadSuiteCalcuttaEvaluationByID(r.Context(), id)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	if strings.TrimSpace(it.CohortID) != cohortID {
		writeError(w, r, http.StatusNotFound, "not_found", "Simulation not found", "id")
		return
	}
	writeJSON(w, http.StatusOK, it)
}

func (s *Server) getCohortSimulationResultHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cohortID := strings.TrimSpace(vars["cohortId"])
	id := strings.TrimSpace(vars["id"])
	if cohortID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "cohortId is required", "cohortId")
		return
	}
	if _, err := uuid.Parse(cohortID); err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "cohortId must be a valid UUID", "cohortId")
		return
	}
	if id == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "id is required", "id")
		return
	}

	it, err := s.loadSuiteCalcuttaEvaluationByID(r.Context(), id)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	if strings.TrimSpace(it.CohortID) != cohortID {
		writeError(w, r, http.StatusNotFound, "not_found", "Simulation not found", "id")
		return
	}

	s.getSuiteCalcuttaEvaluationResultHandler(w, r)
}

func (s *Server) getCohortSimulationSnapshotEntryHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cohortID := strings.TrimSpace(vars["cohortId"])
	id := strings.TrimSpace(vars["id"])
	if cohortID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "cohortId is required", "cohortId")
		return
	}
	if _, err := uuid.Parse(cohortID); err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "cohortId must be a valid UUID", "cohortId")
		return
	}
	if id == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "id is required", "id")
		return
	}

	it, err := s.loadSuiteCalcuttaEvaluationByID(r.Context(), id)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	if strings.TrimSpace(it.CohortID) != cohortID {
		writeError(w, r, http.StatusNotFound, "not_found", "Simulation not found", "id")
		return
	}

	s.getSuiteCalcuttaEvaluationSnapshotEntryHandler(w, r)
}

func (s *Server) createCohortSimulationHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cohortID := strings.TrimSpace(vars["cohortId"])
	if cohortID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "cohortId is required", "cohortId")
		return
	}
	if _, err := uuid.Parse(cohortID); err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "cohortId must be a valid UUID", "cohortId")
		return
	}

	var req dtos.CreateSimulationRunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}
	if req.CohortID == nil {
		v := cohortID
		req.CohortID = &v
	}

	b, _ := json.Marshal(req)
	cloned := r.Clone(r.Context())
	cloned.Body = io.NopCloser(bytes.NewReader(b))
	s.createSuiteCalcuttaEvaluationHandler(w, cloned)
}

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
		FROM derived.simulation_runs
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
	var req dtos.CreateSimulationRunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}
	if err := req.Validate(); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	ctx := r.Context()

	simulationBatchID := ""
	if req.SimulationRunBatchID != nil {
		simulationBatchID = *req.SimulationRunBatchID
	}

	// Resolve cohort_id.
	cohortID := ""
	evalOptimizerKey := ""
	if simulationBatchID != "" {
		// Inherit cohort + config from simulation batch.
		var execCohortID string
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
				e.cohort_id::text,
				e.optimizer_key,
				e.n_sims,
				e.seed,
				e.starting_state_key,
				e.excluded_entry_name,
				COALESCE(s.game_outcomes_algorithm_id::text, ''::text) AS game_outcomes_algorithm_id,
				COALESCE(s.market_share_algorithm_id::text, ''::text) AS market_share_algorithm_id,
				COALESCE(s.optimizer_key, ''::text) AS cohort_optimizer_key,
				COALESCE(s.n_sims, 0)::int AS cohort_n_sims,
				COALESCE(s.seed, 0)::int AS cohort_seed
			FROM derived.simulation_run_batches e
			JOIN derived.synthetic_calcutta_cohorts s ON s.id = e.cohort_id AND s.deleted_at IS NULL
			WHERE e.id = $1::uuid
				AND e.deleted_at IS NULL
			LIMIT 1
		`, simulationBatchID).Scan(
			&execCohortID,
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

		cohortID = execCohortID

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

		// Resolve run IDs from cohort algorithms if not explicitly provided.
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
					writeError(w, r, http.StatusConflict, "missing_run", "Missing game-outcome run for simulation batch", "gameOutcomeRunId")
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
					writeError(w, r, http.StatusConflict, "missing_run", "Missing market-share run for simulation batch", "marketShareRunId")
					return
				}
				req.MarketShareRunID = &resolved
			}
		}
	} else {
		if req.CohortID != nil {
			cohortID = *req.CohortID
		}
		if strings.TrimSpace(cohortID) == "" {
			writeError(w, r, http.StatusBadRequest, "validation_error", "cohortId is required", "cohortId")
			return
		}
		if req.OptimizerKey != nil {
			evalOptimizerKey = *req.OptimizerKey
		} else {
			_ = s.pool.QueryRow(ctx, `
				SELECT COALESCE(optimizer_key, '')
				FROM derived.synthetic_calcutta_cohorts
				WHERE id = $1::uuid
					AND deleted_at IS NULL
				LIMIT 1
			`, cohortID).Scan(&evalOptimizerKey)
		}
	}

	var syntheticCalcuttaID string
	var syntheticSnapshotID *string
	var existingExcludedEntryName *string
	if err := s.pool.QueryRow(ctx, `
		INSERT INTO derived.synthetic_calcuttas (
			cohort_id,
			calcutta_id
		)
		VALUES ($1::uuid, $2::uuid)
		ON CONFLICT (cohort_id, calcutta_id) WHERE deleted_at IS NULL
		DO UPDATE SET
			updated_at = NOW(),
			deleted_at = NULL
		RETURNING id::text, calcutta_snapshot_id::text, excluded_entry_name
	`, cohortID, req.CalcuttaID).Scan(&syntheticCalcuttaID, &syntheticSnapshotID, &existingExcludedEntryName); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	if syntheticSnapshotID == nil || strings.TrimSpace(*syntheticSnapshotID) == "" {
		tx, err := s.pool.Begin(ctx)
		if err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
		committed := false
		defer func() {
			if committed {
				return
			}
			_ = tx.Rollback(ctx)
		}()

		createdSnapshotID, err := createSyntheticCalcuttaSnapshot(ctx, tx, req.CalcuttaID, existingExcludedEntryName, "", nil)
		if err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
		_, err = tx.Exec(ctx, `
			UPDATE derived.synthetic_calcuttas
			SET calcutta_snapshot_id = $2::uuid,
				updated_at = NOW()
			WHERE id = $1::uuid
				AND deleted_at IS NULL
		`, syntheticCalcuttaID, createdSnapshotID)
		if err != nil {
			writeErrorFromErr(w, r, err)
			return
		}

		if err := tx.Commit(ctx); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
		committed = true
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
		INSERT INTO derived.simulation_runs (
			simulation_run_batch_id,
			synthetic_calcutta_id,
			cohort_id,
			calcutta_id,
			game_outcome_run_id,
			market_share_run_id,
			optimizer_key,
			n_sims,
			seed,
			starting_state_key,
			excluded_entry_name
		)
		VALUES ($1::uuid, $2::uuid, $3::uuid, $4::uuid, $5::uuid, $6::uuid, $7, $8::int, $9::int, $10, $11::text)
		RETURNING id, status
	`
	var execID any
	if simulationBatchID != "" {
		execID = simulationBatchID
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
	if err := s.pool.QueryRow(ctx, q, execID, syntheticCalcuttaID, cohortID, req.CalcuttaID, goRun, msRun, evalOptimizerKey, effNSims, effSeed, req.StartingStateKey, excluded).Scan(&evalID, &status); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, createSuiteCalcuttaEvaluationResponse{ID: evalID, Status: status})
}

func (s *Server) listSuiteCalcuttaEvaluationsHandler(w http.ResponseWriter, r *http.Request) {
	calcuttaID := r.URL.Query().Get("calcutta_id")
	cohortID := r.URL.Query().Get("cohort_id")
	simulationBatchID := r.URL.Query().Get("simulation_batch_id")
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

	items, err := s.loadSuiteCalcuttaEvaluations(r.Context(), calcuttaID, cohortID, simulationBatchID, limit, offset)
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
		FROM derived.strategy_generation_run_bids reb
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
			WITH focus AS (
				SELECT se.display_name
				FROM derived.simulation_runs sr
				JOIN core.calcutta_snapshot_entries se
					ON se.id = sr.focus_snapshot_entry_id
					AND se.deleted_at IS NULL
				WHERE sr.id = $2::uuid
					AND sr.deleted_at IS NULL
				LIMIT 1
			),
			ranked AS (
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
			WHERE r.entry_name = (SELECT display_name FROM focus)
			ORDER BY r.rank ASC
			LIMIT 1
		`, *eval.CalcuttaEvaluationRunID, eval.ID).Scan(
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
