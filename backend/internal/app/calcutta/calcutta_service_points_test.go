package calcutta

import (
	"testing"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

func rounds() []*models.CalcuttaRound {
	return []*models.CalcuttaRound{
		{ID: "round1", CalcuttaID: "calcutta1", Round: 1, Points: 0},
		{ID: "round2", CalcuttaID: "calcutta1", Round: 2, Points: 50},
		{ID: "round3", CalcuttaID: "calcutta1", Round: 3, Points: 100},
		{ID: "round4", CalcuttaID: "calcutta1", Round: 4, Points: 150},
		{ID: "round5", CalcuttaID: "calcutta1", Round: 5, Points: 200},
		{ID: "round6", CalcuttaID: "calcutta1", Round: 6, Points: 250},
		{ID: "round7", CalcuttaID: "calcutta1", Round: 7, Points: 300},
	}
}

func TestThatTeamWithNoProgressScoresZeroPoints(t *testing.T) {
	service := newTestCalcuttaService()
	team := &models.TournamentTeam{ID: "team1", Byes: 0, Wins: 0}

	points := service.CalculatePoints(team, rounds())
	if points != 0.0 {
		t.Fatalf("expected 0.0, got %v", points)
	}
}

func TestThatTeamWithOneWinScoresZeroPoints(t *testing.T) {
	service := newTestCalcuttaService()
	team := &models.TournamentTeam{ID: "team2", Byes: 0, Wins: 1}

	points := service.CalculatePoints(team, rounds())
	if points != 0.0 {
		t.Fatalf("expected 0.0, got %v", points)
	}
}

func TestThatTeamWithOneByeScoresZeroPoints(t *testing.T) {
	service := newTestCalcuttaService()
	team := &models.TournamentTeam{ID: "team3", Byes: 1, Wins: 0}

	points := service.CalculatePoints(team, rounds())
	if points != 0.0 {
		t.Fatalf("expected 0.0, got %v", points)
	}
}

func TestThatTeamWithTwoWinsScores50Points(t *testing.T) {
	service := newTestCalcuttaService()
	team := &models.TournamentTeam{ID: "team4", Byes: 0, Wins: 2}

	points := service.CalculatePoints(team, rounds())
	if points != 50.0 {
		t.Fatalf("expected 50.0, got %v", points)
	}
}

func TestThatTeamWithThreeWinsScores150Points(t *testing.T) {
	service := newTestCalcuttaService()
	team := &models.TournamentTeam{ID: "team5", Byes: 0, Wins: 3}

	points := service.CalculatePoints(team, rounds())
	if points != 150.0 {
		t.Fatalf("expected 150.0, got %v", points)
	}
}

func TestThatTeamWithFourWinsScores300Points(t *testing.T) {
	service := newTestCalcuttaService()
	team := &models.TournamentTeam{ID: "team6", Byes: 0, Wins: 4}

	points := service.CalculatePoints(team, rounds())
	if points != 300.0 {
		t.Fatalf("expected 300.0, got %v", points)
	}
}

func TestThatTeamWithFiveWinsScores500Points(t *testing.T) {
	service := newTestCalcuttaService()
	team := &models.TournamentTeam{ID: "team7", Byes: 0, Wins: 5}

	points := service.CalculatePoints(team, rounds())
	if points != 500.0 {
		t.Fatalf("expected 500.0, got %v", points)
	}
}

func TestThatTeamWithSevenWinsScores1050Points(t *testing.T) {
	service := newTestCalcuttaService()
	team := &models.TournamentTeam{ID: "team8", Byes: 0, Wins: 7}

	points := service.CalculatePoints(team, rounds())
	if points != 1050.0 {
		t.Fatalf("expected 1050.0, got %v", points)
	}
}

func TestThatTeamWithOneByeAndTwoWinsScores150Points(t *testing.T) {
	service := newTestCalcuttaService()
	team := &models.TournamentTeam{ID: "team9", Byes: 1, Wins: 2}

	points := service.CalculatePoints(team, rounds())
	if points != 150.0 {
		t.Fatalf("expected 150.0, got %v", points)
	}
}

func TestThatTeamWithOneByeAndSixWinsScores1050Points(t *testing.T) {
	service := newTestCalcuttaService()
	team := &models.TournamentTeam{ID: "team10", Byes: 1, Wins: 6}

	points := service.CalculatePoints(team, rounds())
	if points != 1050.0 {
		t.Fatalf("expected 1050.0, got %v", points)
	}
}
