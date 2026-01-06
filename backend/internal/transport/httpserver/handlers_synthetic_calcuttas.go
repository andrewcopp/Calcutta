package httpserver

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
)

type syntheticCalcuttaListItem struct {
	ID                        string    `json:"id"`
	SuiteID                   string    `json:"suite_id"`
	CalcuttaID                string    `json:"calcutta_id"`
	CalcuttaSnapshotID        *string   `json:"calcutta_snapshot_id,omitempty"`
	FocusStrategyGenerationID *string   `json:"focus_strategy_generation_run_id,omitempty"`
	FocusEntryName            *string   `json:"focus_entry_name,omitempty"`
	StartingStateKey          *string   `json:"starting_state_key,omitempty"`
	ExcludedEntryName         *string   `json:"excluded_entry_name,omitempty"`
	CreatedAt                 time.Time `json:"created_at"`
	UpdatedAt                 time.Time `json:"updated_at"`
}

type listSyntheticCalcuttasResponse struct {
	Items []syntheticCalcuttaListItem `json:"items"`
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

type createSyntheticCalcuttaResponse struct {
	ID string `json:"id"`
}

func (s *Server) registerSyntheticCalcuttaRoutes(r *mux.Router) {
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
}

func (s *Server) handleListSyntheticCalcuttas(w http.ResponseWriter, r *http.Request) {
	cohortID := strings.TrimSpace(r.URL.Query().Get("cohort_id"))
	calcuttaID := strings.TrimSpace(r.URL.Query().Get("calcutta_id"))

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
			sc.id::text,
			sc.cohort_id::text,
			sc.calcutta_id::text,
			sc.calcutta_snapshot_id::text,
			sc.focus_strategy_generation_run_id::text,
			sc.focus_entry_name,
			sc.starting_state_key,
			sc.excluded_entry_name,
			sc.created_at,
			sc.updated_at
		FROM derived.synthetic_calcuttas sc
		WHERE sc.deleted_at IS NULL
			AND ($1::uuid IS NULL OR sc.cohort_id = $1::uuid)
			AND ($2::uuid IS NULL OR sc.calcutta_id = $2::uuid)
		ORDER BY sc.created_at DESC
		LIMIT $3::int
		OFFSET $4::int
	`, nullUUIDParam(cohortID), nullUUIDParam(calcuttaID), limit, offset)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	defer rows.Close()

	items := make([]syntheticCalcuttaListItem, 0)
	for rows.Next() {
		var it syntheticCalcuttaListItem
		if err := rows.Scan(
			&it.ID,
			&it.SuiteID,
			&it.CalcuttaID,
			&it.CalcuttaSnapshotID,
			&it.FocusStrategyGenerationID,
			&it.FocusEntryName,
			&it.StartingStateKey,
			&it.ExcludedEntryName,
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

	writeJSON(w, http.StatusOK, listSyntheticCalcuttasResponse{Items: items})
}

func (s *Server) handleGetSyntheticCalcutta(w http.ResponseWriter, r *http.Request) {
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

	var it syntheticCalcuttaListItem
	if err := s.pool.QueryRow(r.Context(), `
		SELECT
			sc.id::text,
			sc.cohort_id::text,
			sc.calcutta_id::text,
			sc.calcutta_snapshot_id::text,
			sc.focus_strategy_generation_run_id::text,
			sc.focus_entry_name,
			sc.starting_state_key,
			sc.excluded_entry_name,
			sc.created_at,
			sc.updated_at
		FROM derived.synthetic_calcuttas sc
		WHERE sc.id = $1::uuid
			AND sc.deleted_at IS NULL
		LIMIT 1
	`, id).Scan(
		&it.ID,
		&it.SuiteID,
		&it.CalcuttaID,
		&it.CalcuttaSnapshotID,
		&it.FocusStrategyGenerationID,
		&it.FocusEntryName,
		&it.StartingStateKey,
		&it.ExcludedEntryName,
		&it.CreatedAt,
		&it.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, r, http.StatusNotFound, "not_found", "Synthetic calcutta not found", "id")
			return
		}
		writeErrorFromErr(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, it)
}

func (s *Server) handleCreateSyntheticCalcutta(w http.ResponseWriter, r *http.Request) {
	var req createSyntheticCalcuttaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}

	req.CohortID = strings.TrimSpace(req.CohortID)
	req.CalcuttaID = strings.TrimSpace(req.CalcuttaID)
	if req.CohortID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "cohortId is required", "cohortId")
		return
	}
	if _, err := uuid.Parse(req.CohortID); err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "cohortId must be a valid UUID", "cohortId")
		return
	}
	if req.CalcuttaID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "calcuttaId is required", "calcuttaId")
		return
	}
	if _, err := uuid.Parse(req.CalcuttaID); err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "calcuttaId must be a valid UUID", "calcuttaId")
		return
	}

	startingStateKey := (*string)(nil)
	if req.StartingStateKey != nil {
		v := strings.TrimSpace(*req.StartingStateKey)
		if v != "" {
			if v != "post_first_four" && v != "current" {
				writeError(w, r, http.StatusBadRequest, "validation_error", "startingStateKey must be 'current' or 'post_first_four'", "startingStateKey")
				return
			}
			startingStateKey = &v
		}
	}

	excludedEntry := (*string)(nil)
	if req.ExcludedEntryName != nil {
		v := strings.TrimSpace(*req.ExcludedEntryName)
		if v != "" {
			excludedEntry = &v
		}
	}

	focusEntryName := (*string)(nil)
	if req.FocusEntryName != nil {
		v := strings.TrimSpace(*req.FocusEntryName)
		if v != "" {
			focusEntryName = &v
		}
	}

	focusStrategyRunID := ""
	if req.FocusStrategyGenerationID != nil {
		v := strings.TrimSpace(*req.FocusStrategyGenerationID)
		if v != "" {
			if _, err := uuid.Parse(v); err != nil {
				writeError(w, r, http.StatusBadRequest, "validation_error", "focusStrategyGenerationRunId must be a valid UUID", "focusStrategyGenerationRunId")
				return
			}
			focusStrategyRunID = v
		}
	}

	providedSnapshotID := ""
	if req.CalcuttaSnapshotID != nil {
		v := strings.TrimSpace(*req.CalcuttaSnapshotID)
		if v != "" {
			if _, err := uuid.Parse(v); err != nil {
				writeError(w, r, http.StatusBadRequest, "validation_error", "calcuttaSnapshotId must be a valid UUID", "calcuttaSnapshotId")
				return
			}
			providedSnapshotID = v
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

	snapshotID := (*string)(nil)
	if providedSnapshotID != "" {
		snapshotID = &providedSnapshotID
	} else {
		created, err := createSyntheticCalcuttaSnapshot(ctx, tx, req.CalcuttaID, excludedEntry, focusStrategyRunID, focusEntryName)
		if err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
		snapshotID = &created
	}

	var syntheticID string
	if err := tx.QueryRow(ctx, `
		UPDATE derived.synthetic_calcuttas
		SET calcutta_snapshot_id = $3::uuid,
			focus_strategy_generation_run_id = $4::uuid,
			focus_entry_name = $5,
			starting_state_key = $6,
			excluded_entry_name = $7,
			updated_at = NOW(),
			deleted_at = NULL
		WHERE cohort_id = $1::uuid
			AND calcutta_id = $2::uuid
			AND deleted_at IS NULL
		RETURNING id::text
	`, req.CohortID, req.CalcuttaID, snapshotID, nullUUIDParam(focusStrategyRunID), focusEntryName, startingStateKey, excludedEntry).Scan(&syntheticID); err != nil {
		if err == pgx.ErrNoRows {
			if err := tx.QueryRow(ctx, `
				INSERT INTO derived.synthetic_calcuttas (
					cohort_id,
					calcutta_id,
					calcutta_snapshot_id,
					focus_strategy_generation_run_id,
					focus_entry_name,
					starting_state_key,
					excluded_entry_name
				)
				VALUES ($1::uuid, $2::uuid, $3::uuid, $4::uuid, $5, $6, $7)
				RETURNING id::text
			`, req.CohortID, req.CalcuttaID, snapshotID, nullUUIDParam(focusStrategyRunID), focusEntryName, startingStateKey, excludedEntry).Scan(&syntheticID); err != nil {
				writeErrorFromErr(w, r, err)
				return
			}
		} else {
			writeErrorFromErr(w, r, err)
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	committed = true

	writeJSON(w, http.StatusCreated, createSyntheticCalcuttaResponse{ID: syntheticID})
}

func createSyntheticCalcuttaSnapshot(
	ctx context.Context,
	tx pgx.Tx,
	calcuttaID string,
	excludedEntryName *string,
	focusStrategyGenerationRunID string,
	focusEntryName *string,
) (string, error) {
	// Reuse suite scenario snapshot implementation. The stored snapshot_type differs only for human readability.
	var snapshotID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO core.calcutta_snapshots (base_calcutta_id, snapshot_type, description)
		VALUES ($1::uuid, 'synthetic_calcutta', 'Synthetic calcutta snapshot')
		RETURNING id
	`, calcuttaID).Scan(&snapshotID); err != nil {
		return "", err
	}

	_, err := tx.Exec(ctx, `
		INSERT INTO core.calcutta_snapshot_payouts (calcutta_snapshot_id, position, amount_cents)
		SELECT $2, position, amount_cents
		FROM core.payouts
		WHERE calcutta_id = $1
			AND deleted_at IS NULL
	`, calcuttaID, snapshotID)
	if err != nil {
		return "", err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO core.calcutta_snapshot_scoring_rules (calcutta_snapshot_id, win_index, points_awarded)
		SELECT $2, win_index, points_awarded
		FROM core.calcutta_scoring_rules
		WHERE calcutta_id = $1
			AND deleted_at IS NULL
	`, calcuttaID, snapshotID)
	if err != nil {
		return "", err
	}

	excluded := ""
	if excludedEntryName != nil {
		excluded = *excludedEntryName
	}

	entryRows, err := tx.Query(ctx, `
		SELECT id::text, name
		FROM core.entries
		WHERE calcutta_id = $1::uuid
			AND deleted_at IS NULL
			AND (name != $2 OR $2 = '')
		ORDER BY created_at ASC
	`, calcuttaID, excluded)
	if err != nil {
		return "", err
	}
	defer entryRows.Close()

	type entryRow struct {
		id   string
		name string
	}
	entries := make([]entryRow, 0)
	for entryRows.Next() {
		var id, name string
		if err := entryRows.Scan(&id, &name); err != nil {
			return "", err
		}
		entries = append(entries, entryRow{id: id, name: name})
	}
	if err := entryRows.Err(); err != nil {
		return "", err
	}

	for _, e := range entries {
		var snapshotEntryID string
		if err := tx.QueryRow(ctx, `
			INSERT INTO core.calcutta_snapshot_entries (calcutta_snapshot_id, entry_id, display_name, is_synthetic)
			VALUES ($1::uuid, $2::uuid, $3, false)
			RETURNING id
		`, snapshotID, e.id, e.name).Scan(&snapshotEntryID); err != nil {
			return "", err
		}

		_, err := tx.Exec(ctx, `
			INSERT INTO core.calcutta_snapshot_entry_teams (calcutta_snapshot_entry_id, team_id, bid_points)
			SELECT $1::uuid, team_id, bid_points
			FROM core.entry_teams
			WHERE entry_id = $2::uuid
				AND deleted_at IS NULL
		`, snapshotEntryID, e.id)
		if err != nil {
			return "", err
		}
	}

	if focusStrategyGenerationRunID != "" {
		name := "Our Strategy"
		if focusEntryName != nil && strings.TrimSpace(*focusEntryName) != "" {
			name = strings.TrimSpace(*focusEntryName)
		}

		var snapshotEntryID string
		if err := tx.QueryRow(ctx, `
			INSERT INTO core.calcutta_snapshot_entries (calcutta_snapshot_id, entry_id, display_name, is_synthetic)
			VALUES ($1::uuid, NULL, $2, true)
			RETURNING id
		`, snapshotID, name).Scan(&snapshotEntryID); err != nil {
			return "", err
		}

		ct, err := tx.Exec(ctx, `
			INSERT INTO core.calcutta_snapshot_entry_teams (calcutta_snapshot_entry_id, team_id, bid_points)
			SELECT $1::uuid, reb.team_id, reb.bid_points
			FROM derived.recommended_entry_bids reb
			WHERE reb.strategy_generation_run_id = $2::uuid
				AND reb.deleted_at IS NULL
		`, snapshotEntryID, focusStrategyGenerationRunID)
		if err != nil {
			return "", err
		}
		if ct.RowsAffected() == 0 {
			return "", pgx.ErrNoRows
		}
	}

	return snapshotID, nil
}
