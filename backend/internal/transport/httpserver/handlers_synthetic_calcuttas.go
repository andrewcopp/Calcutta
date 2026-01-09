package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/app/synthetic_scenarios"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
)

type syntheticCalcuttaListItem struct {
	ID                        string          `json:"id"`
	CohortID                  string          `json:"cohort_id"`
	CalcuttaID                string          `json:"calcutta_id"`
	CalcuttaSnapshotID        *string         `json:"calcutta_snapshot_id,omitempty"`
	HighlightedEntryID        *string         `json:"highlighted_entry_id,omitempty"`
	FocusStrategyGenerationID *string         `json:"focus_strategy_generation_run_id,omitempty"`
	FocusEntryName            *string         `json:"focus_entry_name,omitempty"`
	LatestSimulationStatus    *string         `json:"latest_simulation_status,omitempty"`
	OurRank                   *int            `json:"our_rank,omitempty"`
	OurMeanNormalizedPayout   *float64        `json:"our_mean_normalized_payout,omitempty"`
	OurPTop1                  *float64        `json:"our_p_top1,omitempty"`
	OurPInMoney               *float64        `json:"our_p_in_money,omitempty"`
	TotalSimulations          *int            `json:"total_simulations,omitempty"`
	StartingStateKey          *string         `json:"starting_state_key,omitempty"`
	ExcludedEntryName         *string         `json:"excluded_entry_name,omitempty"`
	Notes                     *string         `json:"notes,omitempty"`
	Metadata                  json.RawMessage `json:"metadata"`
	CreatedAt                 time.Time       `json:"created_at"`
	UpdatedAt                 time.Time       `json:"updated_at"`
}

type listSyntheticCalcuttasResponse struct {
	Items []syntheticCalcuttaListItem `json:"items"`
}

type createSyntheticCalcuttaRequest struct {
	CohortID                  string  `json:"cohortId"`
	CalcuttaID                string  `json:"calcuttaId"`
	SourceCalcuttaID          *string `json:"sourceCalcuttaId"`
	CalcuttaSnapshotID        *string `json:"calcuttaSnapshotId"`
	FocusStrategyGenerationID *string `json:"focusStrategyGenerationRunId"`
	FocusEntryName            *string `json:"focusEntryName"`
	StartingStateKey          *string `json:"startingStateKey"`
	ExcludedEntryName         *string `json:"excludedEntryName"`
}

type createSyntheticCalcuttaResponse struct {
	ID string `json:"id"`
}

type patchSyntheticCalcuttaRequest struct {
	HighlightedEntryID *string          `json:"highlightedEntryId"`
	Notes              *string          `json:"notes"`
	Metadata           *json.RawMessage `json:"metadata"`
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
	r.HandleFunc(
		"/api/synthetic-calcuttas/{id}",
		s.requirePermission("analytics.suite_scenarios.write", s.handlePatchSyntheticCalcutta),
	).Methods("PATCH", "OPTIONS")
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

	var cohortIDPtr *string
	if cohortID != "" {
		v := cohortID
		cohortIDPtr = &v
	}
	var calcuttaIDPtr *string
	if calcuttaID != "" {
		v := calcuttaID
		calcuttaIDPtr = &v
	}

	items, err := s.app.SyntheticScenarios.ListSyntheticCalcuttas(r.Context(), cohortIDPtr, calcuttaIDPtr, limit, offset)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	resp := make([]syntheticCalcuttaListItem, 0, len(items))
	for _, it := range items {
		resp = append(resp, syntheticCalcuttaListItem{
			ID:                        it.ID,
			CohortID:                  it.CohortID,
			CalcuttaID:                it.CalcuttaID,
			CalcuttaSnapshotID:        it.CalcuttaSnapshotID,
			HighlightedEntryID:        it.HighlightedEntryID,
			FocusStrategyGenerationID: it.FocusStrategyGenerationID,
			FocusEntryName:            it.FocusEntryName,
			LatestSimulationStatus:    it.LatestSimulationStatus,
			OurRank:                   it.OurRank,
			OurMeanNormalizedPayout:   it.OurMeanNormalizedPayout,
			OurPTop1:                  it.OurPTop1,
			OurPInMoney:               it.OurPInMoney,
			TotalSimulations:          it.TotalSimulations,
			StartingStateKey:          it.StartingStateKey,
			ExcludedEntryName:         it.ExcludedEntryName,
			Notes:                     it.Notes,
			Metadata:                  it.Metadata,
			CreatedAt:                 it.CreatedAt,
			UpdatedAt:                 it.UpdatedAt,
		})
	}
	writeJSON(w, http.StatusOK, listSyntheticCalcuttasResponse{Items: resp})
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

	it, err := s.app.SyntheticScenarios.GetSyntheticCalcutta(r.Context(), id)
	if err != nil {
		if errors.Is(err, synthetic_scenarios.ErrSyntheticCalcuttaNotFound) {
			writeError(w, r, http.StatusNotFound, "not_found", "Synthetic calcutta not found", "id")
			return
		}
		writeErrorFromErr(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, syntheticCalcuttaListItem{
		ID:                        it.ID,
		CohortID:                  it.CohortID,
		CalcuttaID:                it.CalcuttaID,
		CalcuttaSnapshotID:        it.CalcuttaSnapshotID,
		HighlightedEntryID:        it.HighlightedEntryID,
		FocusStrategyGenerationID: it.FocusStrategyGenerationID,
		FocusEntryName:            it.FocusEntryName,
		StartingStateKey:          it.StartingStateKey,
		ExcludedEntryName:         it.ExcludedEntryName,
		Notes:                     it.Notes,
		Metadata:                  it.Metadata,
		CreatedAt:                 it.CreatedAt,
		UpdatedAt:                 it.UpdatedAt,
	})
}

func (s *Server) handlePatchSyntheticCalcutta(w http.ResponseWriter, r *http.Request) {
	writeError(w, r, http.StatusGone, "gone", "Synthetic calcutta write endpoints have been removed; use simulated-calcuttas", "")
	return
}

func (s *Server) handleCreateSyntheticCalcutta(w http.ResponseWriter, r *http.Request) {
	writeError(w, r, http.StatusGone, "gone", "Synthetic calcutta write endpoints have been removed; use simulated-calcuttas", "")
	return
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
			FROM derived.strategy_generation_run_bids reb
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
