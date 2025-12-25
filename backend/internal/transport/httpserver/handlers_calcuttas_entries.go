package httpserver

import (
	"net/http"

	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/dtos"
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

func (s *Server) calcuttaEntryTeamsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	calcuttaID := vars["calcuttaId"]
	entryID := vars["entryId"]

	if calcuttaID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Calcutta ID is required", "calcuttaId")
		return
	}
	if entryID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Entry ID is required", "entryId")
		return
	}

	teams, err := s.calcuttaService.GetEntryTeams(r.Context(), entryID)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, dtos.NewEntryTeamListResponse(teams))
}
