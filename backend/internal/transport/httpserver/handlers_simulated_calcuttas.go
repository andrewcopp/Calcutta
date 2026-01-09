package httpserver

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
 			id::text,
 			name,
 			description,
 			tournament_id::text,
 			base_calcutta_id::text,
 			starting_state_key,
 			excluded_entry_name,
 			highlighted_simulated_entry_id::text,
 			metadata_json,
 			created_at,
 			updated_at
 		FROM derived.simulated_calcuttas
 		WHERE deleted_at IS NULL
 			AND ($1::uuid IS NULL OR tournament_id = $1::uuid)
 		ORDER BY created_at DESC
 		LIMIT $2::int
 		OFFSET $3::int
 	`, nullUUIDParam(tournamentID), limit, offset)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	defer rows.Close()

	items := make([]simulatedCalcuttaListItem, 0)
	for rows.Next() {
		var it simulatedCalcuttaListItem
		if err := rows.Scan(
			&it.ID,
			&it.Name,
			&it.Description,
			&it.TournamentID,
			&it.BaseCalcuttaID,
			&it.StartingStateKey,
			&it.ExcludedEntryName,
			&it.HighlightedSimulatedEntry,
			&it.Metadata,
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

	writeJSON(w, http.StatusOK, listSimulatedCalcuttasResponse{Items: items})
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

	ctx := r.Context()
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{AccessMode: pgx.ReadOnly})
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var it simulatedCalcuttaListItem
	if err := tx.QueryRow(ctx, `
 		SELECT
 			id::text,
 			name,
 			description,
 			tournament_id::text,
 			base_calcutta_id::text,
 			starting_state_key,
 			excluded_entry_name,
 			highlighted_simulated_entry_id::text,
 			metadata_json,
 			created_at,
 			updated_at
 		FROM derived.simulated_calcuttas
 		WHERE id = $1::uuid
 			AND deleted_at IS NULL
 		LIMIT 1
 	`, id).Scan(
		&it.ID,
		&it.Name,
		&it.Description,
		&it.TournamentID,
		&it.BaseCalcuttaID,
		&it.StartingStateKey,
		&it.ExcludedEntryName,
		&it.HighlightedSimulatedEntry,
		&it.Metadata,
		&it.CreatedAt,
		&it.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, r, http.StatusNotFound, "not_found", "Simulated calcutta not found", "id")
			return
		}
		writeErrorFromErr(w, r, err)
		return
	}

	payoutRows, err := tx.Query(ctx, `
 		SELECT position::int, amount_cents::int
 		FROM derived.simulated_calcutta_payouts
 		WHERE simulated_calcutta_id = $1::uuid
 			AND deleted_at IS NULL
 		ORDER BY position ASC
 	`, id)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	defer payoutRows.Close()

	payouts := make([]simulatedCalcuttaPayout, 0)
	for payoutRows.Next() {
		var p simulatedCalcuttaPayout
		if err := payoutRows.Scan(&p.Position, &p.AmountCents); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
		payouts = append(payouts, p)
	}
	if err := payoutRows.Err(); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	ruleRows, err := tx.Query(ctx, `
 		SELECT win_index::int, points_awarded::int
 		FROM derived.simulated_calcutta_scoring_rules
 		WHERE simulated_calcutta_id = $1::uuid
 			AND deleted_at IS NULL
 		ORDER BY win_index ASC
 	`, id)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	defer ruleRows.Close()

	rules := make([]simulatedCalcuttaScoringRule, 0)
	for ruleRows.Next() {
		var rr simulatedCalcuttaScoringRule
		if err := ruleRows.Scan(&rr.WinIndex, &rr.PointsAwarded); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
		rules = append(rules, rr)
	}
	if err := ruleRows.Err(); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, getSimulatedCalcuttaResponse{SimulatedCalcutta: it, Payouts: payouts, ScoringRules: rules})
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

	ctx := r.Context()
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

	var createdID string
	if err := tx.QueryRow(ctx, `
 		INSERT INTO derived.simulated_calcuttas (
 			name,
 			description,
 			tournament_id,
 			base_calcutta_id,
 			starting_state_key,
 			excluded_entry_name,
 			metadata_json
 		)
 		VALUES ($1, $2, $3::uuid, NULL, $4, $5, $6::jsonb)
 		RETURNING id::text
 	`, req.Name, nullUUIDParamPtr(req.Description), req.TournamentID, startingStateKey, excludedEntryName, []byte(metadataJSON)).Scan(&createdID); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	for _, p := range req.Payouts {
		if _, err := tx.Exec(ctx, `
 			INSERT INTO derived.simulated_calcutta_payouts (
 				simulated_calcutta_id,
 				position,
 				amount_cents
 			)
 			VALUES ($1::uuid, $2::int, $3::int)
 		`, createdID, p.Position, p.AmountCents); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				writeError(w, r, http.StatusConflict, "conflict", "Duplicate payout position", "payouts")
				return
			}
			writeErrorFromErr(w, r, err)
			return
		}
	}

	for _, sr := range req.ScoringRules {
		if _, err := tx.Exec(ctx, `
 			INSERT INTO derived.simulated_calcutta_scoring_rules (
 				simulated_calcutta_id,
 				win_index,
 				points_awarded
 			)
 			VALUES ($1::uuid, $2::int, $3::int)
 		`, createdID, sr.WinIndex, sr.PointsAwarded); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				writeError(w, r, http.StatusConflict, "conflict", "Duplicate scoring rule winIndex", "scoringRules")
				return
			}
			writeErrorFromErr(w, r, err)
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	committed = true

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

	ctx := r.Context()
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

	var tournamentID string
	var calcuttaName string
	if err := tx.QueryRow(ctx, `
 		SELECT tournament_id::text, name
 		FROM core.calcuttas
 		WHERE id = $1::uuid
 			AND deleted_at IS NULL
 		LIMIT 1
 	`, req.CalcuttaID).Scan(&tournamentID, &calcuttaName); err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, r, http.StatusNotFound, "not_found", "Calcutta not found", "calcuttaId")
			return
		}
		writeErrorFromErr(w, r, err)
		return
	}

	name := "Simulated " + strings.TrimSpace(calcuttaName)
	if req.Name != nil {
		v := strings.TrimSpace(*req.Name)
		if v != "" {
			name = v
		}
	}
	var desc any
	if req.Description != nil {
		v := strings.TrimSpace(*req.Description)
		if v == "" {
			desc = nil
		} else {
			desc = v
		}
	} else {
		desc = nil
	}

	var createdID string
	if err := tx.QueryRow(ctx, `
 		INSERT INTO derived.simulated_calcuttas (
 			name,
 			description,
 			tournament_id,
 			base_calcutta_id,
 			starting_state_key,
 			excluded_entry_name,
 			metadata_json
 		)
 		VALUES ($1, $2, $3::uuid, $4::uuid, $5, $6, $7::jsonb)
 		RETURNING id::text
 	`, name, desc, tournamentID, req.CalcuttaID, startingStateKey, excludedEntryName, []byte(metadataJSON)).Scan(&createdID); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	if _, err := tx.Exec(ctx, `
 		INSERT INTO derived.simulated_calcutta_payouts (simulated_calcutta_id, position, amount_cents)
 		SELECT $1::uuid, position, amount_cents
 		FROM core.payouts
 		WHERE calcutta_id = $2::uuid
 			AND deleted_at IS NULL
 	`, createdID, req.CalcuttaID); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	if _, err := tx.Exec(ctx, `
 		INSERT INTO derived.simulated_calcutta_scoring_rules (simulated_calcutta_id, win_index, points_awarded)
 		SELECT $1::uuid, win_index, points_awarded
 		FROM core.calcutta_scoring_rules
 		WHERE calcutta_id = $2::uuid
 			AND deleted_at IS NULL
 	`, createdID, req.CalcuttaID); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	entryRows, err := tx.Query(ctx, `
 		SELECT id::text, name
 		FROM core.entries
 		WHERE calcutta_id = $1::uuid
 			AND deleted_at IS NULL
 		ORDER BY created_at ASC
 	`, req.CalcuttaID)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	defer entryRows.Close()

	copiedEntries := 0
	for entryRows.Next() {
		var entryID string
		var entryName string
		if err := entryRows.Scan(&entryID, &entryName); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}

		var simulatedEntryID string
		if err := tx.QueryRow(ctx, `
 			INSERT INTO derived.simulated_entries (
 				simulated_calcutta_id,
 				display_name,
 				source_kind,
 				source_entry_id,
 				source_candidate_id
 			)
 			VALUES ($1::uuid, $2, 'from_real_entry', $3::uuid, NULL)
 			RETURNING id::text
 		`, createdID, strings.TrimSpace(entryName), entryID).Scan(&simulatedEntryID); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}

		teamRows, err := tx.Query(ctx, `
 			SELECT team_id::text, bid_points::int
 			FROM core.entry_teams
 			WHERE entry_id = $1::uuid
 				AND deleted_at IS NULL
 			ORDER BY bid_points DESC
 		`, entryID)
		if err != nil {
			writeErrorFromErr(w, r, err)
			return
		}

		for teamRows.Next() {
			var teamID string
			var bidPoints int
			if err := teamRows.Scan(&teamID, &bidPoints); err != nil {
				teamRows.Close()
				writeErrorFromErr(w, r, err)
				return
			}
			if _, err := tx.Exec(ctx, `
 				INSERT INTO derived.simulated_entry_teams (
 					simulated_entry_id,
 					team_id,
 					bid_points
 				)
 				VALUES ($1::uuid, $2::uuid, $3::int)
 			`, simulatedEntryID, teamID, bidPoints); err != nil {
				teamRows.Close()
				writeErrorFromErr(w, r, err)
				return
			}
		}
		if err := teamRows.Err(); err != nil {
			teamRows.Close()
			writeErrorFromErr(w, r, err)
			return
		}
		teamRows.Close()

		copiedEntries++
	}
	if err := entryRows.Err(); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	committed = true

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

	ctx := r.Context()

	var existingName string
	var existingDescription *string
	var existingHighlighted *string
	var existingMetadata json.RawMessage
	if err := s.pool.QueryRow(ctx, `
 		SELECT name, description, highlighted_simulated_entry_id::text, metadata_json
 		FROM derived.simulated_calcuttas
 		WHERE id = $1::uuid
 			AND deleted_at IS NULL
 		LIMIT 1
 	`, id).Scan(&existingName, &existingDescription, &existingHighlighted, &existingMetadata); err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, r, http.StatusNotFound, "not_found", "Simulated calcutta not found", "id")
			return
		}
		writeErrorFromErr(w, r, err)
		return
	}

	newName := existingName
	if req.Name != nil {
		v := strings.TrimSpace(*req.Name)
		if v == "" {
			writeError(w, r, http.StatusBadRequest, "validation_error", "name cannot be empty", "name")
			return
		}
		newName = v
	}

	newDescription := existingDescription
	if req.Description != nil {
		newDescription = normalizeOptionalStringPtr(req.Description)
	}

	newHighlighted := existingHighlighted
	if req.HighlightedSimulatedEntry != nil {
		v := strings.TrimSpace(*req.HighlightedSimulatedEntry)
		if v == "" {
			newHighlighted = nil
		} else {
			if _, err := uuid.Parse(v); err != nil {
				writeError(w, r, http.StatusBadRequest, "validation_error", "highlightedSimulatedEntryId must be a valid UUID", "highlightedSimulatedEntryId")
				return
			}
			var ok bool
			if err := s.pool.QueryRow(ctx, `
 				SELECT EXISTS(
 					SELECT 1
 					FROM derived.simulated_entries e
 					WHERE e.id = $1::uuid
 						AND e.simulated_calcutta_id = $2::uuid
 						AND e.deleted_at IS NULL
 				)
 			`, v, id).Scan(&ok); err != nil {
				writeErrorFromErr(w, r, err)
				return
			}
			if !ok {
				writeError(w, r, http.StatusBadRequest, "validation_error", "highlightedSimulatedEntryId must belong to this simulated calcutta", "highlightedSimulatedEntryId")
				return
			}
			newHighlighted = &v
		}
	}

	newMetadata := existingMetadata
	if req.Metadata != nil {
		mj, err := normalizeOptionalJSONObj(req.Metadata)
		if err != nil {
			writeError(w, r, http.StatusBadRequest, "validation_error", "metadata must be valid JSON object", "metadata")
			return
		}
		newMetadata = mj
	}
	if len(newMetadata) == 0 {
		newMetadata = json.RawMessage([]byte("{}"))
	}

	var highlightedParam any
	if newHighlighted != nil && strings.TrimSpace(*newHighlighted) != "" {
		highlightedParam = *newHighlighted
	} else {
		highlightedParam = nil
	}

	_, err := s.pool.Exec(ctx, `
 		UPDATE derived.simulated_calcuttas
 		SET name = $2,
 			description = $3,
 			highlighted_simulated_entry_id = $4::uuid,
 			metadata_json = $5::jsonb,
 			updated_at = NOW()
 		WHERE id = $1::uuid
 			AND deleted_at IS NULL
 	`, id, newName, nullUUIDParamPtr(newDescription), highlightedParam, []byte(newMetadata))
	if err != nil {
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

	ctx := r.Context()

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

	var exists bool
	if err := tx.QueryRow(ctx, `
 		SELECT EXISTS(
 			SELECT 1
 			FROM derived.simulated_calcuttas
 			WHERE id = $1::uuid
 				AND deleted_at IS NULL
 		)
 	`, id).Scan(&exists); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	if !exists {
		writeError(w, r, http.StatusNotFound, "not_found", "Simulated calcutta not found", "id")
		return
	}

	if _, err := tx.Exec(ctx, `
 		UPDATE derived.simulated_calcutta_payouts
 		SET deleted_at = NOW(),
 			updated_at = NOW()
 		WHERE simulated_calcutta_id = $1::uuid
 			AND deleted_at IS NULL
 	`, id); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	if _, err := tx.Exec(ctx, `
 		UPDATE derived.simulated_calcutta_scoring_rules
 		SET deleted_at = NOW(),
 			updated_at = NOW()
 		WHERE simulated_calcutta_id = $1::uuid
 			AND deleted_at IS NULL
 	`, id); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	for _, p := range req.Payouts {
		if _, err := tx.Exec(ctx, `
 			INSERT INTO derived.simulated_calcutta_payouts (
 				simulated_calcutta_id,
 				position,
 				amount_cents
 			)
 			VALUES ($1::uuid, $2::int, $3::int)
 		`, id, p.Position, p.AmountCents); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				writeError(w, r, http.StatusConflict, "conflict", "Duplicate payout position", "payouts")
				return
			}
			writeErrorFromErr(w, r, err)
			return
		}
	}

	for _, sr := range req.ScoringRules {
		if _, err := tx.Exec(ctx, `
 			INSERT INTO derived.simulated_calcutta_scoring_rules (
 				simulated_calcutta_id,
 				win_index,
 				points_awarded
 			)
 			VALUES ($1::uuid, $2::int, $3::int)
 		`, id, sr.WinIndex, sr.PointsAwarded); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				writeError(w, r, http.StatusConflict, "conflict", "Duplicate scoring rule winIndex", "scoringRules")
				return
			}
			writeErrorFromErr(w, r, err)
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	committed = true

	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}
