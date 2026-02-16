package calcuttas

import (
	"net/http"

	"github.com/gorilla/mux"
)

type Handlers struct {
	ListCalcuttas             http.HandlerFunc
	ListCalcuttasWithRankings http.HandlerFunc
	CreateCalcutta            http.HandlerFunc
	GetCalcutta               http.HandlerFunc
	GetDashboard              http.HandlerFunc
	UpdateCalcutta            http.HandlerFunc
	ListCalcuttaEntries       http.HandlerFunc
	CreateEntry               http.HandlerFunc
	CreateInvitation          http.HandlerFunc
	ListInvitations           http.HandlerFunc
	AcceptInvitation          http.HandlerFunc
	ListEntryTeams            http.HandlerFunc
	ListEntryPortfolios       http.HandlerFunc
	UpdateEntry               http.HandlerFunc
	Reinvite                  http.HandlerFunc
	ListPayouts               http.HandlerFunc
	ReplacePayouts            http.HandlerFunc
}

func RegisterRoutes(r *mux.Router, h Handlers) {
	r.HandleFunc("/api/calcuttas", h.ListCalcuttas).Methods("GET")
	r.HandleFunc("/api/calcuttas", h.CreateCalcutta).Methods("POST")
	r.HandleFunc("/api/calcuttas/list-with-rankings", h.ListCalcuttasWithRankings).Methods("GET")
	r.HandleFunc("/api/calcuttas/{id}/dashboard", h.GetDashboard).Methods("GET")
	r.HandleFunc("/api/calcuttas/{id}", h.GetCalcutta).Methods("GET")
	r.HandleFunc("/api/calcuttas/{id}", h.UpdateCalcutta).Methods("PATCH")
	r.HandleFunc("/api/calcuttas/{id}/entries", h.ListCalcuttaEntries).Methods("GET")
	r.HandleFunc("/api/calcuttas/{id}/entries", h.CreateEntry).Methods("POST")
	r.HandleFunc("/api/calcuttas/{id}/invitations", h.CreateInvitation).Methods("POST")
	r.HandleFunc("/api/calcuttas/{id}/invitations", h.ListInvitations).Methods("GET")
	r.HandleFunc("/api/calcuttas/{id}/invitations/{invitationId}/accept", h.AcceptInvitation).Methods("POST")
	r.HandleFunc("/api/calcuttas/{id}/reinvite", h.Reinvite).Methods("POST")
	r.HandleFunc("/api/calcuttas/{id}/payouts", h.ListPayouts).Methods("GET")
	r.HandleFunc("/api/calcuttas/{id}/payouts", h.ReplacePayouts).Methods("PUT")
	r.HandleFunc("/api/calcuttas/{calcuttaId}/entries/{entryId}/teams", h.ListEntryTeams).Methods("GET")
	r.HandleFunc("/api/entries/{id}/portfolios", h.ListEntryPortfolios).Methods("GET")
	r.HandleFunc("/api/entries/{id}", h.UpdateEntry).Methods("PATCH")
}
