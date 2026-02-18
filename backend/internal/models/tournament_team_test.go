package models

import (
	"testing"
	"time"
)

func TestThatTournamentTeamValidateAcceptsValidTeamWithDefaultConfig(t *testing.T) {
	// GIVEN a valid tournament team with default config values
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
	config := DefaultTournamentConfig()

	// WHEN validating the team
	err := team.Validate(config)

	// THEN no error is returned
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestThatTournamentTeamValidateAcceptsValidTeamWithMaxValues(t *testing.T) {
	// GIVEN a valid tournament team with max allowed values
	now := time.Now()
	team := &TournamentTeam{
		ID:           "2",
		SchoolID:     "school2",
		TournamentID: "tournament1",
		Seed:         16,
		Byes:         1,
		Wins:         7,
		Created:      now,
		Updated:      now,
	}
	config := DefaultTournamentConfig()

	// WHEN validating the team
	err := team.Validate(config)

	// THEN no error is returned
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestThatTournamentTeamValidateRejectsSeedZero(t *testing.T) {
	// GIVEN a tournament team with seed 0
	now := time.Now()
	team := &TournamentTeam{
		ID:           "3",
		SchoolID:     "school3",
		TournamentID: "tournament1",
		Seed:         0,
		Byes:         0,
		Wins:         0,
		Created:      now,
		Updated:      now,
	}
	config := DefaultTournamentConfig()

	// WHEN validating the team
	err := team.Validate(config)

	// THEN an error is returned
	if err == nil {
		t.Error("expected error for seed 0")
	}
}

func TestThatTournamentTeamValidateRejectsSeedSeventeen(t *testing.T) {
	// GIVEN a tournament team with seed 17
	now := time.Now()
	team := &TournamentTeam{
		ID:           "4",
		SchoolID:     "school4",
		TournamentID: "tournament1",
		Seed:         17,
		Byes:         0,
		Wins:         0,
		Created:      now,
		Updated:      now,
	}
	config := DefaultTournamentConfig()

	// WHEN validating the team
	err := team.Validate(config)

	// THEN an error is returned
	if err == nil {
		t.Error("expected error for seed 17")
	}
}

func TestThatTournamentTeamValidateRejectsNegativeByes(t *testing.T) {
	// GIVEN a tournament team with byes -1
	now := time.Now()
	team := &TournamentTeam{
		ID:           "5",
		SchoolID:     "school5",
		TournamentID: "tournament1",
		Seed:         1,
		Byes:         -1,
		Wins:         0,
		Created:      now,
		Updated:      now,
	}
	config := DefaultTournamentConfig()

	// WHEN validating the team
	err := team.Validate(config)

	// THEN an error is returned
	if err == nil {
		t.Error("expected error for byes -1")
	}
}

func TestThatTournamentTeamValidateRejectsByesGreaterThanOne(t *testing.T) {
	// GIVEN a tournament team with byes 2
	now := time.Now()
	team := &TournamentTeam{
		ID:           "6",
		SchoolID:     "school6",
		TournamentID: "tournament1",
		Seed:         1,
		Byes:         2,
		Wins:         0,
		Created:      now,
		Updated:      now,
	}
	config := DefaultTournamentConfig()

	// WHEN validating the team
	err := team.Validate(config)

	// THEN an error is returned
	if err == nil {
		t.Error("expected error for byes 2")
	}
}

func TestThatTournamentTeamValidateRejectsNegativeWins(t *testing.T) {
	// GIVEN a tournament team with wins -1
	now := time.Now()
	team := &TournamentTeam{
		ID:           "7",
		SchoolID:     "school7",
		TournamentID: "tournament1",
		Seed:         1,
		Byes:         0,
		Wins:         -1,
		Created:      now,
		Updated:      now,
	}
	config := DefaultTournamentConfig()

	// WHEN validating the team
	err := team.Validate(config)

	// THEN an error is returned
	if err == nil {
		t.Error("expected error for wins -1")
	}
}

func TestThatTournamentTeamValidateRejectsWinsGreaterThanSeven(t *testing.T) {
	// GIVEN a tournament team with wins 8
	now := time.Now()
	team := &TournamentTeam{
		ID:           "8",
		SchoolID:     "school8",
		TournamentID: "tournament1",
		Seed:         1,
		Byes:         0,
		Wins:         8,
		Created:      now,
		Updated:      now,
	}
	config := DefaultTournamentConfig()

	// WHEN validating the team
	err := team.Validate(config)

	// THEN an error is returned
	if err == nil {
		t.Error("expected error for wins 8")
	}
}

func TestThatTournamentTeamValidateAcceptsCustomConfig(t *testing.T) {
	// GIVEN a tournament team with values valid for a custom config
	now := time.Now()
	team := &TournamentTeam{
		ID:           "9",
		SchoolID:     "school9",
		TournamentID: "tournament1",
		Seed:         20,
		Byes:         2,
		Wins:         8,
		Created:      now,
		Updated:      now,
	}
	config := &TournamentConfig{
		MinSeed: 1,
		MaxSeed: 20,
		MinByes: 0,
		MaxByes: 2,
		MinWins: 0,
		MaxWins: 8,
	}

	// WHEN validating the team
	err := team.Validate(config)

	// THEN no error is returned
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestThatTournamentTeamValidateDefaultAcceptsValidTeam(t *testing.T) {
	// GIVEN a valid tournament team
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

	// WHEN validating with default config
	err := team.ValidateDefault()

	// THEN no error is returned
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}
