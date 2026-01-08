package httpserver

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
)

type labEntriesCoverageItem struct {
	CohortID                 string `json:"cohort_id"`
	CohortName               string `json:"cohort_name"`
	AdvancementAlgorithmID   string `json:"advancement_algorithm_id"`
	AdvancementAlgorithmName string `json:"advancement_algorithm_name"`
	InvestmentAlgorithmID    string `json:"investment_algorithm_id"`
	InvestmentAlgorithmName  string `json:"investment_algorithm_name"`
	OptimizerKey             string `json:"optimizer_key"`
	Covered                  int    `json:"covered"`
	Total                    int    `json:"total"`
}

type labEntriesCoverageResponse struct {
	Items []labEntriesCoverageItem `json:"items"`
}

type labEntriesSuiteScenarioItem struct {
	ScenarioID              string  `json:"scenario_id"`
	CalcuttaID              string  `json:"calcutta_id"`
	CalcuttaName            string  `json:"calcutta_name"`
	TournamentName          string  `json:"tournament_name"`
	Season                  string  `json:"season"`
	Picks                   string  `json:"picks"`
	EntryCreatedAt          *string `json:"entry_created_at,omitempty"`
	ScenarioCreatedAt       string  `json:"scenario_created_at"`
	StrategyGenerationRunID *string `json:"strategy_generation_run_id,omitempty"`
}

func formatLabEntryPicks(maxTeams int, seeds []int32) string {
	if maxTeams <= 0 {
		return ""
	}

	if len(seeds) > 1 {
		sort.Slice(seeds, func(i, j int) bool { return seeds[i] < seeds[j] })
	}

	parts := make([]string, 0, maxTeams)
	for i := 0; i < maxTeams; i++ {
		if i < len(seeds) {
			parts = append(parts, strconv.Itoa(int(seeds[i])))
		} else {
			parts = append(parts, "-")
		}
	}
	return strings.Join(parts, ",")
}

type labEntriesSuiteDetailResponse struct {
	Cohort struct {
		ID                   string `json:"id"`
		Name                 string `json:"name"`
		AdvancementAlgorithm struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"advancement_algorithm"`
		InvestmentAlgorithm struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"investment_algorithm"`
		OptimizerKey      string  `json:"optimizer_key"`
		StartingStateKey  string  `json:"starting_state_key"`
		ExcludedEntryName *string `json:"excluded_entry_name,omitempty"`
	} `json:"cohort"`
	Items []labEntriesSuiteScenarioItem `json:"items"`
}

type labEntryReportTeam struct {
	TeamID         string  `json:"team_id"`
	SchoolName     string  `json:"school_name"`
	Seed           int     `json:"seed"`
	Region         string  `json:"region"`
	ExpectedPoints float64 `json:"expected_points"`
	ExpectedMarket float64 `json:"expected_market"`
	PredictedROI   float64 `json:"predicted_roi"`
	OurBid         float64 `json:"our_bid"`
	ObservedROI    float64 `json:"observed_roi"`
}

type labEntryReportScoringRule struct {
	WinIndex      int `json:"win_index"`
	PointsAwarded int `json:"points_awarded"`
}

type labEntryReportResponse struct {
	ScenarioID               string                      `json:"scenario_id"`
	CohortID                 string                      `json:"cohort_id"`
	CalcuttaID               string                      `json:"calcutta_id"`
	CalcuttaName             string                      `json:"calcutta_name"`
	Season                   string                      `json:"season"`
	TournamentName           string                      `json:"tournament_name"`
	AdvancementAlgorithmID   string                      `json:"advancement_algorithm_id"`
	AdvancementAlgorithmName string                      `json:"advancement_algorithm_name"`
	InvestmentAlgorithmID    string                      `json:"investment_algorithm_id"`
	InvestmentAlgorithmName  string                      `json:"investment_algorithm_name"`
	OptimizerKey             string                      `json:"optimizer_key"`
	StrategyGenerationRunID  *string                     `json:"strategy_generation_run_id,omitempty"`
	GameOutcomeRunID         *string                     `json:"game_outcome_run_id,omitempty"`
	MarketShareRunID         *string                     `json:"market_share_run_id,omitempty"`
	BudgetPoints             int                         `json:"budget_points"`
	MinTeams                 int                         `json:"min_teams"`
	MaxTeams                 int                         `json:"max_teams"`
	MaxBidPoints             int                         `json:"max_bid_points"`
	AssumedEntries           int                         `json:"assumed_entries"`
	ExcludedEntryName        *string                     `json:"excluded_entry_name,omitempty"`
	ScoringRules             []labEntryReportScoringRule `json:"scoring_rules"`
	Teams                    []labEntryReportTeam        `json:"teams"`
}

type createLabSuiteSandboxExecutionResponse struct {
	ExecutionID     string `json:"executionId"`
	EvaluationCount int    `json:"evaluationCount"`
}

type generateLabEntriesFailure struct {
	ScenarioID string `json:"scenario_id"`
	CalcuttaID string `json:"calcutta_id"`
	Message    string `json:"message"`
}

type generateLabEntriesResponse struct {
	Created  int                         `json:"created"`
	Skipped  int                         `json:"skipped"`
	Failed   int                         `json:"failed"`
	Failures []generateLabEntriesFailure `json:"failures"`
}

type syncLabEntriesAutoCohortsResponse struct {
	Upserted int `json:"upserted"`
}

func (s *Server) registerLabEntriesRoutes(r *mux.Router) {
	r.HandleFunc(
		"/api/lab/entries",
		s.requirePermission("analytics.suites.read", s.listLabEntriesCoverageHandler),
	).Methods("GET", "OPTIONS")
	r.HandleFunc(
		"/api/lab/entries/auto-cohorts/sync",
		s.requirePermission("analytics.suites.write", s.syncLabEntriesAutoCohortsHandler),
	).Methods("POST", "OPTIONS")
	r.HandleFunc(
		"/api/lab/entries/cohorts/{id}",
		s.requirePermission("analytics.suites.read", s.getLabEntriesCohortDetailHandler),
	).Methods("GET", "OPTIONS")
	r.HandleFunc(
		"/api/lab/entries/scenarios/{id}",
		s.requirePermission("analytics.suites.read", s.getLabEntryReportHandler),
	).Methods("GET", "OPTIONS")
	r.HandleFunc(
		"/api/lab/entries/cohorts/{id}/sandbox-executions",
		s.requirePermission("analytics.suite_executions.write", s.createLabCohortSandboxExecutionHandler),
	).Methods("POST", "OPTIONS")
	r.HandleFunc(
		"/api/lab/entries/cohorts/{id}/generate-entries",
		s.requirePermission("analytics.strategy_generation_runs.write", s.generateLabEntriesForCohortHandler),
	).Methods("POST", "OPTIONS")
}

func (s *Server) generateLabEntriesForCohortHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cohortID := strings.TrimSpace(vars["id"])
	if cohortID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "id is required", "id")
		return
	}
	if _, err := uuid.Parse(cohortID); err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "id must be a valid UUID", "id")
		return
	}

	ctx := r.Context()

	var goAlgID string
	var msAlgID string
	var optimizerKey string
	var startingStateKey string
	var cohortExcludedEntryName *string
	if err := s.pool.QueryRow(ctx, `
		SELECT
			game_outcomes_algorithm_id::text,
			market_share_algorithm_id::text,
			COALESCE(optimizer_key, ''::text) AS optimizer_key,
			COALESCE(NULLIF(starting_state_key, ''), 'post_first_four') AS starting_state_key,
			excluded_entry_name
		FROM derived.synthetic_calcutta_cohorts
		WHERE id = $1::uuid
			AND deleted_at IS NULL
		LIMIT 1
	`, cohortID).Scan(&goAlgID, &msAlgID, &optimizerKey, &startingStateKey, &cohortExcludedEntryName); err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, r, http.StatusNotFound, "not_found", "Cohort not found", "id")
			return
		}
		writeErrorFromErr(w, r, err)
		return
	}
	optimizerKey = strings.TrimSpace(optimizerKey)
	if optimizerKey == "" {
		optimizerKey = "minlp_v1"
	}
	startingStateKey = strings.TrimSpace(startingStateKey)
	if startingStateKey == "" {
		startingStateKey = "post_first_four"
	}

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

	// Ensure this cohort has scenarios for all calcuttas. (AUTO cohorts start empty.)
	// We preserve any existing scenario-level overrides by only filling in NULL/empty fields.
	if _, err := tx.Exec(ctx, `
		INSERT INTO derived.synthetic_calcuttas (
			cohort_id,
			calcutta_id,
			starting_state_key,
			excluded_entry_name
		)
		SELECT
			$1::uuid,
			c.id,
			$2,
			$3
		FROM core.calcuttas c
		WHERE c.deleted_at IS NULL
			AND NOT EXISTS (
				SELECT 1
				FROM derived.synthetic_calcuttas sc
				WHERE sc.cohort_id = $1::uuid
					AND sc.calcutta_id = c.id
					AND sc.deleted_at IS NULL
			)
	`, cohortID, startingStateKey, cohortExcludedEntryName); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	if _, err := tx.Exec(ctx, `
		UPDATE derived.synthetic_calcuttas
		SET starting_state_key = COALESCE(NULLIF(starting_state_key, ''), $2),
			excluded_entry_name = COALESCE(NULLIF(excluded_entry_name, ''), $3),
			updated_at = NOW()
		WHERE cohort_id = $1::uuid
			AND deleted_at IS NULL
	`, cohortID, startingStateKey, cohortExcludedEntryName); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	type scenarioRow struct {
		ScenarioID string
		CalcuttaID string
		Excluded   *string
		FocusRunID *string
	}

	rows, err := tx.Query(ctx, `
		SELECT
			id::text,
			calcutta_id::text,
			excluded_entry_name,
			focus_strategy_generation_run_id::text
		FROM derived.synthetic_calcuttas
		WHERE cohort_id = $1::uuid
			AND deleted_at IS NULL
		ORDER BY created_at ASC
	`, cohortID)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	scenarios := make([]scenarioRow, 0)
	for rows.Next() {
		var sc scenarioRow
		if err := rows.Scan(&sc.ScenarioID, &sc.CalcuttaID, &sc.Excluded, &sc.FocusRunID); err != nil {
			rows.Close()
			writeErrorFromErr(w, r, err)
			return
		}
		scenarios = append(scenarios, sc)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		writeErrorFromErr(w, r, err)
		return
	}
	rows.Close()

	resp := generateLabEntriesResponse{Failures: make([]generateLabEntriesFailure, 0)}

	for _, sc := range scenarios {

		if sc.FocusRunID != nil && strings.TrimSpace(*sc.FocusRunID) != "" {
			resp.Skipped++
			continue
		}

		effExcluded := sc.Excluded
		if effExcluded == nil {
			effExcluded = cohortExcludedEntryName
		}

		var msRunID string
		var msArtifactID string
		if err := tx.QueryRow(ctx, `
			SELECT r.id::text, a.id::text
			FROM derived.market_share_runs r
			JOIN derived.run_artifacts a
				ON a.run_kind = 'market_share'
				AND a.run_id = r.id
				AND a.artifact_kind = 'metrics'
				AND a.deleted_at IS NULL
			WHERE r.calcutta_id = $1::uuid
				AND r.algorithm_id = $2::uuid
				AND r.deleted_at IS NULL
			ORDER BY (CASE WHEN $3::text IS NOT NULL AND r.params_json->>'excluded_entry_name' = $3::text THEN 1 ELSE 0 END) DESC,
				r.created_at DESC
			LIMIT 1
		`, sc.CalcuttaID, msAlgID, effExcluded).Scan(&msRunID, &msArtifactID); err != nil {
			if err == pgx.ErrNoRows {
				resp.Failed++
				resp.Failures = append(resp.Failures, generateLabEntriesFailure{ScenarioID: sc.ScenarioID, CalcuttaID: sc.CalcuttaID, Message: "Missing market-share run (with metrics artifact) for cohort algorithm"})
				continue
			}
			writeErrorFromErr(w, r, err)
			return
		}

		var tournamentID string
		if err := tx.QueryRow(ctx, `
			SELECT tournament_id::text
			FROM core.calcuttas
			WHERE id = $1::uuid
				AND deleted_at IS NULL
			LIMIT 1
		`, sc.CalcuttaID).Scan(&tournamentID); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}

		var goRunID string
		if err := tx.QueryRow(ctx, `
			SELECT gor.id::text
			FROM derived.game_outcome_runs gor
			WHERE gor.tournament_id = $1::uuid
				AND gor.algorithm_id = $2::uuid
				AND gor.deleted_at IS NULL
			ORDER BY gor.created_at DESC
			LIMIT 1
		`, tournamentID, goAlgID).Scan(&goRunID); err != nil {
			if err == pgx.ErrNoRows {
				resp.Failed++
				resp.Failures = append(resp.Failures, generateLabEntriesFailure{ScenarioID: sc.ScenarioID, CalcuttaID: sc.CalcuttaID, Message: "Missing game-outcome run for cohort algorithm"})
				continue
			}
			writeErrorFromErr(w, r, err)
			return
		}

		runKeyUUID := uuid.NewString()
		runKeyText := runKeyUUID
		name := fmt.Sprintf("lab_entries_%s", optimizerKey)
		params := map[string]any{"market_share_run_id": msRunID, "market_share_artifact_id": msArtifactID, "game_outcome_run_id": goRunID, "excluded_entry_name": effExcluded, "source": "lab_entries_generate"}
		paramsJSON, _ := json.Marshal(params)

		gitSHA := strings.TrimSpace(os.Getenv("GIT_SHA"))
		var gitSHAParam any
		if gitSHA != "" {
			gitSHAParam = gitSHA
		} else {
			gitSHAParam = nil
		}

		var runID string
		if err := tx.QueryRow(ctx, `
			INSERT INTO derived.strategy_generation_runs (
				run_key,
				run_key_uuid,
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
				git_sha
			)
			VALUES ($1, $2::uuid, $3, NULL, $4::uuid, 'lab_entries_generation', 'pgo_dp', 'predicted_market_share', $5, $6::uuid, $7::uuid, $8::text, $9::text, $10::jsonb, $11)
			RETURNING id::text
		`, runKeyText, runKeyUUID, name, sc.CalcuttaID, optimizerKey, msRunID, goRunID, effExcluded, startingStateKey, string(paramsJSON), gitSHAParam).Scan(&runID); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}

		if _, err := tx.Exec(ctx, `
			UPDATE derived.synthetic_calcuttas
			SET focus_strategy_generation_run_id = $2::uuid,
				focus_entry_name = COALESCE(focus_entry_name, 'Our Strategy'),
				updated_at = NOW()
			WHERE id = $1::uuid
				AND deleted_at IS NULL
		`, sc.ScenarioID, runID); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}

		resp.Created++
	}

	if err := tx.Commit(ctx); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	committed = true

	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) createLabCohortSandboxExecutionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cohortID := strings.TrimSpace(vars["id"])
	if cohortID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "id is required", "id")
		return
	}
	if _, err := uuid.Parse(cohortID); err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "id must be a valid UUID", "id")
		return
	}

	type createLabCohortSandboxExecutionRequest struct {
		NSims             *int    `json:"nSims"`
		ExcludedEntryID   *string `json:"excludedEntryId"`
		ExcludedEntryName *string `json:"excludedEntryName"`
	}

	var req createLabCohortSandboxExecutionRequest
	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err != io.EOF {
			writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
			return
		}
	}
	if req.NSims != nil && *req.NSims < 0 {
		writeError(w, r, http.StatusBadRequest, "validation_error", "nSims must be >= 0", "nSims")
		return
	}
	if req.ExcludedEntryID != nil {
		v := strings.TrimSpace(*req.ExcludedEntryID)
		if v != "" {
			if _, err := uuid.Parse(v); err != nil {
				writeError(w, r, http.StatusBadRequest, "validation_error", "excludedEntryId must be a valid UUID", "excludedEntryId")
				return
			}
		}
		*req.ExcludedEntryID = v
	}
	if req.ExcludedEntryName != nil {
		v := strings.TrimSpace(*req.ExcludedEntryName)
		*req.ExcludedEntryName = v
	}

	ctx := r.Context()

	// Load suite-level defaults.
	var goAlgID string
	var msAlgID string
	var optimizerKey string
	var nSims int
	var seed int
	var startingStateKey string
	var suiteExcludedEntryName *string
	if err := s.pool.QueryRow(ctx, `
		SELECT
			game_outcomes_algorithm_id::text,
			market_share_algorithm_id::text,
			COALESCE(optimizer_key, ''::text) AS optimizer_key,
			n_sims,
			seed,
			COALESCE(NULLIF(starting_state_key, ''), 'post_first_four') AS starting_state_key,
			excluded_entry_name
		FROM derived.synthetic_calcutta_cohorts
		WHERE id = $1::uuid
			AND deleted_at IS NULL
		LIMIT 1
	`, cohortID).Scan(&goAlgID, &msAlgID, &optimizerKey, &nSims, &seed, &startingStateKey, &suiteExcludedEntryName); err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, r, http.StatusNotFound, "not_found", "Cohort not found", "id")
			return
		}
		writeErrorFromErr(w, r, err)
		return
	}

	if req.NSims != nil {
		nSims = *req.NSims
	}

	// Allow override by entry ID (resolve to name) or explicit name.
	if req.ExcludedEntryID != nil && strings.TrimSpace(*req.ExcludedEntryID) != "" {
		var excludedName string
		if err := s.pool.QueryRow(ctx, `
			SELECT ce.name
			FROM core.entries ce
			JOIN derived.synthetic_calcuttas sc
				ON sc.calcutta_id = ce.calcutta_id
				AND sc.cohort_id = $2::uuid
				AND sc.deleted_at IS NULL
			WHERE ce.id = $1::uuid
				AND ce.deleted_at IS NULL
			LIMIT 1
		`, *req.ExcludedEntryID, cohortID).Scan(&excludedName); err != nil {
			if err == pgx.ErrNoRows {
				writeError(w, r, http.StatusBadRequest, "validation_error", "excludedEntryId not found for this cohort", "excludedEntryId")
				return
			}
			writeErrorFromErr(w, r, err)
			return
		}
		excludedName = strings.TrimSpace(excludedName)
		if excludedName == "" {
			suiteExcludedEntryName = nil
		} else {
			suiteExcludedEntryName = &excludedName
		}
	} else if req.ExcludedEntryName != nil {
		if strings.TrimSpace(*req.ExcludedEntryName) == "" {
			suiteExcludedEntryName = nil
		} else {
			name := strings.TrimSpace(*req.ExcludedEntryName)
			suiteExcludedEntryName = &name
		}
	}

	// Load focused scenarios. We will resolve game_outcome_run_id + market_share_run_id
	// based on suite algorithms + each scenario's tournament/calcutta.
	type focusedScenarioRow struct {
		SyntheticCalcuttaID string
		CalcuttaID          string
		StrategyGenRunID    string
		ExcludedEntryName   *string
		StartingStateKey    string
	}

	rows, err := s.pool.Query(ctx, `
		SELECT
			sc.id::text,
			sc.calcutta_id::text,
			sc.focus_strategy_generation_run_id::text,
			sc.excluded_entry_name,
			COALESCE(NULLIF(sc.starting_state_key, ''), ''::text) AS starting_state_key
		FROM derived.synthetic_calcuttas sc
		WHERE sc.cohort_id = $1::uuid
			AND sc.deleted_at IS NULL
			AND sc.focus_strategy_generation_run_id IS NOT NULL
		ORDER BY sc.created_at ASC
	`, cohortID)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	defer rows.Close()

	focused := make([]focusedScenarioRow, 0)
	for rows.Next() {
		var fr focusedScenarioRow
		if err := rows.Scan(
			&fr.SyntheticCalcuttaID,
			&fr.CalcuttaID,
			&fr.StrategyGenRunID,
			&fr.ExcludedEntryName,
			&fr.StartingStateKey,
		); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
		focused = append(focused, fr)
	}
	if err := rows.Err(); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	if len(focused) == 0 {
		writeError(w, r, http.StatusConflict, "missing_focus", "No focused scenarios found for suite", "id")
		return
	}

	// Create suite_execution + evaluations transactionally.
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

	// Use suite defaults for execution; the evaluation rows can override excluded_entry_name per scenario.
	var executionID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO derived.simulation_run_batches (
			cohort_id,
			name,
			optimizer_key,
			n_sims,
			seed,
			starting_state_key,
			excluded_entry_name,
			status
		)
		VALUES ($1::uuid, NULL, $2, $3::int, $4::int, $5, $6::text, 'running')
		RETURNING id::text
	`, cohortID, optimizerKey, nSims, seed, startingStateKey, suiteExcludedEntryName).Scan(&executionID); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	for _, fr := range focused {
		var tournamentID string
		if err := tx.QueryRow(ctx, `
			SELECT tournament_id::text
			FROM core.calcuttas
			WHERE id = $1::uuid
				AND deleted_at IS NULL
			LIMIT 1
		`, fr.CalcuttaID).Scan(&tournamentID); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}

		var syntheticCalcuttaID string
		if err := tx.QueryRow(ctx, `
			SELECT id::text
			FROM derived.synthetic_calcuttas
			WHERE cohort_id = $1::uuid
				AND calcutta_id = $2::uuid
				AND deleted_at IS NULL
			LIMIT 1
		`, cohortID, fr.CalcuttaID).Scan(&syntheticCalcuttaID); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}

		var goRunID string
		if err := tx.QueryRow(ctx, `
			SELECT id::text
			FROM derived.game_outcome_runs
			WHERE tournament_id = $1::uuid
				AND algorithm_id = $2::uuid
				AND deleted_at IS NULL
			ORDER BY created_at DESC
			LIMIT 1
		`, tournamentID, goAlgID).Scan(&goRunID); err != nil {
			if err == pgx.ErrNoRows {
				writeError(w, r, http.StatusConflict, "missing_run", "Missing game-outcome run for suite algorithm", "gameOutcomeRunId")
				return
			}
			writeErrorFromErr(w, r, err)
			return
		}

		var msRunID string
		effExcluded := fr.ExcludedEntryName
		if effExcluded == nil {
			effExcluded = suiteExcludedEntryName
		}
		if err := tx.QueryRow(ctx, `
			SELECT id::text
			FROM derived.market_share_runs
			WHERE calcutta_id = $1::uuid
				AND algorithm_id = $2::uuid
				AND deleted_at IS NULL
			ORDER BY (CASE WHEN $3::text IS NOT NULL AND params_json->>'excluded_entry_name' = $3::text THEN 1 ELSE 0 END) DESC,
				created_at DESC
			LIMIT 1
		`, fr.CalcuttaID, msAlgID, effExcluded).Scan(&msRunID); err != nil {
			if err == pgx.ErrNoRows {
				writeError(w, r, http.StatusConflict, "missing_run", "Missing market-share run for suite algorithm", "marketShareRunId")
				return
			}
			writeErrorFromErr(w, r, err)
			return
		}

		effStarting := startingStateKey
		if strings.TrimSpace(fr.StartingStateKey) != "" {
			effStarting = fr.StartingStateKey
		}

		_, err := tx.Exec(ctx, `
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
				excluded_entry_name,
				strategy_generation_run_id
			)
			VALUES ($1::uuid, $2::uuid, $3::uuid, $4::uuid, $5::uuid, $6::uuid, $7, $8::int, $9::int, $10, $11::text, $12::uuid)
		`, executionID, syntheticCalcuttaID, cohortID, fr.CalcuttaID, goRunID, msRunID, optimizerKey, nSims, seed, effStarting, effExcluded, fr.StrategyGenRunID)
		if err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	committed = true

	writeJSON(w, http.StatusCreated, createLabSuiteSandboxExecutionResponse{ExecutionID: executionID, EvaluationCount: len(focused)})
}

func (s *Server) listLabEntriesCoverageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{AccessMode: pgx.ReadOnly})
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	rows, err := tx.Query(ctx, `
		WITH scenario_counts AS (
			SELECT
				cohort_id,
				COUNT(*)::int AS total,
				COUNT(*) FILTER (WHERE focus_strategy_generation_run_id IS NOT NULL)::int AS covered
			FROM derived.synthetic_calcuttas
			WHERE deleted_at IS NULL
			GROUP BY cohort_id
		)
		SELECT
			s.id::text,
			s.name,
			goa.id::text,
			goa.name,
			msa.id::text,
			msa.name,
			s.optimizer_key,
			COALESCE(sc.covered, 0)::int,
			COALESCE(sc.total, 0)::int
		FROM derived.synthetic_calcutta_cohorts s
		JOIN derived.algorithms goa
			ON goa.id = s.game_outcomes_algorithm_id
			AND goa.deleted_at IS NULL
		JOIN derived.algorithms msa
			ON msa.id = s.market_share_algorithm_id
			AND msa.deleted_at IS NULL
		LEFT JOIN scenario_counts sc
			ON sc.cohort_id = s.id
		WHERE s.deleted_at IS NULL
			AND COALESCE(s.params_json->>'auto', 'false') = 'true'
		ORDER BY s.created_at DESC
	`)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	defer rows.Close()

	items := make([]labEntriesCoverageItem, 0)
	for rows.Next() {
		var it labEntriesCoverageItem
		if err := rows.Scan(
			&it.CohortID,
			&it.CohortName,
			&it.AdvancementAlgorithmID,
			&it.AdvancementAlgorithmName,
			&it.InvestmentAlgorithmID,
			&it.InvestmentAlgorithmName,
			&it.OptimizerKey,
			&it.Covered,
			&it.Total,
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

	if err := tx.Commit(ctx); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, labEntriesCoverageResponse{Items: items})
}

func (s *Server) syncLabEntriesAutoCohortsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	type alg struct {
		ID   string
		Name string
	}

	goAlgs := make([]alg, 0)
	{
		rows, err := s.pool.Query(ctx, `
			SELECT DISTINCT
				a.id::text,
				a.name
			FROM derived.algorithms a
			JOIN derived.game_outcome_runs gor
				ON gor.algorithm_id = a.id
				AND gor.deleted_at IS NULL
			JOIN derived.run_artifacts ra
				ON ra.run_kind = 'game_outcome'
				AND ra.run_id = gor.id
				AND ra.deleted_at IS NULL
			WHERE a.kind = 'game_outcomes'
				AND a.deleted_at IS NULL
			ORDER BY a.name ASC
		`)
		if err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
		defer rows.Close()
		for rows.Next() {
			var a alg
			if err := rows.Scan(&a.ID, &a.Name); err != nil {
				writeErrorFromErr(w, r, err)
				return
			}
			goAlgs = append(goAlgs, a)
		}
		if err := rows.Err(); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
	}
	if len(goAlgs) == 0 {
		rows, err := s.pool.Query(ctx, `
			SELECT a.id::text, a.name
			FROM derived.algorithms a
			WHERE a.kind = 'game_outcomes'
				AND a.deleted_at IS NULL
			ORDER BY a.name ASC
		`)
		if err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
		defer rows.Close()
		for rows.Next() {
			var a alg
			if err := rows.Scan(&a.ID, &a.Name); err != nil {
				writeErrorFromErr(w, r, err)
				return
			}
			goAlgs = append(goAlgs, a)
		}
		if err := rows.Err(); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
	}

	msAlgs := make([]alg, 0)
	{
		rows, err := s.pool.Query(ctx, `
			SELECT DISTINCT
				a.id::text,
				a.name
			FROM derived.algorithms a
			JOIN derived.market_share_runs msr
				ON msr.algorithm_id = a.id
				AND msr.deleted_at IS NULL
			JOIN derived.run_artifacts ra
				ON ra.run_kind = 'market_share'
				AND ra.run_id = msr.id
				AND ra.deleted_at IS NULL
			WHERE a.kind = 'market_share'
				AND a.deleted_at IS NULL
			ORDER BY a.name ASC
		`)
		if err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
		defer rows.Close()
		for rows.Next() {
			var a alg
			if err := rows.Scan(&a.ID, &a.Name); err != nil {
				writeErrorFromErr(w, r, err)
				return
			}
			msAlgs = append(msAlgs, a)
		}
		if err := rows.Err(); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
	}
	if len(msAlgs) == 0 {
		rows, err := s.pool.Query(ctx, `
			SELECT a.id::text, a.name
			FROM derived.algorithms a
			WHERE a.kind = 'market_share'
				AND a.deleted_at IS NULL
			ORDER BY a.name ASC
		`)
		if err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
		defer rows.Close()
		for rows.Next() {
			var a alg
			if err := rows.Scan(&a.ID, &a.Name); err != nil {
				writeErrorFromErr(w, r, err)
				return
			}
			msAlgs = append(msAlgs, a)
		}
		if err := rows.Err(); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
	}

	optimizerKeys := make([]string, 0)
	if s.app != nil && s.app.ModelCatalogs != nil {
		desc := s.app.ModelCatalogs.ListEntryOptimizers()
		for _, d := range desc {
			if d.Deprecated {
				continue
			}
			if strings.TrimSpace(d.ID) == "" {
				continue
			}
			optimizerKeys = append(optimizerKeys, strings.TrimSpace(d.ID))
		}
	}
	if len(optimizerKeys) == 0 {
		optimizerKeys = append(optimizerKeys, "minlp_v1")
	}

	upserted := 0
	for _, goa := range goAlgs {
		for _, msa := range msAlgs {
			for _, opt := range optimizerKeys {
				name := fmt.Sprintf("AUTO: %s + %s + %s", goa.Name, msa.Name, opt)
				stableID := uuid.NewSHA1(uuid.NameSpaceURL, []byte(fmt.Sprintf("lab_entries|go=%s|ms=%s|opt=%s", goa.ID, msa.ID, opt))).String()
				tag, err := s.pool.Exec(ctx, `
					INSERT INTO derived.synthetic_calcutta_cohorts (
						id,
						name,
						description,
						game_outcomes_algorithm_id,
						market_share_algorithm_id,
						optimizer_key,
						n_sims,
						seed,
						starting_state_key,
						params_json
					)
					VALUES (
						$1::uuid,
						$2,
						NULL,
						$3::uuid,
						$4::uuid,
						$5,
						5000,
						42,
						'post_first_four',
						'{"auto": true}'::jsonb
					)
					ON CONFLICT (id)
					DO UPDATE SET
						name = EXCLUDED.name,
						description = EXCLUDED.description,
						game_outcomes_algorithm_id = EXCLUDED.game_outcomes_algorithm_id,
						market_share_algorithm_id = EXCLUDED.market_share_algorithm_id,
						optimizer_key = EXCLUDED.optimizer_key,
						n_sims = EXCLUDED.n_sims,
						seed = EXCLUDED.seed,
						starting_state_key = EXCLUDED.starting_state_key,
						params_json = EXCLUDED.params_json,
						updated_at = NOW(),
						deleted_at = NULL
				`, stableID, name, goa.ID, msa.ID, opt)
				if err != nil {
					writeErrorFromErr(w, r, err)
					return
				}
				upserted += int(tag.RowsAffected())
			}
		}
	}

	writeJSON(w, http.StatusOK, syncLabEntriesAutoCohortsResponse{Upserted: upserted})
}

func (s *Server) getLabEntriesCohortDetailHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cohortID := strings.TrimSpace(vars["id"])
	if cohortID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "id is required", "id")
		return
	}
	if _, err := uuid.Parse(cohortID); err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "id must be a valid UUID", "id")
		return
	}

	resp := labEntriesSuiteDetailResponse{}
	var excl *string
	if err := s.pool.QueryRow(r.Context(), `
		SELECT
			s.id::text,
			s.name,
			goa.id::text,
			goa.name,
			msa.id::text,
			msa.name,
			s.optimizer_key,
			COALESCE(NULLIF(s.starting_state_key, ''), 'post_first_four') AS starting_state_key,
			s.excluded_entry_name
		FROM derived.synthetic_calcutta_cohorts s
		JOIN derived.algorithms goa ON goa.id = s.game_outcomes_algorithm_id AND goa.deleted_at IS NULL
		JOIN derived.algorithms msa ON msa.id = s.market_share_algorithm_id AND msa.deleted_at IS NULL
		WHERE s.id = $1::uuid
			AND s.deleted_at IS NULL
		LIMIT 1
	`, cohortID).Scan(
		&resp.Cohort.ID,
		&resp.Cohort.Name,
		&resp.Cohort.AdvancementAlgorithm.ID,
		&resp.Cohort.AdvancementAlgorithm.Name,
		&resp.Cohort.InvestmentAlgorithm.ID,
		&resp.Cohort.InvestmentAlgorithm.Name,
		&resp.Cohort.OptimizerKey,
		&resp.Cohort.StartingStateKey,
		&excl,
	); err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, r, http.StatusNotFound, "not_found", "Cohort not found", "id")
			return
		}
		writeErrorFromErr(w, r, err)
		return
	}
	resp.Cohort.ExcludedEntryName = excl

	rows, err := s.pool.Query(r.Context(), `
		SELECT
			sc.id::text,
			sc.calcutta_id::text,
			c.name,
			t.name,
			seas.year::text,
			COALESCE(c.max_teams, 10)::int,
			(
				SELECT ARRAY_AGG(tt.seed ORDER BY tt.seed ASC, tt.id ASC)
				FROM derived.recommended_entry_bids reb
				JOIN core.teams tt ON tt.id = reb.team_id AND tt.deleted_at IS NULL
				WHERE reb.strategy_generation_run_id = sc.focus_strategy_generation_run_id
					AND reb.deleted_at IS NULL
					AND reb.bid_points > 0
			),
			sgr.created_at::text,
			sc.created_at::text,
			sc.focus_strategy_generation_run_id::text
		FROM derived.synthetic_calcuttas sc
		JOIN core.calcuttas c ON c.id = sc.calcutta_id AND c.deleted_at IS NULL
		JOIN core.tournaments t ON t.id = c.tournament_id AND t.deleted_at IS NULL
		JOIN core.seasons seas ON seas.id = t.season_id
		LEFT JOIN derived.strategy_generation_runs sgr
			ON sgr.id = sc.focus_strategy_generation_run_id
			AND sgr.deleted_at IS NULL
		WHERE sc.cohort_id = $1::uuid
			AND sc.deleted_at IS NULL
		ORDER BY seas.year DESC
	`, cohortID)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	defer rows.Close()

	items := make([]labEntriesSuiteScenarioItem, 0)
	for rows.Next() {
		var it labEntriesSuiteScenarioItem
		var maxTeams int
		var seeds []int32
		if err := rows.Scan(
			&it.ScenarioID,
			&it.CalcuttaID,
			&it.CalcuttaName,
			&it.TournamentName,
			&it.Season,
			&maxTeams,
			&seeds,
			&it.EntryCreatedAt,
			&it.ScenarioCreatedAt,
			&it.StrategyGenerationRunID,
		); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}

		if it.StrategyGenerationRunID == nil || strings.TrimSpace(*it.StrategyGenerationRunID) == "" {
			it.Picks = "—"
		} else {
			it.Picks = formatLabEntryPicks(maxTeams, seeds)
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	resp.Items = items
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) getLabEntryReportHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	scenarioID := strings.TrimSpace(vars["id"])
	if scenarioID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "id is required", "id")
		return
	}
	if _, err := uuid.Parse(scenarioID); err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "id must be a valid UUID", "id")
		return
	}

	// Load scenario + suite + calcutta context.
	var suiteID string
	var suiteName string
	var calcuttaID string
	var calcuttaName string
	var tournamentID string
	var tournamentName string
	var seasonYear string
	var advAlgID string
	var advAlgName string
	var invAlgID string
	var invAlgName string
	var optimizerKey string
	var strategyRunID *string
	var budgetPoints int
	var minTeams int
	var maxTeams int
	var maxBidPoints int
	var suiteExcluded *string
	var scenarioExcluded *string

	err := s.pool.QueryRow(r.Context(), `
		SELECT
			sc.id::text,
			sc.cohort_id::text,
			su.name,
			sc.calcutta_id::text,
			c.name,
			c.tournament_id::text,
			t.name,
			seas.year::text,
			su.game_outcomes_algorithm_id::text,
			goa.name,
			su.market_share_algorithm_id::text,
			msa.name,
			su.optimizer_key,
			sc.focus_strategy_generation_run_id::text,
			COALESCE(c.budget_points, 100)::int,
			COALESCE(c.min_teams, 3)::int,
			COALESCE(c.max_teams, 10)::int,
			COALESCE(c.max_bid, 50)::int,
			su.excluded_entry_name,
			sc.excluded_entry_name
		FROM derived.synthetic_calcuttas sc
		JOIN derived.synthetic_calcutta_cohorts su ON su.id = sc.cohort_id AND su.deleted_at IS NULL
		JOIN derived.algorithms goa ON goa.id = su.game_outcomes_algorithm_id AND goa.deleted_at IS NULL
		JOIN derived.algorithms msa ON msa.id = su.market_share_algorithm_id AND msa.deleted_at IS NULL
		JOIN core.calcuttas c ON c.id = sc.calcutta_id AND c.deleted_at IS NULL
		JOIN core.tournaments t ON t.id = c.tournament_id AND t.deleted_at IS NULL
		JOIN core.seasons seas ON seas.id = t.season_id
		WHERE sc.id = $1::uuid
			AND sc.deleted_at IS NULL
		LIMIT 1
	`, scenarioID).Scan(
		new(string),
		&suiteID,
		&suiteName,
		&calcuttaID,
		&calcuttaName,
		&tournamentID,
		&tournamentName,
		&seasonYear,
		&advAlgID,
		&advAlgName,
		&invAlgID,
		&invAlgName,
		&optimizerKey,
		&strategyRunID,
		&budgetPoints,
		&minTeams,
		&maxTeams,
		&maxBidPoints,
		&suiteExcluded,
		&scenarioExcluded,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, r, http.StatusNotFound, "not_found", "Suite scenario not found", "id")
			return
		}
		writeErrorFromErr(w, r, err)
		return
	}

	excluded := scenarioExcluded
	if excluded == nil {
		excluded = suiteExcluded
	}

	// Resolve latest matching prediction runs for this suite.
	var gameOutcomeRunID *string
	{
		var id string
		err := s.pool.QueryRow(r.Context(), `
			SELECT gor.id::text
			FROM derived.game_outcome_runs gor
			WHERE gor.tournament_id = $1::uuid
				AND gor.algorithm_id = $2::uuid
				AND gor.deleted_at IS NULL
			ORDER BY gor.created_at DESC
			LIMIT 1
		`, tournamentID, advAlgID).Scan(&id)
		if err != nil {
			if err != pgx.ErrNoRows {
				writeErrorFromErr(w, r, err)
				return
			}
		} else {
			gameOutcomeRunID = &id
		}
	}

	var marketShareRunID *string
	{
		var id string
		err := s.pool.QueryRow(r.Context(), `
			SELECT msr.id::text
			FROM derived.market_share_runs msr
			WHERE msr.calcutta_id = $1::uuid
				AND msr.algorithm_id = $2::uuid
				AND msr.deleted_at IS NULL
			ORDER BY (CASE WHEN $3::text IS NOT NULL AND msr.params_json->>'excluded_entry_name' = $3::text THEN 1 ELSE 0 END) DESC,
				msr.created_at DESC
			LIMIT 1
		`, calcuttaID, invAlgID, excluded).Scan(&id)
		if err != nil {
			if err != pgx.ErrNoRows {
				writeErrorFromErr(w, r, err)
				return
			}
		} else {
			marketShareRunID = &id
		}
	}

	// Assumed entries for pool sizing.
	assumedEntries := 47
	{
		var n int
		if err := s.pool.QueryRow(r.Context(), `
			SELECT COUNT(*)::int
			FROM core.entries e
			WHERE e.calcutta_id = $1::uuid
				AND e.deleted_at IS NULL
		`, calcuttaID).Scan(&n); err == nil {
			if n > 0 {
				assumedEntries = n
			}
		}
	}
	poolSize := float64(assumedEntries * budgetPoints)

	// Scoring rules.
	scoring := make([]labEntryReportScoringRule, 0)
	{
		rows, err := s.pool.Query(r.Context(), `
			SELECT win_index::int, points_awarded::int
			FROM core.calcutta_scoring_rules
			WHERE calcutta_id = $1::uuid
				AND deleted_at IS NULL
			ORDER BY win_index ASC
		`, calcuttaID)
		if err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
		defer rows.Close()
		for rows.Next() {
			var rr labEntryReportScoringRule
			if err := rows.Scan(&rr.WinIndex, &rr.PointsAwarded); err != nil {
				writeErrorFromErr(w, r, err)
				return
			}
			scoring = append(scoring, rr)
		}
		if err := rows.Err(); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
	}

	// Predicted returns + market share (artifact-driven, run-scoped).
	_, selectedGORunID, returns, err := s.app.Analytics.GetCalcuttaPredictedReturns(r.Context(), calcuttaID, nil, gameOutcomeRunID)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	selectedMSRunID, _, marketShare, err := s.app.Analytics.GetCalcuttaPredictedMarketShare(r.Context(), calcuttaID, marketShareRunID, gameOutcomeRunID)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	mktByTeam := make(map[string]float64, len(marketShare))
	for _, ms := range marketShare {
		mktByTeam[ms.TeamID] = ms.PredictedShare
	}

	bidByTeam := make(map[string]float64)
	if strategyRunID != nil && *strategyRunID != "" {
		rows, err := s.pool.Query(r.Context(), `
			SELECT team_id::text, bid_points::double precision
			FROM derived.recommended_entry_bids
			WHERE strategy_generation_run_id = $1::uuid
				AND deleted_at IS NULL
		`, *strategyRunID)
		if err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
		defer rows.Close()
		for rows.Next() {
			var tid string
			var bid float64
			if err := rows.Scan(&tid, &bid); err != nil {
				writeErrorFromErr(w, r, err)
				return
			}
			bidByTeam[tid] = bid
		}
		if err := rows.Err(); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
	}

	teams := make([]labEntryReportTeam, 0, len(returns))
	for _, tr := range returns {
		share := mktByTeam[tr.TeamID]
		expectedMarket := share * poolSize
		predROI := 0.0
		if expectedMarket > 0 {
			predROI = tr.ExpectedValue / expectedMarket
		}
		ourBid := bidByTeam[tr.TeamID]
		observedROI := 0.0
		den := expectedMarket + ourBid
		if den > 0 {
			observedROI = tr.ExpectedValue / den
		}
		teams = append(teams, labEntryReportTeam{
			TeamID:         tr.TeamID,
			SchoolName:     tr.SchoolName,
			Seed:           tr.Seed,
			Region:         tr.Region,
			ExpectedPoints: tr.ExpectedValue,
			ExpectedMarket: expectedMarket,
			PredictedROI:   predROI,
			OurBid:         ourBid,
			ObservedROI:    observedROI,
		})
	}

	resp := labEntryReportResponse{
		ScenarioID:               scenarioID,
		CohortID:                 suiteID,
		CalcuttaID:               calcuttaID,
		CalcuttaName:             calcuttaName,
		Season:                   seasonYear,
		TournamentName:           tournamentName,
		AdvancementAlgorithmID:   advAlgID,
		AdvancementAlgorithmName: advAlgName,
		InvestmentAlgorithmID:    invAlgID,
		InvestmentAlgorithmName:  invAlgName,
		OptimizerKey:             optimizerKey,
		StrategyGenerationRunID:  strategyRunID,
		GameOutcomeRunID:         selectedGORunID,
		MarketShareRunID:         selectedMSRunID,
		BudgetPoints:             budgetPoints,
		MinTeams:                 minTeams,
		MaxTeams:                 maxTeams,
		MaxBidPoints:             maxBidPoints,
		AssumedEntries:           assumedEntries,
		ExcludedEntryName:        excluded,
		ScoringRules:             scoring,
		Teams:                    teams,
	}
	writeJSON(w, http.StatusOK, resp)
}
