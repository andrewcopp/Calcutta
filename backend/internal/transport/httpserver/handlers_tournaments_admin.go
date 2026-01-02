package httpserver

import (
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

	writeError(
		w,
		r,
		http.StatusGone,
		"gone",
		"Portfolios are fully derived and no longer require recalculation",
		"",
	)
}
