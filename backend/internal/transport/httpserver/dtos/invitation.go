package dtos

import (
	"strings"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
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
	ID         string     `json:"id"`
	CalcuttaID string     `json:"calcuttaId"`
	UserID     string     `json:"userId"`
	InvitedBy  string     `json:"invitedBy"`
	Status     string     `json:"status"`
	RevokedAt  *time.Time `json:"revokedAt,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
}

func NewInvitationResponse(i *models.CalcuttaInvitation) *InvitationResponse {
	return &InvitationResponse{
		ID:         i.ID,
		CalcuttaID: i.CalcuttaID,
		UserID:     i.UserID,
		InvitedBy:  i.InvitedBy,
		Status:     i.Status,
		RevokedAt:  i.RevokedAt,
		CreatedAt:  i.CreatedAt,
		UpdatedAt:  i.UpdatedAt,
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
