package httpserver

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/app/suite_scenarios"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type simulatedCalcuttaListItem struct {
	ID                        string          `json:"id"`
	Name                      string          `json:"name"`
	Description               *string         `json:"description,omitempty"`
	TournamentID              string          `json:"tournament_id"`
	BaseCalcuttaID            *string         `json:"base_calcutta_id,omitempty"`
	StartingStateKey          string          `json:"starting_state_key"`
	ExcludedEntryName         *string         `json:"excluded_entry_name,omitempty"`
	HighlightedSimulatedEntry *string         `json:"highlighted_simulated_entry_id,omitempty"`
	Metadata                  json.RawMessage `json:"metadata"`
	CreatedAt                 time.Time       `json:"created_at"`
	UpdatedAt                 time.Time       `json:"updated_at"`
}

type simulatedCalcuttaPayout struct {
	Position    int `json:"position"`
	AmountCents int `json:"amountCents"`
}

type simulatedCalcuttaScoringRule struct {
	WinIndex      int `json:"winIndex"`
	PointsAwarded int `json:"pointsAwarded"`
}

type getSimulatedCalcuttaResponse struct {
	SimulatedCalcutta simulatedCalcuttaListItem      `json:"simulated_calcutta"`
	Payouts           []simulatedCalcuttaPayout      `json:"payouts"`
	ScoringRules      []simulatedCalcuttaScoringRule `json:"scoring_rules"`
}

type listSimulatedCalcuttasResponse struct {
	Items []simulatedCalcuttaListItem `json:"items"`
}

type createSimulatedCalcuttaRequest struct {
	Name              string                         `json:"name"`
	Description       *string                        `json:"description"`
	TournamentID      string                         `json:"tournamentId"`
	StartingStateKey  *string                        `json:"startingStateKey"`
	ExcludedEntryName *string                        `json:"excludedEntryName"`
	Payouts           []simulatedCalcuttaPayout      `json:"payouts"`
	ScoringRules      []simulatedCalcuttaScoringRule `json:"scoringRules"`
	Metadata          *json.RawMessage               `json:"metadata"`
}

type createSimulatedCalcuttaResponse struct {
	ID string `json:"id"`
}

type createSimulatedCalcuttaFromCalcuttaRequest struct {
	CalcuttaID        string           `json:"calcuttaId"`
	Name              *string          `json:"name"`
	Description       *string          `json:"description"`
	StartingStateKey  *string          `json:"startingStateKey"`
	ExcludedEntryName *string          `json:"excludedEntryName"`
	Metadata          *json.RawMessage `json:"metadata"`
}

type createSimulatedCalcuttaFromCalcuttaResponse struct {
	ID            string `json:"id"`
	CopiedEntries int    `json:"copiedEntries"`
}

type patchSimulatedCalcuttaRequest struct {
	Name                      *string          `json:"name"`
	Description               *string          `json:"description"`
	HighlightedSimulatedEntry *string          `json:"highlightedSimulatedEntryId"`
	Metadata                  *json.RawMessage `json:"metadata"`
}

type replaceSimulatedCalcuttaRulesRequest struct {
	Payouts      []simulatedCalcuttaPayout      `json:"payouts"`
	ScoringRules []simulatedCalcuttaScoringRule `json:"scoringRules"`
}

func (s *Server) registerSimulatedCalcuttaRoutes(r *mux.Router) {
	r.HandleFunc(
		"/api/simulated-calcuttas",
		s.requirePermission("analytics.suite_scenarios.read", s.handleListSimulatedCalcuttas),
	).Methods("GET", "OPTIONS")
	r.HandleFunc(
		"/api/simulated-calcuttas",
		s.requirePermission("analytics.suite_scenarios.write", s.handleCreateSimulatedCalcutta),
	).Methods("POST", "OPTIONS")
	r.HandleFunc(
		"/api/simulated-calcuttas/from-calcutta",
		s.requirePermission("analytics.suite_scenarios.write", s.handleCreateSimulatedCalcuttaFromCalcutta),
	).Methods("POST", "OPTIONS")
	r.HandleFunc(
		"/api/simulated-calcuttas/{id}",
		s.requirePermission("analytics.suite_scenarios.read", s.handleGetSimulatedCalcutta),
	).Methods("GET", "OPTIONS")
	r.HandleFunc(
		"/api/simulated-calcuttas/{id}",
		s.requirePermission("analytics.suite_scenarios.write", s.handlePatchSimulatedCalcutta),
	).Methods("PATCH", "OPTIONS")
	r.HandleFunc(
		"/api/simulated-calcuttas/{id}/rules",
		s.requirePermission("analytics.suite_scenarios.write", s.handleReplaceSimulatedCalcuttaRules),
	).Methods("PUT", "OPTIONS")
}

func validateStartingStateKey(key string) error {
	if key == "post_first_four" || key == "current" {
		return nil
	}
	return errors.New("invalid starting_state_key")
}

func normalizeOptionalStringPtr(p *string) *string {
	if p == nil {
		return nil
	}
	v := strings.TrimSpace(*p)
	if v == "" {
		return nil
	}
	return &v
}

func normalizeOptionalJSONObj(req *json.RawMessage) (json.RawMessage, error) {
	if req == nil {
		return json.RawMessage([]byte("{}")), nil
	}
	b := []byte(*req)
	if len(b) == 0 {
		return json.RawMessage([]byte("{}")), nil
	}
	var parsed any
	if err := json.Unmarshal(b, &parsed); err != nil {
		return nil, err
	}
	if _, ok := parsed.(map[string]any); !ok {
		return nil, errors.New("metadata must be an object")
	}
	return json.RawMessage(b), nil
}

func (s *Server) handleListSimulatedCalcuttas(w http.ResponseWriter, r *http.Request) {
	tournamentID := strings.TrimSpace(r.URL.Query().Get("tournament_id"))
	if tournamentID != "" {
		if _, err := uuid.Parse(tournamentID); err != nil {
			writeError(w, r, http.StatusBadRequest, "validation_error", "tournament_id must be a valid UUID", "tournament_id")
			return
		}
	}
	baseCalcuttaID := strings.TrimSpace(r.URL.Query().Get("base_calcutta_id"))
	if baseCalcuttaID != "" {
		if _, err := uuid.Parse(baseCalcuttaID); err != nil {
			writeError(w, r, http.StatusBadRequest, "validation_error", "base_calcutta_id must be a valid UUID", "base_calcutta_id")
			return
		}
	}
	cohortID := strings.TrimSpace(r.URL.Query().Get("cohort_id"))
	if cohortID != "" {
		if _, err := uuid.Parse(cohortID); err != nil {
			writeError(w, r, http.StatusBadRequest, "validation_error", "cohort_id must be a valid UUID", "cohort_id")
			return
		}
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

	var tournamentIDPtr *string
	if tournamentID != "" {
		v := tournamentID
		tournamentIDPtr = &v
	}
	var baseCalcuttaIDPtr *string
	if baseCalcuttaID != "" {
		v := baseCalcuttaID
		baseCalcuttaIDPtr = &v
	}
	var cohortIDPtr *string
	if cohortID != "" {
		v := cohortID
		cohortIDPtr = &v
	}

	items, err := s.app.SuiteScenarios.ListSimulatedCalcuttasWithFilters(r.Context(), tournamentIDPtr, baseCalcuttaIDPtr, cohortIDPtr, limit, offset)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	respItems := make([]simulatedCalcuttaListItem, 0, len(items))
	for _, it := range items {
		respItems = append(respItems, simulatedCalcuttaListItem{
			ID:                        it.ID,
			Name:                      it.Name,
			Description:               it.Description,
			TournamentID:              it.TournamentID,
			BaseCalcuttaID:            it.BaseCalcuttaID,
			StartingStateKey:          it.StartingStateKey,
			ExcludedEntryName:         it.ExcludedEntryName,
			HighlightedSimulatedEntry: it.HighlightedSimulatedEntry,
			Metadata:                  it.Metadata,
			CreatedAt:                 it.CreatedAt,
			UpdatedAt:                 it.UpdatedAt,
		})
	}

	writeJSON(w, http.StatusOK, listSimulatedCalcuttasResponse{Items: respItems})
}

func (s *Server) handleGetSimulatedCalcutta(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := strings.TrimSpace(vars["id"])
	if id == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "id is required", "id")
		return
	}
	if _, err := uuid.Parse(id); err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "id must be a valid UUID", "id")
		return
	}

	calcutta, payouts, rules, err := s.app.SuiteScenarios.GetSimulatedCalcutta(r.Context(), id)
	if err != nil {
		if errors.Is(err, suite_scenarios.ErrSimulatedCalcuttaNotFound) {
			writeError(w, r, http.StatusNotFound, "not_found", "Simulated calcutta not found", "id")
			return
		}
		writeErrorFromErr(w, r, err)
		return
	}

	it := simulatedCalcuttaListItem{
		ID:                        calcutta.ID,
		Name:                      calcutta.Name,
		Description:               calcutta.Description,
		TournamentID:              calcutta.TournamentID,
		BaseCalcuttaID:            calcutta.BaseCalcuttaID,
		StartingStateKey:          calcutta.StartingStateKey,
		ExcludedEntryName:         calcutta.ExcludedEntryName,
		HighlightedSimulatedEntry: calcutta.HighlightedSimulatedEntry,
		Metadata:                  calcutta.Metadata,
		CreatedAt:                 calcutta.CreatedAt,
		UpdatedAt:                 calcutta.UpdatedAt,
	}

	payoutResp := make([]simulatedCalcuttaPayout, 0, len(payouts))
	for _, p := range payouts {
		payoutResp = append(payoutResp, simulatedCalcuttaPayout{Position: p.Position, AmountCents: p.AmountCents})
	}
	ruleResp := make([]simulatedCalcuttaScoringRule, 0, len(rules))
	for _, rr := range rules {
		ruleResp = append(ruleResp, simulatedCalcuttaScoringRule{WinIndex: rr.WinIndex, PointsAwarded: rr.PointsAwarded})
	}

	writeJSON(w, http.StatusOK, getSimulatedCalcuttaResponse{SimulatedCalcutta: it, Payouts: payoutResp, ScoringRules: ruleResp})
}

func (s *Server) handleCreateSimulatedCalcutta(w http.ResponseWriter, r *http.Request) {
	var req createSimulatedCalcuttaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.TournamentID = strings.TrimSpace(req.TournamentID)
	if req.Description != nil {
		v := strings.TrimSpace(*req.Description)
		if v == "" {
			req.Description = nil
		} else {
			req.Description = &v
		}
	}
	startingStateKey := "post_first_four"
	if req.StartingStateKey != nil {
		v := strings.TrimSpace(*req.StartingStateKey)
		if v != "" {
			if err := validateStartingStateKey(v); err != nil {
				writeError(w, r, http.StatusBadRequest, "validation_error", "startingStateKey must be 'current' or 'post_first_four'", "startingStateKey")
				return
			}
			startingStateKey = v
		}
	}

	excludedEntryName := normalizeOptionalStringPtr(req.ExcludedEntryName)

	if req.Name == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "name is required", "name")
		return
	}
	if req.TournamentID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "tournamentId is required", "tournamentId")
		return
	}
	if _, err := uuid.Parse(req.TournamentID); err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "tournamentId must be a valid UUID", "tournamentId")
		return
	}
	if len(req.Payouts) == 0 {
		writeError(w, r, http.StatusBadRequest, "validation_error", "payouts is required", "payouts")
		return
	}
	if len(req.ScoringRules) == 0 {
		writeError(w, r, http.StatusBadRequest, "validation_error", "scoringRules is required", "scoringRules")
		return
	}

	for i := range req.Payouts {
		if req.Payouts[i].Position <= 0 {
			writeError(w, r, http.StatusBadRequest, "validation_error", "payouts.position must be positive", "payouts")
			return
		}
		if req.Payouts[i].AmountCents <= 0 {
			writeError(w, r, http.StatusBadRequest, "validation_error", "payouts.amountCents must be positive", "payouts")
			return
		}
	}
	for i := range req.ScoringRules {
		if req.ScoringRules[i].WinIndex < 0 {
			writeError(w, r, http.StatusBadRequest, "validation_error", "scoringRules.winIndex must be non-negative", "scoringRules")
			return
		}
		if req.ScoringRules[i].PointsAwarded <= 0 {
			writeError(w, r, http.StatusBadRequest, "validation_error", "scoringRules.pointsAwarded must be positive", "scoringRules")
			return
		}
	}

	metadataJSON, err := normalizeOptionalJSONObj(req.Metadata)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "metadata must be valid JSON object", "metadata")
		return
	}

	params := suite_scenarios.CreateSimulatedCalcuttaParams{
		Name:              req.Name,
		Description:       req.Description,
		TournamentID:      req.TournamentID,
		StartingStateKey:  startingStateKey,
		ExcludedEntryName: excludedEntryName,
		Metadata:          metadataJSON,
		Payouts:           make([]suite_scenarios.SimulatedCalcuttaPayout, 0, len(req.Payouts)),
		ScoringRules:      make([]suite_scenarios.SimulatedCalcuttaScoringRule, 0, len(req.ScoringRules)),
	}
	for _, p := range req.Payouts {
		params.Payouts = append(params.Payouts, suite_scenarios.SimulatedCalcuttaPayout{Position: p.Position, AmountCents: p.AmountCents})
	}
	for _, sr := range req.ScoringRules {
		params.ScoringRules = append(params.ScoringRules, suite_scenarios.SimulatedCalcuttaScoringRule{WinIndex: sr.WinIndex, PointsAwarded: sr.PointsAwarded})
	}

	createdID, err := s.app.SuiteScenarios.CreateSimulatedCalcutta(r.Context(), params)
	if err != nil {
		if errors.Is(err, suite_scenarios.ErrDuplicatePayoutPosition) {
			writeError(w, r, http.StatusConflict, "conflict", "Duplicate payout position", "payouts")
			return
		}
		if errors.Is(err, suite_scenarios.ErrDuplicateScoringRuleWinIndex) {
			writeError(w, r, http.StatusConflict, "conflict", "Duplicate scoring rule winIndex", "scoringRules")
			return
		}
		writeErrorFromErr(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, createSimulatedCalcuttaResponse{ID: createdID})
}

func (s *Server) handleCreateSimulatedCalcuttaFromCalcutta(w http.ResponseWriter, r *http.Request) {
	var req createSimulatedCalcuttaFromCalcuttaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}

	req.CalcuttaID = strings.TrimSpace(req.CalcuttaID)
	if req.CalcuttaID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "calcuttaId is required", "calcuttaId")
		return
	}
	if _, err := uuid.Parse(req.CalcuttaID); err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "calcuttaId must be a valid UUID", "calcuttaId")
		return
	}

	startingStateKey := "post_first_four"
	if req.StartingStateKey != nil {
		v := strings.TrimSpace(*req.StartingStateKey)
		if v != "" {
			if err := validateStartingStateKey(v); err != nil {
				writeError(w, r, http.StatusBadRequest, "validation_error", "startingStateKey must be 'current' or 'post_first_four'", "startingStateKey")
				return
			}
			startingStateKey = v
		}
	}

	excludedEntryName := normalizeOptionalStringPtr(req.ExcludedEntryName)
	metadataJSON, err := normalizeOptionalJSONObj(req.Metadata)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "metadata must be valid JSON object", "metadata")
		return
	}

	params := suite_scenarios.CreateSimulatedCalcuttaFromCalcuttaParams{
		CalcuttaID:        req.CalcuttaID,
		Name:              req.Name,
		Description:       req.Description,
		StartingStateKey:  startingStateKey,
		ExcludedEntryName: excludedEntryName,
		Metadata:          metadataJSON,
	}

	createdID, copiedEntries, err := s.app.SuiteScenarios.CreateSimulatedCalcuttaFromCalcutta(r.Context(), params)
	if err != nil {
		if errors.Is(err, suite_scenarios.ErrCalcuttaNotFound) {
			writeError(w, r, http.StatusNotFound, "not_found", "Calcutta not found", "calcuttaId")
			return
		}
		writeErrorFromErr(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, createSimulatedCalcuttaFromCalcuttaResponse{ID: createdID, CopiedEntries: copiedEntries})
}

func (s *Server) handlePatchSimulatedCalcutta(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := strings.TrimSpace(vars["id"])
	if id == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "id is required", "id")
		return
	}
	if _, err := uuid.Parse(id); err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "id must be a valid UUID", "id")
		return
	}

	var req patchSimulatedCalcuttaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}

	if req.Name != nil {
		v := strings.TrimSpace(*req.Name)
		if v == "" {
			writeError(w, r, http.StatusBadRequest, "validation_error", "name cannot be empty", "name")
			return
		}
		req.Name = &v
	}
	if req.Description != nil {
		v := strings.TrimSpace(*req.Description)
		req.Description = &v
		if v == "" {
			req.Description = nil
		}
	}
	if req.HighlightedSimulatedEntry != nil {
		v := strings.TrimSpace(*req.HighlightedSimulatedEntry)
		if v != "" {
			if _, err := uuid.Parse(v); err != nil {
				writeError(w, r, http.StatusBadRequest, "validation_error", "highlightedSimulatedEntryId must be a valid UUID", "highlightedSimulatedEntryId")
				return
			}
		}
		req.HighlightedSimulatedEntry = &v
	}

	var metadataPtr *json.RawMessage
	if req.Metadata != nil {
		mj, err := normalizeOptionalJSONObj(req.Metadata)
		if err != nil {
			writeError(w, r, http.StatusBadRequest, "validation_error", "metadata must be valid JSON object", "metadata")
			return
		}
		mj2 := json.RawMessage(mj)
		metadataPtr = &mj2
	}

	params := suite_scenarios.PatchSimulatedCalcuttaParams{
		Name:                      req.Name,
		Description:               req.Description,
		HighlightedSimulatedEntry: req.HighlightedSimulatedEntry,
		Metadata:                  metadataPtr,
	}

	if err := s.app.SuiteScenarios.PatchSimulatedCalcutta(r.Context(), id, params); err != nil {
		if errors.Is(err, suite_scenarios.ErrSimulatedCalcuttaNotFound) {
			writeError(w, r, http.StatusNotFound, "not_found", "Simulated calcutta not found", "id")
			return
		}
		if errors.Is(err, suite_scenarios.ErrHighlightedEntryDoesNotBelong) {
			writeError(w, r, http.StatusBadRequest, "validation_error", "highlightedSimulatedEntryId must belong to this simulated calcutta", "highlightedSimulatedEntryId")
			return
		}
		writeErrorFromErr(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleReplaceSimulatedCalcuttaRules(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := strings.TrimSpace(vars["id"])
	if id == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "id is required", "id")
		return
	}
	if _, err := uuid.Parse(id); err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "id must be a valid UUID", "id")
		return
	}

	var req replaceSimulatedCalcuttaRulesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}

	if len(req.Payouts) == 0 {
		writeError(w, r, http.StatusBadRequest, "validation_error", "payouts is required", "payouts")
		return
	}
	if len(req.ScoringRules) == 0 {
		writeError(w, r, http.StatusBadRequest, "validation_error", "scoringRules is required", "scoringRules")
		return
	}

	for i := range req.Payouts {
		if req.Payouts[i].Position <= 0 {
			writeError(w, r, http.StatusBadRequest, "validation_error", "payouts.position must be positive", "payouts")
			return
		}
		if req.Payouts[i].AmountCents <= 0 {
			writeError(w, r, http.StatusBadRequest, "validation_error", "payouts.amountCents must be positive", "payouts")
			return
		}
	}
	for i := range req.ScoringRules {
		if req.ScoringRules[i].WinIndex < 0 {
			writeError(w, r, http.StatusBadRequest, "validation_error", "scoringRules.winIndex must be non-negative", "scoringRules")
			return
		}
		if req.ScoringRules[i].PointsAwarded <= 0 {
			writeError(w, r, http.StatusBadRequest, "validation_error", "scoringRules.pointsAwarded must be positive", "scoringRules")
			return
		}
	}

	params := suite_scenarios.ReplaceSimulatedCalcuttaRulesParams{
		Payouts:      make([]suite_scenarios.SimulatedCalcuttaPayout, 0, len(req.Payouts)),
		ScoringRules: make([]suite_scenarios.SimulatedCalcuttaScoringRule, 0, len(req.ScoringRules)),
	}
	for _, p := range req.Payouts {
		params.Payouts = append(params.Payouts, suite_scenarios.SimulatedCalcuttaPayout{Position: p.Position, AmountCents: p.AmountCents})
	}
	for _, sr := range req.ScoringRules {
		params.ScoringRules = append(params.ScoringRules, suite_scenarios.SimulatedCalcuttaScoringRule{WinIndex: sr.WinIndex, PointsAwarded: sr.PointsAwarded})
	}

	if err := s.app.SuiteScenarios.ReplaceSimulatedCalcuttaRules(r.Context(), id, params); err != nil {
		if errors.Is(err, suite_scenarios.ErrSimulatedCalcuttaNotFound) {
			writeError(w, r, http.StatusNotFound, "not_found", "Simulated calcutta not found", "id")
			return
		}
		if errors.Is(err, suite_scenarios.ErrDuplicatePayoutPosition) {
			writeError(w, r, http.StatusConflict, "conflict", "Duplicate payout position", "payouts")
			return
		}
		if errors.Is(err, suite_scenarios.ErrDuplicateScoringRuleWinIndex) {
			writeError(w, r, http.StatusConflict, "conflict", "Duplicate scoring rule winIndex", "scoringRules")
			return
		}
		writeErrorFromErr(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}
