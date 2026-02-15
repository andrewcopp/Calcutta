package dtos

import (
	"strings"
	"time"

	"github.com/andrewcopp/Calcutta/backend/pkg/models"
)

type CreateInvitationRequest struct {
	UserID string `json:"userId"`
}

func (r *CreateInvitationRequest) Validate() error {
	if strings.TrimSpace(r.UserID) == "" {
		return ErrFieldRequired("userId")
	}
	return nil
}

type InvitationResponse struct {
	ID         string    `json:"id"`
	CalcuttaID string    `json:"calcuttaId"`
	UserID     string    `json:"userId"`
	InvitedBy  string    `json:"invitedBy"`
	Status     string    `json:"status"`
	Created    time.Time `json:"created"`
	Updated    time.Time `json:"updated"`
}

func NewInvitationResponse(i *models.CalcuttaInvitation) *InvitationResponse {
	return &InvitationResponse{
		ID:         i.ID,
		CalcuttaID: i.CalcuttaID,
		UserID:     i.UserID,
		InvitedBy:  i.InvitedBy,
		Status:     i.Status,
		Created:    i.Created,
		Updated:    i.Updated,
	}
}

func NewInvitationListResponse(invitations []*models.CalcuttaInvitation) []*InvitationResponse {
	if invitations == nil {
		return []*InvitationResponse{}
	}
	responses := make([]*InvitationResponse, len(invitations))
	for i, inv := range invitations {
		responses[i] = NewInvitationResponse(inv)
	}
	return responses
}
