package dtos

import (
	"time"

	"github.com/andrewcopp/Calcutta/backend/pkg/models"
)

type CreateCalcuttaRequest struct {
	Name         string `json:"name"`
	TournamentID string `json:"tournamentId"`
}

func (r *CreateCalcuttaRequest) Validate() error {
	if r.Name == "" {
		return ErrFieldRequired("name")
	}
	if r.TournamentID == "" {
		return ErrFieldRequired("tournamentId")
	}
	return nil
}

func (r *CreateCalcuttaRequest) ToModel() *models.Calcutta {
	return &models.Calcutta{
		Name:         r.Name,
		TournamentID: r.TournamentID,
	}
}

type CalcuttaResponse struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	TournamentID string    `json:"tournamentId"`
	Created      time.Time `json:"created"`
	Updated      time.Time `json:"updated"`
}

func NewCalcuttaResponse(c *models.Calcutta) *CalcuttaResponse {
	return &CalcuttaResponse{
		ID:           c.ID,
		Name:         c.Name,
		TournamentID: c.TournamentID,
		Created:      c.Created,
		Updated:      c.Updated,
	}
}

func NewCalcuttaListResponse(calcuttas []*models.Calcutta) []*CalcuttaResponse {
	if calcuttas == nil {
		return []*CalcuttaResponse{}
	}

	responses := make([]*CalcuttaResponse, len(calcuttas))
	for i, c := range calcuttas {
		responses[i] = NewCalcuttaResponse(c)
	}
	return responses
}
