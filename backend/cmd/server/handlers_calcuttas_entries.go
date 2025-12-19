package main

import (
	"net/http"

	"github.com/andrewcopp/Calcutta/backend/cmd/server/dtos"
	"github.com/gorilla/mux"
)

func (s *Server) calcuttaEntriesHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	calcuttaID := vars["id"]
	if calcuttaID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Calcutta ID is required", "id")
		return
	}

	entries, err := s.calcuttaService.GetEntries(r.Context(), calcuttaID)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, dtos.NewEntryListResponse(entries))
}

func (s *Server) entryTeamsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	entryID := vars["id"]
	if entryID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Entry ID is required", "id")
		return
	}

	teams, err := s.calcuttaService.GetEntryTeams(r.Context(), entryID)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, dtos.NewEntryTeamListResponse(teams))
}
