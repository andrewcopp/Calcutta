package dtos

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

type CreateTournamentRequest struct {
	Competition string `json:"competition"`
	Year        int    `json:"year"`
	Rounds      int    `json:"rounds"`
}

func (r *CreateTournamentRequest) Validate() error {
	if strings.TrimSpace(r.Competition) == "" {
		return ErrFieldRequired("competition")
	}
	if r.Year <= 2000 {
		return ErrFieldInvalid("year", "must be greater than 2000")
	}
	if r.Rounds <= 0 {
		return ErrFieldInvalid("rounds", "must be greater than 0")
	}
	return nil
}

func (r *CreateTournamentRequest) DerivedName() string {
	return fmt.Sprintf("%s (%d)", r.Competition, r.Year)
}

type TournamentResponse struct {
	ID                   string     `json:"id"`
	Name                 string     `json:"name"`
	Rounds               int        `json:"rounds"`
	Winner               string     `json:"winner,omitempty"`
	FinalFourTopLeft     string     `json:"finalFourTopLeft,omitempty"`
	FinalFourBottomLeft  string     `json:"finalFourBottomLeft,omitempty"`
	FinalFourTopRight    string     `json:"finalFourTopRight,omitempty"`
	FinalFourBottomRight string     `json:"finalFourBottomRight,omitempty"`
	StartingAt           *time.Time `json:"startingAt,omitempty"`
	CreatedAt            time.Time  `json:"createdAt"`
	UpdatedAt            time.Time  `json:"updatedAt"`
}

func NewTournamentResponse(t *models.Tournament, winner string) *TournamentResponse {
	return &TournamentResponse{
		ID:                   t.ID,
		Name:                 t.Name,
		Rounds:               t.Rounds,
		Winner:               winner,
		FinalFourTopLeft:     t.FinalFourTopLeft,
		FinalFourBottomLeft:  t.FinalFourBottomLeft,
		FinalFourTopRight:    t.FinalFourTopRight,
		FinalFourBottomRight: t.FinalFourBottomRight,
		StartingAt:           t.StartingAt,
		CreatedAt:            t.CreatedAt,
		UpdatedAt:            t.UpdatedAt,
	}
}

type NullableTime struct {
	Present bool
	Value   *time.Time
}

func (nt *NullableTime) UnmarshalJSON(b []byte) error {
	nt.Present = true

	if string(b) == "null" {
		nt.Value = nil
		return nil
	}

	var t time.Time
	if err := json.Unmarshal(b, &t); err != nil {
		return err
	}
	nt.Value = &t
	return nil
}

type UpdateTournamentRequest struct {
	StartingAt           NullableTime `json:"startingAt"`
	FinalFourTopLeft     *string      `json:"finalFourTopLeft,omitempty"`
	FinalFourBottomLeft  *string      `json:"finalFourBottomLeft,omitempty"`
	FinalFourTopRight    *string      `json:"finalFourTopRight,omitempty"`
	FinalFourBottomRight *string      `json:"finalFourBottomRight,omitempty"`
}

func (r *UpdateTournamentRequest) Validate() error {
	if r.StartingAt.Present || r.FinalFourTopLeft != nil || r.FinalFourBottomLeft != nil || r.FinalFourTopRight != nil || r.FinalFourBottomRight != nil {
		return nil
	}
	return ErrFieldRequired("at least one field")
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
	IsEliminated *bool `json:"isEliminated,omitempty"`
}

type TournamentTeamResponse struct {
	ID           string          `json:"id"`
	TournamentID string          `json:"tournamentId"`
	SchoolID     string          `json:"schoolId"`
	Seed         int             `json:"seed"`
	Region       string          `json:"region"`
	Byes         int             `json:"byes"`
	Wins         int             `json:"wins"`
	IsEliminated bool            `json:"isEliminated"`
	CreatedAt    time.Time       `json:"createdAt"`
	UpdatedAt    time.Time       `json:"updatedAt"`
	School       *SchoolResponse `json:"school,omitempty"`
	KenPom       *KenPomResponse `json:"kenPom,omitempty"`
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
		IsEliminated: t.IsEliminated,
		CreatedAt:    t.CreatedAt,
		UpdatedAt:    t.UpdatedAt,
	}
	if school != nil {
		resp.School = NewSchoolResponse(school)
	}
	if t.KenPom != nil {
		resp.KenPom = NewKenPomResponse(t.KenPom)
	}
	return resp
}

type KenPomResponse struct {
	NetRtg float64 `json:"netRtg"`
	ORtg   float64 `json:"oRtg"`
	DRtg   float64 `json:"dRtg"`
	AdjT   float64 `json:"adjT"`
}

func NewKenPomResponse(k *models.KenPomStats) *KenPomResponse {
	resp := &KenPomResponse{}
	if k.NetRtg != nil {
		resp.NetRtg = *k.NetRtg
	}
	if k.ORtg != nil {
		resp.ORtg = *k.ORtg
	}
	if k.DRtg != nil {
		resp.DRtg = *k.DRtg
	}
	if k.AdjT != nil {
		resp.AdjT = *k.AdjT
	}
	return resp
}

type UpdateKenPomStatsRequest struct {
	Stats []KenPomStatEntry `json:"stats"`
}

type KenPomStatEntry struct {
	TeamID string  `json:"teamId"`
	NetRtg float64 `json:"netRtg"`
	ORtg   float64 `json:"oRtg"`
	DRtg   float64 `json:"dRtg"`
	AdjT   float64 `json:"adjT"`
}

func (r *UpdateKenPomStatsRequest) Validate() error {
	if len(r.Stats) == 0 {
		return ErrFieldRequired("stats")
	}
	for i, s := range r.Stats {
		if strings.TrimSpace(s.TeamID) == "" {
			return ErrFieldInvalid("stats", fmt.Sprintf("stats[%d]: teamId is required", i))
		}
	}
	return nil
}

// ReplaceTeamsRequest is the request body for PUT /api/tournaments/{id}/teams
type ReplaceTeamsRequest struct {
	Teams []ReplaceTeamEntry `json:"teams"`
}

// ReplaceTeamEntry represents a single team in the bulk replace request.
type ReplaceTeamEntry struct {
	SchoolID string `json:"schoolId"`
	Seed     int    `json:"seed"`
	Region   string `json:"region"`
}

// Validate returns all validation errors found in the request.
func (r *ReplaceTeamsRequest) Validate() []string {
	var errs []string

	if len(r.Teams) == 0 {
		errs = append(errs, "teams array is required")
		return errs
	}

	schoolIDs := make(map[string]int)
	for i, t := range r.Teams {
		if strings.TrimSpace(t.SchoolID) == "" {
			errs = append(errs, fmt.Sprintf("teams[%d]: schoolId is required", i))
		} else {
			schoolIDs[t.SchoolID]++
		}
		if t.Seed < 1 || t.Seed > 16 {
			errs = append(errs, fmt.Sprintf("teams[%d]: seed must be between 1 and 16", i))
		}
		if strings.TrimSpace(t.Region) == "" {
			errs = append(errs, fmt.Sprintf("teams[%d]: region is required", i))
		}
	}

	for schoolID, count := range schoolIDs {
		if count > 1 {
			errs = append(errs, fmt.Sprintf("school %s appears %d times", schoolID, count))
		}
	}

	return errs
}

// BracketValidationErrorResponse is the error response for bracket validation failures.
type BracketValidationErrorResponse struct {
	Code   string   `json:"code"`
	Errors []string `json:"errors"`
}

// CompetitionResponse is the response for a competition.
type CompetitionResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// SeasonResponse is the response for a season.
type SeasonResponse struct {
	ID   string `json:"id"`
	Year int    `json:"year"`
}
