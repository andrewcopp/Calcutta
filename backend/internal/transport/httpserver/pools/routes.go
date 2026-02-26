package pools

import (
	"net/http"

	"github.com/gorilla/mux"
)

type Handlers struct {
	ListPools               http.HandlerFunc
	CreatePool              http.HandlerFunc
	GetPool                 http.HandlerFunc
	GetDashboard            http.HandlerFunc
	UpdatePool              http.HandlerFunc
	ListPortfolios          http.HandlerFunc
	CreatePortfolio         http.HandlerFunc
	CreateInvitation        http.HandlerFunc
	ListInvitations         http.HandlerFunc
	AcceptInvitation        http.HandlerFunc
	RevokeInvitation        http.HandlerFunc
	ListMyInvitations       http.HandlerFunc
	ListInvestments         http.HandlerFunc
	ListOwnership           http.HandlerFunc
	UpdatePortfolio         http.HandlerFunc
	Reinvite                http.HandlerFunc
	ListPayouts             http.HandlerFunc
	ReplacePayouts          http.HandlerFunc
}

const uuidPattern = `[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`

func RegisterRoutes(r *mux.Router, h Handlers) {
	r.HandleFunc("/api/v1/pools", h.ListPools).Methods("GET")
	r.HandleFunc("/api/v1/pools", h.CreatePool).Methods("POST")
	r.HandleFunc("/api/v1/pools/{id:"+uuidPattern+"}/dashboard", h.GetDashboard).Methods("GET")
	r.HandleFunc("/api/v1/pools/{id:"+uuidPattern+"}", h.GetPool).Methods("GET")
	r.HandleFunc("/api/v1/pools/{id:"+uuidPattern+"}", h.UpdatePool).Methods("PATCH")
	r.HandleFunc("/api/v1/pools/{id:"+uuidPattern+"}/portfolios", h.ListPortfolios).Methods("GET")
	r.HandleFunc("/api/v1/pools/{id:"+uuidPattern+"}/portfolios", h.CreatePortfolio).Methods("POST")
	r.HandleFunc("/api/v1/pools/{id:"+uuidPattern+"}/invitations", h.CreateInvitation).Methods("POST")
	r.HandleFunc("/api/v1/pools/{id:"+uuidPattern+"}/invitations", h.ListInvitations).Methods("GET")
	r.HandleFunc("/api/v1/pools/{id:"+uuidPattern+"}/invitations/{invitationId:"+uuidPattern+"}/accept", h.AcceptInvitation).Methods("POST")
	r.HandleFunc("/api/v1/pools/{id:"+uuidPattern+"}/invitations/{invitationId:"+uuidPattern+"}/revoke", h.RevokeInvitation).Methods("POST")
	r.HandleFunc("/api/v1/me/invitations", h.ListMyInvitations).Methods("GET")
	r.HandleFunc("/api/v1/pools/{id:"+uuidPattern+"}/reinvite", h.Reinvite).Methods("POST")
	r.HandleFunc("/api/v1/pools/{id:"+uuidPattern+"}/payouts", h.ListPayouts).Methods("GET")
	r.HandleFunc("/api/v1/pools/{id:"+uuidPattern+"}/payouts", h.ReplacePayouts).Methods("PUT")
	r.HandleFunc("/api/v1/pools/{poolId:"+uuidPattern+"}/portfolios/{portfolioId:"+uuidPattern+"}/investments", h.ListInvestments).Methods("GET")
	r.HandleFunc("/api/v1/pools/{poolId:"+uuidPattern+"}/portfolios/{portfolioId:"+uuidPattern+"}/ownership", h.ListOwnership).Methods("GET")
	r.HandleFunc("/api/v1/pools/{poolId:"+uuidPattern+"}/portfolios/{portfolioId:"+uuidPattern+"}", h.UpdatePortfolio).Methods("PATCH")
}
