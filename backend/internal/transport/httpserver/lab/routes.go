package lab

import (
	"net/http"

	"github.com/gorilla/mux"
)

// Handlers contains all lab-related HTTP handlers.
type Handlers struct {
	ListModels                   http.HandlerFunc
	GetModel                     http.HandlerFunc
	GetLeaderboard               http.HandlerFunc
	StartPipeline                http.HandlerFunc
	GetModelPipelineProgress     http.HandlerFunc
	GetPipelineRun               http.HandlerFunc
	CancelPipeline               http.HandlerFunc
	ListEntries                  http.HandlerFunc
	GetEntry                     http.HandlerFunc
	GetEntryByModelAndCalcutta   http.HandlerFunc
	ListEvaluations              http.HandlerFunc
	GetEvaluation                http.HandlerFunc
	GetEvaluationEntryResults    http.HandlerFunc
	GetEvaluationEntryProfile    http.HandlerFunc
	GetEvaluationSummary         http.HandlerFunc
}

// RegisterRoutes registers lab routes on the given router.
func RegisterRoutes(r *mux.Router, h Handlers) {
	// Models
	r.HandleFunc("/api/v1/lab/models", h.ListModels).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/v1/lab/models/leaderboard", h.GetLeaderboard).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/v1/lab/models/{id}/calcutta/{calcuttaId}/entry", h.GetEntryByModelAndCalcutta).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/v1/lab/models/{id}/pipeline/start", h.StartPipeline).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/v1/lab/models/{id}/pipeline/progress", h.GetModelPipelineProgress).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/v1/lab/models/{id}", h.GetModel).Methods("GET", "OPTIONS")

	// Pipeline runs
	r.HandleFunc("/api/v1/lab/pipeline-runs/{id}", h.GetPipelineRun).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/v1/lab/pipeline-runs/{id}/cancel", h.CancelPipeline).Methods("POST", "OPTIONS")

	// Entries
	r.HandleFunc("/api/v1/lab/entries", h.ListEntries).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/v1/lab/entries/{id}", h.GetEntry).Methods("GET", "OPTIONS")

	// Evaluations
	r.HandleFunc("/api/v1/lab/evaluations", h.ListEvaluations).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/v1/lab/evaluations/{id}/summary", h.GetEvaluationSummary).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/v1/lab/evaluations/{id}/entries", h.GetEvaluationEntryResults).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/v1/lab/evaluations/{id}", h.GetEvaluation).Methods("GET", "OPTIONS")

	// Entry Results (individual entry profile from an evaluation)
	r.HandleFunc("/api/v1/lab/entry-results/{id}", h.GetEvaluationEntryProfile).Methods("GET", "OPTIONS")
}
