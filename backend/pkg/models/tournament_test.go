package models

import (
	"testing"
	"time"
)

func TestGetTournamentState(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		games    []TournamentGame
		expected TournamentState
	}{
		{
			name:     "Empty games list",
			games:    []TournamentGame{},
			expected: TournamentStateFuture,
		},
		{
			name: "All future games",
			games: []TournamentGame{
				{
					TipoffTime: now.Add(24 * time.Hour),
					IsFinal:    false,
				},
				{
					TipoffTime: now.Add(48 * time.Hour),
					IsFinal:    false,
				},
			},
			expected: TournamentStateFuture,
		},
		{
			name: "All completed games",
			games: []TournamentGame{
				{
					TipoffTime: now.Add(-48 * time.Hour),
					IsFinal:    true,
				},
				{
					TipoffTime: now.Add(-24 * time.Hour),
					IsFinal:    true,
				},
			},
			expected: TournamentStateCompleted,
		},
		{
			name: "Mix of future and completed games",
			games: []TournamentGame{
				{
					TipoffTime: now.Add(-48 * time.Hour),
					IsFinal:    true,
				},
				{
					TipoffTime: now.Add(24 * time.Hour),
					IsFinal:    false,
				},
			},
			expected: TournamentStateInProgress,
		},
		{
			name: "Mix of completed and in-progress games",
			games: []TournamentGame{
				{
					TipoffTime: now.Add(-48 * time.Hour),
					IsFinal:    true,
				},
				{
					TipoffTime: now.Add(-1 * time.Hour),
					IsFinal:    false,
				},
			},
			expected: TournamentStateInProgress,
		},
		{
			name: "Mix of future and in-progress games",
			games: []TournamentGame{
				{
					TipoffTime: now.Add(24 * time.Hour),
					IsFinal:    false,
				},
				{
					TipoffTime: now.Add(-1 * time.Hour),
					IsFinal:    false,
				},
			},
			expected: TournamentStateInProgress,
		},
		{
			name: "Mix of all states",
			games: []TournamentGame{
				{
					TipoffTime: now.Add(-48 * time.Hour),
					IsFinal:    true,
				},
				{
					TipoffTime: now.Add(-1 * time.Hour),
					IsFinal:    false,
				},
				{
					TipoffTime: now.Add(24 * time.Hour),
					IsFinal:    false,
				},
			},
			expected: TournamentStateInProgress,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetTournamentState(tt.games)
			if result != tt.expected {
				t.Errorf("GetTournamentState() = %v, want %v", result, tt.expected)
			}
		})
	}
}
