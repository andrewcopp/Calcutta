package calcuttas

import (
	"net/http"

	"github.com/gorilla/mux"
)

type Handlers struct {
	ListCalcuttas             http.HandlerFunc
	CreateCalcutta            http.HandlerFunc
	GetCalcutta               http.HandlerFunc
	GetDashboard              http.HandlerFunc
	UpdateCalcutta            http.HandlerFunc
	ListCalcuttaEntries       http.HandlerFunc
	CreateEntry               http.HandlerFunc
	CreateInvitation          http.HandlerFunc
	ListInvitations           http.HandlerFunc
	AcceptInvitation          http.HandlerFunc
	RevokeInvitation          http.HandlerFunc
	ListMyInvitations         http.HandlerFunc
	ListEntryTeams            http.HandlerFunc
	ListEntryPortfolios       http.HandlerFunc
	UpdateEntry               http.HandlerFunc
	Reinvite                  http.HandlerFunc
	ListPayouts               http.HandlerFunc
	ReplacePayouts            http.HandlerFunc
}

const uuidPattern = `[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`

func RegisterRoutes(r *mux.Router, h Handlers) {
	r.HandleFunc("/api/v1/calcuttas", h.ListCalcuttas).Methods("GET")
	r.HandleFunc("/api/v1/calcuttas", h.CreateCalcutta).Methods("POST")
	r.HandleFunc("/api/v1/calcuttas/{id:"+uuidPattern+"}/dashboard", h.GetDashboard).Methods("GET")
	r.HandleFunc("/api/v1/calcuttas/{id:"+uuidPattern+"}", h.GetCalcutta).Methods("GET")
	r.HandleFunc("/api/v1/calcuttas/{id:"+uuidPattern+"}", h.UpdateCalcutta).Methods("PATCH")
	r.HandleFunc("/api/v1/calcuttas/{id:"+uuidPattern+"}/entries", h.ListCalcuttaEntries).Methods("GET")
	r.HandleFunc("/api/v1/calcuttas/{id:"+uuidPattern+"}/entries", h.CreateEntry).Methods("POST")
	r.HandleFunc("/api/v1/calcuttas/{id:"+uuidPattern+"}/invitations", h.CreateInvitation).Methods("POST")
	r.HandleFunc("/api/v1/calcuttas/{id:"+uuidPattern+"}/invitations", h.ListInvitations).Methods("GET")
	r.HandleFunc("/api/v1/calcuttas/{id:"+uuidPattern+"}/invitations/{invitationId:"+uuidPattern+"}/accept", h.AcceptInvitation).Methods("POST")
	r.HandleFunc("/api/v1/calcuttas/{id:"+uuidPattern+"}/invitations/{invitationId:"+uuidPattern+"}/revoke", h.RevokeInvitation).Methods("POST")
	r.HandleFunc("/api/v1/me/invitations", h.ListMyInvitations).Methods("GET")
	r.HandleFunc("/api/v1/calcuttas/{id:"+uuidPattern+"}/reinvite", h.Reinvite).Methods("POST")
	r.HandleFunc("/api/v1/calcuttas/{id:"+uuidPattern+"}/payouts", h.ListPayouts).Methods("GET")
	r.HandleFunc("/api/v1/calcuttas/{id:"+uuidPattern+"}/payouts", h.ReplacePayouts).Methods("PUT")
	r.HandleFunc("/api/v1/calcuttas/{calcuttaId:"+uuidPattern+"}/entries/{entryId:"+uuidPattern+"}/teams", h.ListEntryTeams).Methods("GET")
	r.HandleFunc("/api/v1/calcuttas/{calcuttaId:"+uuidPattern+"}/entries/{entryId:"+uuidPattern+"}/portfolios", h.ListEntryPortfolios).Methods("GET")
	r.HandleFunc("/api/v1/calcuttas/{calcuttaId:"+uuidPattern+"}/entries/{entryId:"+uuidPattern+"}", h.UpdateEntry).Methods("PATCH")
}
