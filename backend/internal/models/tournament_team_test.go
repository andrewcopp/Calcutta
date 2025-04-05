package models

import (
	"testing"
	"time"
)

func TestTournamentTeam_Validate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		team    *TournamentTeam
		wantErr bool
	}{
		{
			name: "valid seed 1",
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
			wantErr: false,
		},
		{
			name: "valid seed 16",
			team: &TournamentTeam{
				ID:           "2",
				SchoolID:     "school2",
				TournamentID: "tournament1",
				Seed:         16,
				Byes:         0,
				Wins:         0,
				Created:      now,
				Updated:      now,
			},
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
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.team.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("TournamentTeam.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
