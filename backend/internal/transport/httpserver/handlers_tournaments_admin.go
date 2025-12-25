package httpserver

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func (s *Server) recalculatePortfoliosHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tournamentID := vars["id"]
	if tournamentID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Tournament ID is required", "id")
		return
	}

	calcuttas, err := s.calcuttaService.GetCalcuttasByTournament(r.Context(), tournamentID)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	for _, calcutta := range calcuttas {
		if err := s.calcuttaService.RecalculatePortfolio(r.Context(), calcutta.ID); err != nil {
			log.Printf("Error recalculating portfolio for calcutta %s: %v", calcutta.ID, err)
			continue
		}
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
