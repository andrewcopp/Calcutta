package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func (s *Server) recalculatePortfoliosHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	tournamentID := vars["id"]
	if tournamentID == "" {
		http.Error(w, "Tournament ID is required", http.StatusBadRequest)
		return
	}

	calcuttas, err := s.calcuttaService.GetCalcuttasByTournament(r.Context(), tournamentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, calcutta := range calcuttas {
		if err := s.calcuttaService.RecalculatePortfolio(r.Context(), calcutta.ID); err != nil {
			log.Printf("Error recalculating portfolio for calcutta %s: %v", calcutta.ID, err)
			continue
		}
	}

	w.WriteHeader(http.StatusOK)
}
