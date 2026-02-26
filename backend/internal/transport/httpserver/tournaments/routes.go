package tournaments

import (
	"net/http"

	"github.com/gorilla/mux"
)

type Handlers struct {
	ListTournaments      http.HandlerFunc
	GetTournament        http.HandlerFunc
	CreateTournament     http.HandlerFunc
	UpdateTournament     http.HandlerFunc
	ListTournamentTeams  http.HandlerFunc
	CreateTournamentTeam http.HandlerFunc
	UpdateTeam           http.HandlerFunc
	ListCompetitions     http.HandlerFunc
	ListSeasons          http.HandlerFunc
	ReplaceTeams         http.HandlerFunc
	UpdateKenPomStats    http.HandlerFunc
	GetPredictions           http.HandlerFunc
	ListPredictionBatches    http.HandlerFunc
}

func RegisterRoutes(r *mux.Router, h Handlers) {
	r.HandleFunc("/api/v1/tournaments", h.ListTournaments).Methods("GET")
	r.HandleFunc("/api/v1/tournaments/{id}", h.GetTournament).Methods("GET")
	r.HandleFunc("/api/v1/tournaments", h.CreateTournament).Methods("POST")
	r.HandleFunc("/api/v1/tournaments/{id}", h.UpdateTournament).Methods("PATCH")
	r.HandleFunc("/api/v1/tournaments/{tournamentId}/teams", h.ListTournamentTeams).Methods("GET")
	r.HandleFunc("/api/v1/tournaments/{tournamentId}/teams", h.CreateTournamentTeam).Methods("POST")
	r.HandleFunc("/api/v1/tournaments/{tournamentId}/teams", h.ReplaceTeams).Methods("PUT")
	r.HandleFunc("/api/v1/tournaments/{tournamentId}/teams/{teamId}", h.UpdateTeam).Methods("PATCH", "OPTIONS")
	r.HandleFunc("/api/v1/tournaments/{tournamentId}/kenpom", h.UpdateKenPomStats).Methods("PUT")
	r.HandleFunc("/api/v1/tournaments/{tournamentId}/predictions", h.GetPredictions).Methods("GET")
	r.HandleFunc("/api/v1/tournaments/{tournamentId}/prediction-batches", h.ListPredictionBatches).Methods("GET")
	r.HandleFunc("/api/v1/competitions", h.ListCompetitions).Methods("GET")
	r.HandleFunc("/api/v1/seasons", h.ListSeasons).Methods("GET")
}
