package httpserver

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type syntheticEntryTeam struct {
	TeamID    string `json:"team_id"`
	BidPoints int    `json:"bid_points"`
}

type syntheticEntryListItem struct {
	ID            string               `json:"id"`
	CandidateID   string               `json:"candidate_id"`
	SnapshotEntry string               `json:"snapshot_entry_id"`
	EntryID       *string              `json:"entry_id,omitempty"`
	DisplayName   string               `json:"display_name"`
	IsSynthetic   bool                 `json:"is_synthetic"`
	Rank          *int                 `json:"rank"`
	Mean          *float64             `json:"mean_normalized_payout"`
	PTop1         *float64             `json:"p_top1"`
	PInMoney      *float64             `json:"p_in_money"`
	Teams         []syntheticEntryTeam `json:"teams"`
	CreatedAt     time.Time            `json:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at"`
}

type listSyntheticEntriesResponse struct {
	Items []syntheticEntryListItem `json:"items"`
}

type createSyntheticEntryRequest struct {
	DisplayName string               `json:"displayName"`
	Teams       []syntheticEntryTeam `json:"teams"`
}

type createSyntheticEntryResponse struct {
	ID string `json:"id"`
}

type importSyntheticEntryRequest struct {
	EntryArtifactID string  `json:"entryArtifactId"`
	DisplayName     *string `json:"displayName"`
}

type importSyntheticEntryResponse struct {
	ID     string `json:"id"`
	NTeams int    `json:"nTeams"`
}

type patchSyntheticEntryRequest struct {
	DisplayName *string               `json:"displayName"`
	Teams       *[]syntheticEntryTeam `json:"teams"`
}

func (s *Server) registerSyntheticEntryRoutes(r *mux.Router) {
	r.HandleFunc(
		"/api/synthetic-calcuttas/{id}/synthetic-entries",
		s.requirePermission("analytics.suite_scenarios.read", s.handleListSyntheticEntries),
	).Methods("GET", "OPTIONS")
	r.HandleFunc(
		"/api/synthetic-calcuttas/{id}/synthetic-entries",
		s.requirePermission("analytics.suite_scenarios.write", s.handleCreateSyntheticEntry),
	).Methods("POST", "OPTIONS")
	r.HandleFunc(
		"/api/synthetic-calcuttas/{id}/synthetic-entries/import",
		s.requirePermission("analytics.suite_scenarios.write", s.handleImportSyntheticEntry),
	).Methods("POST", "OPTIONS")

	// Candidate alias routes (preferred naming).
	r.HandleFunc(
		"/api/synthetic-calcuttas/{id}/candidates",
		s.requirePermission("analytics.suite_scenarios.read", s.handleListSyntheticEntries),
	).Methods("GET", "OPTIONS")
	r.HandleFunc(
		"/api/synthetic-calcuttas/{id}/candidates",
		s.requirePermission("analytics.suite_scenarios.write", s.handleCreateSyntheticEntry),
	).Methods("POST", "OPTIONS")
	r.HandleFunc(
		"/api/synthetic-calcuttas/{id}/candidates/import",
		s.requirePermission("analytics.suite_scenarios.write", s.handleImportSyntheticEntry),
	).Methods("POST", "OPTIONS")
	r.HandleFunc(
		"/api/synthetic-calcuttas/{syntheticCalcuttaId}/candidates/{id}",
		s.requirePermission("analytics.suite_scenarios.write", s.handlePatchSyntheticEntry),
	).Methods("PATCH", "OPTIONS")
	r.HandleFunc(
		"/api/synthetic-calcuttas/{syntheticCalcuttaId}/candidates/{id}",
		s.requirePermission("analytics.suite_scenarios.write", s.handleDeleteSyntheticEntry),
	).Methods("DELETE", "OPTIONS")

	// TODO: prefer nested routes long-term; keep flat resource routes for now.
	r.HandleFunc(
		"/api/synthetic-entries/{id}",
		s.requirePermission("analytics.suite_scenarios.write", s.handlePatchSyntheticEntry),
	).Methods("PATCH", "OPTIONS")
	r.HandleFunc(
		"/api/synthetic-entries/{id}",
		s.requirePermission("analytics.suite_scenarios.write", s.handleDeleteSyntheticEntry),
	).Methods("DELETE", "OPTIONS")

	// Candidate attachment alias routes.
	r.HandleFunc(
		"/api/synthetic-calcutta-candidates/{id}",
		s.requirePermission("analytics.suite_scenarios.write", s.handlePatchSyntheticEntry),
	).Methods("PATCH", "OPTIONS")
	r.HandleFunc(
		"/api/synthetic-calcutta-candidates/{id}",
		s.requirePermission("analytics.suite_scenarios.write", s.handleDeleteSyntheticEntry),
	).Methods("DELETE", "OPTIONS")
}

func (s *Server) handleListSyntheticEntries(w http.ResponseWriter, r *http.Request) {
	writeError(w, r, http.StatusGone, "gone", "Synthetic entry endpoints have been removed; use simulated-calcuttas entries", "")
}

func (s *Server) handleCreateSyntheticEntry(w http.ResponseWriter, r *http.Request) {
	writeError(w, r, http.StatusGone, "gone", "Synthetic entry write endpoints have been removed; use simulated-calcuttas entries", "")
}

func (s *Server) handleImportSyntheticEntry(w http.ResponseWriter, r *http.Request) {
	writeError(w, r, http.StatusGone, "gone", "Synthetic entry write endpoints have been removed; use simulated-calcuttas entries", "")
}

func (s *Server) handlePatchSyntheticEntry(w http.ResponseWriter, r *http.Request) {
	writeError(w, r, http.StatusGone, "gone", "Synthetic entry write endpoints have been removed; use simulated-calcuttas entries", "")
}

func (s *Server) handleDeleteSyntheticEntry(w http.ResponseWriter, r *http.Request) {
	writeError(w, r, http.StatusGone, "gone", "Synthetic entry write endpoints have been removed; use simulated-calcuttas entries", "")
}
