package lab

import (
	"net/http"

	"github.com/gorilla/mux"
)

// Handlers contains all lab-related HTTP handlers.
type Handlers struct {
	ListModels      http.HandlerFunc
	GetModel        http.HandlerFunc
	GetLeaderboard  http.HandlerFunc
	ListEntries     http.HandlerFunc
	GetEntry        http.HandlerFunc
	ListEvaluations http.HandlerFunc
	GetEvaluation   http.HandlerFunc
}

// RegisterRoutes registers lab routes on the given router.
func RegisterRoutes(r *mux.Router, h Handlers) {
	// Models
	r.HandleFunc("/api/lab/models", h.ListModels).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/lab/models/leaderboard", h.GetLeaderboard).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/lab/models/{id}", h.GetModel).Methods("GET", "OPTIONS")

	// Entries
	r.HandleFunc("/api/lab/entries", h.ListEntries).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/lab/entries/{id}", h.GetEntry).Methods("GET", "OPTIONS")

	// Evaluations
	r.HandleFunc("/api/lab/evaluations", h.ListEvaluations).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/lab/evaluations/{id}", h.GetEvaluation).Methods("GET", "OPTIONS")
}
