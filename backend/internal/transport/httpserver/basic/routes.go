package basic

import (
	"net/http"

	"github.com/gorilla/mux"
)

type Handlers struct {
	Health  http.HandlerFunc
	Ready   http.HandlerFunc
	Metrics http.HandlerFunc
}

type Options struct {
	MetricsEnabled bool
}

func RegisterRoutes(r *mux.Router, opts Options, h Handlers) {
	// Health & basic
	r.PathPrefix("/").Methods("OPTIONS").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r.HandleFunc("/healthz", h.Health).Methods("GET")
	r.HandleFunc("/readyz", h.Ready).Methods("GET")
	r.HandleFunc("/health/live", h.Health).Methods("GET")
	r.HandleFunc("/health/ready", h.Ready).Methods("GET")
	if opts.MetricsEnabled {
		r.HandleFunc("/metrics", h.Metrics).Methods("GET")
	}
	r.HandleFunc("/api/health", h.Health).Methods("GET")
	r.HandleFunc("/api/ready", h.Ready).Methods("GET")
}
