package httpserver

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func (s *Server) registerSyntheticCompatibilityRoutes(r *mux.Router) {
	r.HandleFunc(
		"/api/synthetic-calcutta-cohorts",
		s.requirePermission("analytics.suites.read", s.handleListSyntheticCalcuttaCohorts),
	).Methods("GET", "OPTIONS")
	r.HandleFunc(
		"/api/synthetic-calcutta-cohorts/{id}",
		s.requirePermission("analytics.suites.read", s.handleGetSyntheticCalcuttaCohort),
	).Methods("GET", "OPTIONS")
	r.HandleFunc(
		"/api/synthetic-calcutta-cohorts/{id}",
		s.requirePermission("analytics.suites.write", s.handlePatchSyntheticCalcuttaCohort),
	).Methods("PATCH", "OPTIONS")

	r.HandleFunc(
		"/api/synthetic-calcuttas",
		s.requirePermission("analytics.suite_scenarios.read", s.handleListSyntheticCalcuttas),
	).Methods("GET", "OPTIONS")
	r.HandleFunc(
		"/api/synthetic-calcuttas",
		s.requirePermission("analytics.suite_scenarios.write", s.handleCreateSyntheticCalcutta),
	).Methods("POST", "OPTIONS")
	r.HandleFunc(
		"/api/synthetic-calcuttas/{id}",
		s.requirePermission("analytics.suite_scenarios.read", s.handleGetSyntheticCalcutta),
	).Methods("GET", "OPTIONS")

	r.HandleFunc(
		"/api/simulation-run-batches",
		s.requirePermission("analytics.suite_executions.write", s.handleCreateSimulationRunBatch),
	).Methods("POST", "OPTIONS")
	r.HandleFunc(
		"/api/simulation-run-batches",
		s.requirePermission("analytics.suite_executions.read", s.handleListSimulationRunBatches),
	).Methods("GET", "OPTIONS")
	r.HandleFunc(
		"/api/simulation-run-batches/{id}",
		s.requirePermission("analytics.suite_executions.read", s.handleGetSimulationRunBatch),
	).Methods("GET", "OPTIONS")

	r.HandleFunc(
		"/api/simulation-runs",
		s.requirePermission("analytics.suite_calcutta_evaluations.write", s.handleCreateSimulationRun),
	).Methods("POST", "OPTIONS")
	r.HandleFunc(
		"/api/simulation-runs",
		s.requirePermission("analytics.suite_calcutta_evaluations.read", s.handleListSimulationRuns),
	).Methods("GET", "OPTIONS")
	r.HandleFunc(
		"/api/simulation-runs/{id}",
		s.requirePermission("analytics.suite_calcutta_evaluations.read", s.getSuiteCalcuttaEvaluationHandler),
	).Methods("GET", "OPTIONS")
	r.HandleFunc(
		"/api/simulation-runs/{id}/result",
		s.requirePermission("analytics.suite_calcutta_evaluations.read", s.getSuiteCalcuttaEvaluationResultHandler),
	).Methods("GET", "OPTIONS")
	r.HandleFunc(
		"/api/simulation-runs/{id}/entries/{snapshotEntryId}",
		s.requirePermission("analytics.suite_calcutta_evaluations.read", s.getSuiteCalcuttaEvaluationSnapshotEntryHandler),
	).Methods("GET", "OPTIONS")
}

func (s *Server) handleListSyntheticCalcuttaCohorts(w http.ResponseWriter, r *http.Request) {
	s.listSuitesHandler(w, r)
}

func (s *Server) handleGetSyntheticCalcuttaCohort(w http.ResponseWriter, r *http.Request) {
	s.getSuiteHandler(w, r)
}

func (s *Server) handlePatchSyntheticCalcuttaCohort(w http.ResponseWriter, r *http.Request) {
	s.updateSuiteHandler(w, r)
}

func (s *Server) handleListSyntheticCalcuttas(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if v := strings.TrimSpace(q.Get("cohort_id")); v != "" {
		q.Set("suite_id", v)
	}
	r.URL.RawQuery = q.Encode()
	s.listSuiteScenariosHandler(w, r)
}

func (s *Server) handleGetSyntheticCalcutta(w http.ResponseWriter, r *http.Request) {
	s.getSuiteScenarioHandler(w, r)
}

type createSyntheticCalcuttaRequest struct {
	CohortID                  string  `json:"cohortId"`
	CalcuttaID                string  `json:"calcuttaId"`
	CalcuttaSnapshotID        *string `json:"calcuttaSnapshotId"`
	FocusStrategyGenerationID *string `json:"focusStrategyGenerationRunId"`
	FocusEntryName            *string `json:"focusEntryName"`
	StartingStateKey          *string `json:"startingStateKey"`
	ExcludedEntryName         *string `json:"excludedEntryName"`
}

func (s *Server) handleCreateSyntheticCalcutta(w http.ResponseWriter, r *http.Request) {
	var req createSyntheticCalcuttaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}

	body := map[string]any{
		"suiteId":                      req.CohortID,
		"calcuttaId":                   req.CalcuttaID,
		"calcuttaSnapshotId":           req.CalcuttaSnapshotID,
		"focusStrategyGenerationRunId": req.FocusStrategyGenerationID,
		"focusEntryName":               req.FocusEntryName,
		"startingStateKey":             req.StartingStateKey,
		"excludedEntryName":            req.ExcludedEntryName,
	}
	b, _ := json.Marshal(body)
	cloned := r.Clone(r.Context())
	cloned.Body = io.NopCloser(bytes.NewReader(b))
	s.createSuiteScenarioHandler(w, cloned)
}

func (s *Server) handleListSimulationRunBatches(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if v := strings.TrimSpace(q.Get("cohort_id")); v != "" {
		q.Set("suite_id", v)
	}
	r.URL.RawQuery = q.Encode()
	s.listSuiteExecutionsHandler(w, r)
}

func (s *Server) handleGetSimulationRunBatch(w http.ResponseWriter, r *http.Request) {
	s.getSuiteExecutionHandler(w, r)
}

type createSimulationRunBatchRequest struct {
	CohortID          string   `json:"cohortId"`
	Name              *string  `json:"name"`
	CalcuttaIDs       []string `json:"calcuttaIds"`
	OptimizerKey      *string  `json:"optimizerKey"`
	NSims             *int     `json:"nSims"`
	Seed              *int     `json:"seed"`
	StartingStateKey  *string  `json:"startingStateKey"`
	ExcludedEntryName *string  `json:"excludedEntryName"`
}

func (s *Server) handleCreateSimulationRunBatch(w http.ResponseWriter, r *http.Request) {
	var req createSimulationRunBatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}

	body := map[string]any{
		"suiteId":           req.CohortID,
		"name":              req.Name,
		"calcuttaIds":       req.CalcuttaIDs,
		"optimizerKey":      req.OptimizerKey,
		"nSims":             req.NSims,
		"seed":              req.Seed,
		"startingStateKey":  req.StartingStateKey,
		"excludedEntryName": req.ExcludedEntryName,
	}
	b, _ := json.Marshal(body)
	cloned := r.Clone(r.Context())
	cloned.Body = io.NopCloser(bytes.NewReader(b))
	s.createSuiteExecutionHandler(w, cloned)
}

func (s *Server) handleListSimulationRuns(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if v := strings.TrimSpace(q.Get("cohort_id")); v != "" {
		q.Set("suite_id", v)
	}
	if v := strings.TrimSpace(q.Get("simulation_run_batch_id")); v != "" {
		q.Set("suite_execution_id", v)
	}
	r.URL.RawQuery = q.Encode()
	s.listSuiteCalcuttaEvaluationsHandler(w, r)
}

type createSimulationRunRequest struct {
	SyntheticCalcuttaID  string  `json:"syntheticCalcuttaId"`
	SimulationRunBatchID *string `json:"simulationRunBatchId"`
	OptimizerKey         *string `json:"optimizerKey"`
	GameOutcomeRunID     *string `json:"gameOutcomeRunId"`
	MarketShareRunID     *string `json:"marketShareRunId"`
	NSims                *int    `json:"nSims"`
	Seed                 *int    `json:"seed"`
}

func (s *Server) handleCreateSimulationRun(w http.ResponseWriter, r *http.Request) {
	var req createSimulationRunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}
	req.SyntheticCalcuttaID = strings.TrimSpace(req.SyntheticCalcuttaID)
	if req.SyntheticCalcuttaID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "syntheticCalcuttaId is required", "syntheticCalcuttaId")
		return
	}
	if _, err := uuid.Parse(req.SyntheticCalcuttaID); err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "syntheticCalcuttaId must be a valid UUID", "syntheticCalcuttaId")
		return
	}

	ctx := r.Context()
	var suiteID string
	var calcuttaID string
	var startingStateKey *string
	var excludedEntryName *string
	if err := s.pool.QueryRow(ctx, `
		SELECT suite_id::text, calcutta_id::text, starting_state_key, excluded_entry_name
		FROM derived.suite_scenarios
		WHERE id = $1::uuid
			AND deleted_at IS NULL
		LIMIT 1
	`, req.SyntheticCalcuttaID).Scan(&suiteID, &calcuttaID, &startingStateKey, &excludedEntryName); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	body := map[string]any{
		"calcuttaId":        calcuttaID,
		"suiteId":           suiteID,
		"suiteExecutionId":  req.SimulationRunBatchID,
		"optimizerKey":      req.OptimizerKey,
		"gameOutcomeRunId":  req.GameOutcomeRunID,
		"marketShareRunId":  req.MarketShareRunID,
		"nSims":             0,
		"seed":              0,
		"startingStateKey":  "post_first_four",
		"excludedEntryName": excludedEntryName,
	}
	if startingStateKey != nil && strings.TrimSpace(*startingStateKey) != "" {
		body["startingStateKey"] = strings.TrimSpace(*startingStateKey)
	}
	if req.NSims != nil {
		body["nSims"] = *req.NSims
	}
	if req.Seed != nil {
		body["seed"] = *req.Seed
	}

	b, _ := json.Marshal(body)
	cloned := r.Clone(ctx)
	cloned.Body = io.NopCloser(bytes.NewReader(b))
	s.createSuiteCalcuttaEvaluationHandler(w, cloned)
}
