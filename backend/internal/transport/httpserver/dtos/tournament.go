package dtos

import (
	"strings"
	"time"

	"github.com/andrewcopp/Calcutta/backend/pkg/models"
)

type CreateTournamentRequest struct {
	Name   string `json:"name"`
	Rounds int    `json:"rounds"`
}

func (r *CreateTournamentRequest) Validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return ErrFieldRequired("name")
	}
	if r.Rounds <= 0 {
		return ErrFieldInvalid("rounds", "must be greater than 0")
	}
	return nil
}

type TournamentResponse struct {
	ID      string    `json:"id"`
	Name    string    `json:"name"`
	Rounds  int       `json:"rounds"`
	Winner  string    `json:"winner,omitempty"`
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}

func NewTournamentResponse(t *models.Tournament, winner string) *TournamentResponse {
	return &TournamentResponse{
		ID:      t.ID,
		Name:    t.Name,
		Rounds:  t.Rounds,
		Winner:  winner,
		Created: t.Created,
		Updated: t.Updated,
	}
}

type CreateTournamentTeamRequest struct {
	SchoolID string `json:"schoolId"`
	Seed     int    `json:"seed"`
	Region   string `json:"region"`
}

func (r *CreateTournamentTeamRequest) Validate() error {
	if strings.TrimSpace(r.SchoolID) == "" {
		return ErrFieldRequired("schoolId")
	}
	if r.Seed < 1 || r.Seed > 16 {
		return ErrFieldInvalid("seed", "must be between 1 and 16")
	}
	return nil
}

type UpdateTournamentTeamRequest struct {
	Wins       *int  `json:"wins,omitempty"`
	Byes       *int  `json:"byes,omitempty"`
	Eliminated *bool `json:"eliminated,omitempty"`
}

type TournamentTeamResponse struct {
	ID           string          `json:"id"`
	TournamentID string          `json:"tournamentId"`
	SchoolID     string          `json:"schoolId"`
	Seed         int             `json:"seed"`
	Region       string          `json:"region"`
	Byes         int             `json:"byes"`
	Wins         int             `json:"wins"`
	Eliminated   bool            `json:"eliminated"`
	Created      time.Time       `json:"created"`
	Updated      time.Time       `json:"updated"`
	School       *SchoolResponse `json:"school,omitempty"`
}

func NewTournamentTeamResponse(t *models.TournamentTeam, school *models.School) *TournamentTeamResponse {
	resp := &TournamentTeamResponse{
		ID:           t.ID,
		TournamentID: t.TournamentID,
		SchoolID:     t.SchoolID,
		Seed:         t.Seed,
		Region:       t.Region,
		Byes:         t.Byes,
		Wins:         t.Wins,
		Eliminated:   t.Eliminated,
		Created:      t.Created,
		Updated:      t.Updated,
	}
	if school != nil {
		resp.School = NewSchoolResponse(school)
	}
	return resp
}
