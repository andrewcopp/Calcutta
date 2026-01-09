package labcandidates

import (
	"net/http"

	"github.com/gorilla/mux"
)

type Handlers struct {
	ListCandidates      http.HandlerFunc
	CreateCandidates    http.HandlerFunc
	GetCandidateDetails http.HandlerFunc
	DeleteCandidate     http.HandlerFunc
}

func RegisterRoutes(r *mux.Router, h Handlers) {
	r.HandleFunc("/api/lab/candidates", h.ListCandidates).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/lab/candidates", h.CreateCandidates).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/lab/candidates/{candidateId}", h.GetCandidateDetails).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/lab/candidates/{candidateId}", h.DeleteCandidate).Methods("DELETE", "OPTIONS")
}
