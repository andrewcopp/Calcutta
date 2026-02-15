package calcuttas

import (
	"net/http"

	"github.com/gorilla/mux"
)

type Handlers struct {
	ListCalcuttas       http.HandlerFunc
	CreateCalcutta      http.HandlerFunc
	GetCalcutta         http.HandlerFunc
	UpdateCalcutta      http.HandlerFunc
	ListCalcuttaEntries http.HandlerFunc
	CreateEntry         http.HandlerFunc
	CreateInvitation    http.HandlerFunc
	ListInvitations     http.HandlerFunc
	AcceptInvitation    http.HandlerFunc
	ListEntryTeams      http.HandlerFunc
	ListEntryPortfolios http.HandlerFunc
	UpdateEntry         http.HandlerFunc
}

func RegisterRoutes(r *mux.Router, h Handlers) {
	r.HandleFunc("/api/calcuttas", h.ListCalcuttas).Methods("GET")
	r.HandleFunc("/api/calcuttas", h.CreateCalcutta).Methods("POST")
	r.HandleFunc("/api/calcuttas/{id}", h.GetCalcutta).Methods("GET")
	r.HandleFunc("/api/calcuttas/{id}", h.UpdateCalcutta).Methods("PATCH")
	r.HandleFunc("/api/calcuttas/{id}/entries", h.ListCalcuttaEntries).Methods("GET")
	r.HandleFunc("/api/calcuttas/{id}/entries", h.CreateEntry).Methods("POST")
	r.HandleFunc("/api/calcuttas/{id}/invitations", h.CreateInvitation).Methods("POST")
	r.HandleFunc("/api/calcuttas/{id}/invitations", h.ListInvitations).Methods("GET")
	r.HandleFunc("/api/calcuttas/{id}/invitations/{invitationId}/accept", h.AcceptInvitation).Methods("POST")
	r.HandleFunc("/api/calcuttas/{calcuttaId}/entries/{entryId}/teams", h.ListEntryTeams).Methods("GET")
	r.HandleFunc("/api/entries/{id}/portfolios", h.ListEntryPortfolios).Methods("GET")
	r.HandleFunc("/api/entries/{id}", h.UpdateEntry).Methods("PATCH")
}
