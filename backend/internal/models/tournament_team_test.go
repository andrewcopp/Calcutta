package models

import (
	"testing"
	"time"
)

func TestTournamentTeam_Validate(t *testing.T) {
	now := time.Now()
	defaultConfig := DefaultTournamentTeamConfig()

	tests := []struct {
		name    string
		team    *TournamentTeam
		config  *TournamentTeamConfig
		wantErr bool
	}{
		{
			name: "valid team with default config",
			team: &TournamentTeam{
				ID:           "1",
				SchoolID:     "school1",
				TournamentID: "tournament1",
				Seed:         1,
				Byes:         0,
				Wins:         0,
				Created:      now,
				Updated:      now,
			},
			config:  defaultConfig,
			wantErr: false,
		},
		{
			name: "valid team with max values",
			team: &TournamentTeam{
				ID:           "2",
				SchoolID:     "school2",
				TournamentID: "tournament1",
				Seed:         16,
				Byes:         1,
				Wins:         7,
				Created:      now,
				Updated:      now,
			},
			config:  defaultConfig,
			wantErr: false,
		},
		{
			name: "invalid seed 0",
			team: &TournamentTeam{
				ID:           "3",
				SchoolID:     "school3",
				TournamentID: "tournament1",
				Seed:         0,
				Byes:         0,
				Wins:         0,
				Created:      now,
				Updated:      now,
			},
			config:  defaultConfig,
			wantErr: true,
		},
		{
			name: "invalid seed 17",
			team: &TournamentTeam{
				ID:           "4",
				SchoolID:     "school4",
				TournamentID: "tournament1",
				Seed:         17,
				Byes:         0,
				Wins:         0,
				Created:      now,
				Updated:      now,
			},
			config:  defaultConfig,
			wantErr: true,
		},
		{
			name: "invalid byes -1",
			team: &TournamentTeam{
				ID:           "5",
				SchoolID:     "school5",
				TournamentID: "tournament1",
				Seed:         1,
				Byes:         -1,
				Wins:         0,
				Created:      now,
				Updated:      now,
			},
			config:  defaultConfig,
			wantErr: true,
		},
		{
			name: "invalid byes 2",
			team: &TournamentTeam{
				ID:           "6",
				SchoolID:     "school6",
				TournamentID: "tournament1",
				Seed:         1,
				Byes:         2,
				Wins:         0,
				Created:      now,
				Updated:      now,
			},
			config:  defaultConfig,
			wantErr: true,
		},
		{
			name: "invalid wins -1",
			team: &TournamentTeam{
				ID:           "7",
				SchoolID:     "school7",
				TournamentID: "tournament1",
				Seed:         1,
				Byes:         0,
				Wins:         -1,
				Created:      now,
				Updated:      now,
			},
			config:  defaultConfig,
			wantErr: true,
		},
		{
			name: "invalid wins 8",
			team: &TournamentTeam{
				ID:           "8",
				SchoolID:     "school8",
				TournamentID: "tournament1",
				Seed:         1,
				Byes:         0,
				Wins:         8,
				Created:      now,
				Updated:      now,
			},
			config:  defaultConfig,
			wantErr: true,
		},
		{
			name: "custom config - expanded tournament",
			team: &TournamentTeam{
				ID:           "9",
				SchoolID:     "school9",
				TournamentID: "tournament1",
				Seed:         20,
				Byes:         2,
				Wins:         8,
				Created:      now,
				Updated:      now,
			},
			config: &TournamentTeamConfig{
				MinSeed: 1,
				MaxSeed: 20,
				MinByes: 0,
				MaxByes: 2,
				MinWins: 0,
				MaxWins: 8,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.team.Validate(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("TournamentTeam.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestTournamentTeam_ValidateDefault(t *testing.T) {
	now := time.Now()
	team := &TournamentTeam{
		ID:           "1",
		SchoolID:     "school1",
		TournamentID: "tournament1",
		Seed:         1,
		Byes:         0,
		Wins:         0,
		Created:      now,
		Updated:      now,
	}

	err := team.ValidateDefault()
	if err != nil {
		t.Errorf("TournamentTeam.ValidateDefault() error = %v", err)
	}
}
