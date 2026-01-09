package httpserver

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
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
	writeError(w, r, http.StatusGone, "gone", "Synthetic calcutta endpoints have been removed; use simulated-calcuttas", "")
}

func (s *Server) handleGetSyntheticCalcutta(w http.ResponseWriter, r *http.Request) {
	writeError(w, r, http.StatusGone, "gone", "Synthetic calcutta endpoints have been removed; use simulated-calcuttas", "")
}

func (s *Server) handlePatchSyntheticCalcutta(w http.ResponseWriter, r *http.Request) {
	writeError(w, r, http.StatusGone, "gone", "Synthetic calcutta write endpoints have been removed; use simulated-calcuttas", "")
}

func (s *Server) handleCreateSyntheticCalcutta(w http.ResponseWriter, r *http.Request) {
	writeError(w, r, http.StatusGone, "gone", "Synthetic calcutta write endpoints have been removed; use simulated-calcuttas", "")
}
