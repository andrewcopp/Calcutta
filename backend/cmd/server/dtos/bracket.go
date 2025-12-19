package dtos

import (
	"github.com/andrewcopp/Calcutta/backend/pkg/models"
)

// BracketResponse represents the complete bracket structure for the frontend
type BracketResponse struct {
	TournamentID string                   `json:"tournamentId"`
	Regions      []string                 `json:"regions"`
	Games        []*BracketGameResponse   `json:"games"`
	FinalFour    *FinalFourConfigResponse `json:"finalFour"`
}

// BracketGameResponse represents a single game in the bracket
type BracketGameResponse struct {
	GameID       string               `json:"gameId"`
	Round        string               `json:"round"`
	Region       string               `json:"region"`
	Team1        *BracketTeamResponse `json:"team1,omitempty"`
	Team2        *BracketTeamResponse `json:"team2,omitempty"`
	Winner       *BracketTeamResponse `json:"winner,omitempty"`
	NextGameID   string               `json:"nextGameId,omitempty"`
	NextGameSlot int                  `json:"nextGameSlot,omitempty"`
	SortOrder    int                  `json:"sortOrder"`
	CanSelect    bool                 `json:"canSelect"` // Whether a winner can be selected
}

// BracketTeamResponse represents a team in the bracket
type BracketTeamResponse struct {
	TeamID   string `json:"teamId"`
	SchoolID string `json:"schoolId"`
	Name     string `json:"name"`
	Seed     int    `json:"seed"`
	Region   string `json:"region"`
}

// FinalFourConfigResponse represents the Final Four matchup configuration
type FinalFourConfigResponse struct {
	TopLeftRegion     string `json:"topLeftRegion"`
	BottomLeftRegion  string `json:"bottomLeftRegion"`
	TopRightRegion    string `json:"topRightRegion"`
	BottomRightRegion string `json:"bottomRightRegion"`
}

// SelectWinnerRequest represents a request to select a winner for a game
type SelectWinnerRequest struct {
	WinnerTeamID string `json:"winnerTeamId"`
}

func (r *SelectWinnerRequest) Validate() error {
	if r.WinnerTeamID == "" {
		return ErrFieldRequired("winnerTeamId")
	}
	return nil
}

// NewBracketResponse converts a bracket structure to a response DTO
func NewBracketResponse(bracket *models.BracketStructure) *BracketResponse {
	games := make([]*BracketGameResponse, 0, len(bracket.Games))
	for _, game := range bracket.Games {
		games = append(games, NewBracketGameResponse(game))
	}

	return &BracketResponse{
		TournamentID: bracket.TournamentID,
		Regions:      bracket.Regions,
		Games:        games,
		FinalFour:    NewFinalFourConfigResponse(bracket.FinalFour),
	}
}

// NewBracketGameResponse converts a bracket game to a response DTO
func NewBracketGameResponse(game *models.BracketGame) *BracketGameResponse {
	resp := &BracketGameResponse{
		GameID:       game.GameID,
		Round:        string(game.Round),
		Region:       game.Region,
		NextGameID:   game.NextGameID,
		NextGameSlot: game.NextGameSlot,
		SortOrder:    game.SortOrder,
		CanSelect:    game.Team1 != nil && game.Team2 != nil && game.Winner == nil,
	}

	if game.Team1 != nil {
		resp.Team1 = NewBracketTeamResponse(game.Team1)
	}
	if game.Team2 != nil {
		resp.Team2 = NewBracketTeamResponse(game.Team2)
	}
	if game.Winner != nil {
		resp.Winner = NewBracketTeamResponse(game.Winner)
	}

	return resp
}

// NewBracketTeamResponse converts a bracket team to a response DTO
func NewBracketTeamResponse(team *models.BracketTeam) *BracketTeamResponse {
	return &BracketTeamResponse{
		TeamID:   team.TeamID,
		SchoolID: team.SchoolID,
		Name:     team.Name,
		Seed:     team.Seed,
		Region:   team.Region,
	}
}

// NewFinalFourConfigResponse converts a Final Four config to a response DTO
func NewFinalFourConfigResponse(config *models.FinalFourConfig) *FinalFourConfigResponse {
	if config == nil {
		return nil
	}
	return &FinalFourConfigResponse{
		TopLeftRegion:     config.TopLeftRegion,
		BottomLeftRegion:  config.BottomLeftRegion,
		TopRightRegion:    config.TopRightRegion,
		BottomRightRegion: config.BottomRightRegion,
	}
}
