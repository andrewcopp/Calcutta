package httpserver

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/dtos"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
)

type createSuiteCalcuttaEvaluationResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type suiteCalcuttaEvaluationListItem struct {
	ID                      string     `json:"id"`
	SuiteID                 string     `json:"suite_id"`
	SuiteName               string     `json:"suite_name"`
	OptimizerKey            string     `json:"optimizer_key"`
	NSims                   int        `json:"n_sims"`
	Seed                    int        `json:"seed"`
	CalcuttaID              string     `json:"calcutta_id"`
	GameOutcomeRunID        *string    `json:"game_outcome_run_id,omitempty"`
	MarketShareRunID        *string    `json:"market_share_run_id,omitempty"`
	StrategyGenerationRunID *string    `json:"strategy_generation_run_id,omitempty"`
	CalcuttaEvaluationRunID *string    `json:"calcutta_evaluation_run_id,omitempty"`
	StartingStateKey        string     `json:"starting_state_key"`
	ExcludedEntryName       *string    `json:"excluded_entry_name,omitempty"`
	Status                  string     `json:"status"`
	ClaimedAt               *time.Time `json:"claimed_at,omitempty"`
	ClaimedBy               *string    `json:"claimed_by,omitempty"`
	ErrorMessage            *string    `json:"error_message,omitempty"`
	CreatedAt               time.Time  `json:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at"`
}

type suiteCalcuttaEvaluationListResponse struct {
	Items []suiteCalcuttaEvaluationListItem `json:"items"`
}

type suiteCalcuttaEvaluationPortfolioBid struct {
	TeamID      string  `json:"team_id"`
	SchoolName  string  `json:"school_name"`
	Seed        int     `json:"seed"`
	Region      string  `json:"region"`
	BidPoints   int     `json:"bid_points"`
	ExpectedROI float64 `json:"expected_roi"`
}

type suiteCalcuttaEvaluationOurStrategyPerformance struct {
	Rank                   int     `json:"rank"`
	EntryName              string  `json:"entry_name"`
	MeanNormalizedPayout   float64 `json:"mean_normalized_payout"`
	MedianNormalizedPayout float64 `json:"median_normalized_payout"`
	PTop1                  float64 `json:"p_top1"`
	PInMoney               float64 `json:"p_in_money"`
	TotalSimulations       int     `json:"total_simulations"`
}

type suiteCalcuttaEvaluationResultResponse struct {
	Evaluation  suiteCalcuttaEvaluationListItem                `json:"evaluation"`
	Portfolio   []suiteCalcuttaEvaluationPortfolioBid          `json:"portfolio"`
	OurStrategy *suiteCalcuttaEvaluationOurStrategyPerformance `json:"our_strategy,omitempty"`
}

func (s *Server) registerSuiteCalcuttaEvaluationRoutes(r *mux.Router) {
	r.HandleFunc(
		"/api/suite-calcutta-evaluations",
		s.requirePermission("analytics.suite_calcutta_evaluations.write", s.createSuiteCalcuttaEvaluationHandler),
	).Methods("POST", "OPTIONS")
	r.HandleFunc(
		"/api/suite-calcutta-evaluations",
		s.requirePermission("analytics.suite_calcutta_evaluations.read", s.listSuiteCalcuttaEvaluationsHandler),
	).Methods("GET", "OPTIONS")
	r.HandleFunc(
		"/api/suite-calcutta-evaluations/{id}",
		s.requirePermission("analytics.suite_calcutta_evaluations.read", s.getSuiteCalcuttaEvaluationHandler),
	).Methods("GET", "OPTIONS")
	r.HandleFunc(
		"/api/suite-calcutta-evaluations/{id}/result",
		s.requirePermission("analytics.suite_calcutta_evaluations.read", s.getSuiteCalcuttaEvaluationResultHandler),
	).Methods("GET", "OPTIONS")
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

	// Resolve (or create) suite_id.
	suiteID := ""
	if req.SuiteID != nil {
		suiteID = *req.SuiteID
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
			suite_id,
			calcutta_id,
			game_outcome_run_id,
			market_share_run_id,
			starting_state_key,
			excluded_entry_name
		)
		VALUES ($1::uuid, $2::uuid, $3::uuid, $4::uuid, $5, $6::text)
		RETURNING id, status
	`
	if err := s.pool.QueryRow(ctx, q, suiteID, req.CalcuttaID, goRun, msRun, req.StartingStateKey, excluded).Scan(&evalID, &status); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, createSuiteCalcuttaEvaluationResponse{ID: evalID, Status: status})
}

func (s *Server) listSuiteCalcuttaEvaluationsHandler(w http.ResponseWriter, r *http.Request) {
	calcuttaID := r.URL.Query().Get("calcutta_id")
	suiteID := r.URL.Query().Get("suite_id")
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

	rows, err := s.pool.Query(r.Context(), `
		SELECT
			r.id,
			r.suite_id,
			COALESCE(s.name, '') AS suite_name,
			COALESCE(s.optimizer_key, '') AS optimizer_key,
			COALESCE(s.n_sims, 0) AS n_sims,
			COALESCE(s.seed, 0) AS seed,
			r.calcutta_id,
			r.game_outcome_run_id,
			r.market_share_run_id,
			r.strategy_generation_run_id,
			r.calcutta_evaluation_run_id,
			r.starting_state_key,
			r.excluded_entry_name,
			r.status,
			r.claimed_at,
			r.claimed_by,
			r.error_message,
			r.created_at,
			r.updated_at
		FROM derived.suite_calcutta_evaluations r
		LEFT JOIN derived.suites s
			ON s.id = r.suite_id
			AND s.deleted_at IS NULL
		WHERE r.deleted_at IS NULL
			AND ($1::uuid IS NULL OR r.calcutta_id = $1::uuid)
			AND ($2::uuid IS NULL OR r.suite_id = $2::uuid)
		ORDER BY r.created_at DESC
		LIMIT $3::int
		OFFSET $4::int
	`, nullUUIDParam(calcuttaID), nullUUIDParam(suiteID), limit, offset)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	defer rows.Close()

	items := make([]suiteCalcuttaEvaluationListItem, 0)
	for rows.Next() {
		var it suiteCalcuttaEvaluationListItem
		if err := rows.Scan(
			&it.ID,
			&it.SuiteID,
			&it.SuiteName,
			&it.OptimizerKey,
			&it.NSims,
			&it.Seed,
			&it.CalcuttaID,
			&it.GameOutcomeRunID,
			&it.MarketShareRunID,
			&it.StrategyGenerationRunID,
			&it.CalcuttaEvaluationRunID,
			&it.StartingStateKey,
			&it.ExcludedEntryName,
			&it.Status,
			&it.ClaimedAt,
			&it.ClaimedBy,
			&it.ErrorMessage,
			&it.CreatedAt,
			&it.UpdatedAt,
		); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
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

	var it suiteCalcuttaEvaluationListItem
	err := s.pool.QueryRow(r.Context(), `
		SELECT
			r.id,
			r.suite_id,
			COALESCE(s.name, '') AS suite_name,
			COALESCE(s.optimizer_key, '') AS optimizer_key,
			COALESCE(s.n_sims, 0) AS n_sims,
			COALESCE(s.seed, 0) AS seed,
			r.calcutta_id,
			r.game_outcome_run_id,
			r.market_share_run_id,
			r.strategy_generation_run_id,
			r.calcutta_evaluation_run_id,
			r.starting_state_key,
			r.excluded_entry_name,
			r.status,
			r.claimed_at,
			r.claimed_by,
			r.error_message,
			r.created_at,
			r.updated_at
		FROM derived.suite_calcutta_evaluations r
		LEFT JOIN derived.suites s
			ON s.id = r.suite_id
			AND s.deleted_at IS NULL
		WHERE r.id = $1::uuid
			AND r.deleted_at IS NULL
		LIMIT 1
	`, id).Scan(
		&it.ID,
		&it.SuiteID,
		&it.SuiteName,
		&it.OptimizerKey,
		&it.NSims,
		&it.Seed,
		&it.CalcuttaID,
		&it.GameOutcomeRunID,
		&it.MarketShareRunID,
		&it.StrategyGenerationRunID,
		&it.CalcuttaEvaluationRunID,
		&it.StartingStateKey,
		&it.ExcludedEntryName,
		&it.Status,
		&it.ClaimedAt,
		&it.ClaimedBy,
		&it.ErrorMessage,
		&it.CreatedAt,
		&it.UpdatedAt,
	)
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

	var eval suiteCalcuttaEvaluationListItem
	err := s.pool.QueryRow(ctx, `
		SELECT
			r.id,
			r.suite_id,
			COALESCE(s.name, '') AS suite_name,
			COALESCE(s.optimizer_key, '') AS optimizer_key,
			COALESCE(s.n_sims, 0) AS n_sims,
			COALESCE(s.seed, 0) AS seed,
			r.calcutta_id,
			r.game_outcome_run_id,
			r.market_share_run_id,
			r.strategy_generation_run_id,
			r.calcutta_evaluation_run_id,
			r.starting_state_key,
			r.excluded_entry_name,
			r.status,
			r.claimed_at,
			r.claimed_by,
			r.error_message,
			r.created_at,
			r.updated_at
		FROM derived.suite_calcutta_evaluations r
		LEFT JOIN derived.suites s
			ON s.id = r.suite_id
			AND s.deleted_at IS NULL
		WHERE r.id = $1::uuid
			AND r.deleted_at IS NULL
		LIMIT 1
	`, id).Scan(
		&eval.ID,
		&eval.SuiteID,
		&eval.SuiteName,
		&eval.OptimizerKey,
		&eval.NSims,
		&eval.Seed,
		&eval.CalcuttaID,
		&eval.GameOutcomeRunID,
		&eval.MarketShareRunID,
		&eval.StrategyGenerationRunID,
		&eval.CalcuttaEvaluationRunID,
		&eval.StartingStateKey,
		&eval.ExcludedEntryName,
		&eval.Status,
		&eval.ClaimedAt,
		&eval.ClaimedBy,
		&eval.ErrorMessage,
		&eval.CreatedAt,
		&eval.UpdatedAt,
	)
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

	writeJSON(w, http.StatusOK, suiteCalcuttaEvaluationResultResponse{
		Evaluation:  eval,
		Portfolio:   portfolio,
		OurStrategy: our,
	})
}

func nullUUIDParam(v string) any {
	if v == "" {
		return nil
	}
	return v
}
