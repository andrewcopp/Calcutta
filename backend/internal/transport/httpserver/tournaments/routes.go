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
}

func RegisterRoutes(r *mux.Router, h Handlers) {
	r.HandleFunc("/api/tournaments", h.ListTournaments).Methods("GET")
	r.HandleFunc("/api/tournaments/{id}", h.GetTournament).Methods("GET")
	r.HandleFunc("/api/tournaments", h.CreateTournament).Methods("POST")
	r.HandleFunc("/api/tournaments/{id}", h.UpdateTournament).Methods("PATCH")
	r.HandleFunc("/api/tournaments/{id}/teams", h.ListTournamentTeams).Methods("GET")
	r.HandleFunc("/api/tournaments/{id}/teams", h.CreateTournamentTeam).Methods("POST")
	r.HandleFunc("/api/tournaments/{id}/teams", h.ReplaceTeams).Methods("PUT")
	r.HandleFunc("/api/tournaments/{tournamentId}/teams/{teamId}", h.UpdateTeam).Methods("PATCH", "OPTIONS")
	r.HandleFunc("/api/tournaments/{id}/kenpom", h.UpdateKenPomStats).Methods("PUT")
	r.HandleFunc("/api/competitions", h.ListCompetitions).Methods("GET")
	r.HandleFunc("/api/seasons", h.ListSeasons).Methods("GET")
}
