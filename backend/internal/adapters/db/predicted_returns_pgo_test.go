package db

import (
	"fmt"
	"math"
	"testing"

	appbracket "github.com/andrewcopp/Calcutta/backend/internal/app/bracket"
	"github.com/andrewcopp/Calcutta/backend/pkg/models"
)

func TestThatPGODPCreatesReachForAllTeams(t *testing.T) {
	givenTournamentID := "t"
	givenTeams := createFullTournamentTeams(givenTournamentID)
	givenFinalFour := &models.FinalFourConfig{TopLeftRegion: "East", BottomLeftRegion: "West", TopRightRegion: "South", BottomRightRegion: "Midwest"}
	givenScoringRules := []scoringRule{{WinIndex: 7, PointsAwarded: 100}}

	whenBuilder := appbracket.NewBracketBuilder()
	whenBracket, err := whenBuilder.BuildBracket(givenTournamentID, givenTeams, givenFinalFour)
	if err != nil {
		t.Fatalf("failed to build bracket: %v", err)
	}
	_, whenReachByTeam, err := computeExpectedValueFromPGO(whenBracket, map[matchupKey]float64{}, givenScoringRules, map[string]float64{}, 10.0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	thenGot := len(whenReachByTeam)
	thenWant := 68
	if thenGot != thenWant {
		t.Fatalf("expected reachByTeam size %d, got %d", thenWant, thenGot)
	}
}

func TestThatPGODPWinChampProbabilitiesSumToOneWithFallback(t *testing.T) {
	givenTournamentID := "t"
	givenTeams := createFullTournamentTeams(givenTournamentID)
	givenFinalFour := &models.FinalFourConfig{TopLeftRegion: "East", BottomLeftRegion: "West", TopRightRegion: "South", BottomRightRegion: "Midwest"}
	givenScoringRules := []scoringRule{{WinIndex: 7, PointsAwarded: 100}}

	whenBuilder := appbracket.NewBracketBuilder()
	whenBracket, err := whenBuilder.BuildBracket(givenTournamentID, givenTeams, givenFinalFour)
	if err != nil {
		t.Fatalf("failed to build bracket: %v", err)
	}
	_, whenReachByTeam, err := computeExpectedValueFromPGO(whenBracket, map[matchupKey]float64{}, givenScoringRules, map[string]float64{}, 10.0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	whenSum := 0.0
	for _, rr := range whenReachByTeam {
		whenSum += rr.WinChamp
	}

	thenGot := whenSum
	thenWant := 1.0
	if math.Abs(thenGot-thenWant) > 1e-6 {
		t.Fatalf("expected sum(win_champ) to be ~1.0, got %.10f", thenGot)
	}
}

func TestThatPGODPProducesNonZeroExpectedValueForSomeTeam(t *testing.T) {
	givenTournamentID := "t"
	givenTeams := createFullTournamentTeams(givenTournamentID)
	givenFinalFour := &models.FinalFourConfig{TopLeftRegion: "East", BottomLeftRegion: "West", TopRightRegion: "South", BottomRightRegion: "Midwest"}
	givenScoringRules := []scoringRule{{WinIndex: 7, PointsAwarded: 100}}

	whenBuilder := appbracket.NewBracketBuilder()
	whenBracket, err := whenBuilder.BuildBracket(givenTournamentID, givenTeams, givenFinalFour)
	if err != nil {
		t.Fatalf("failed to build bracket: %v", err)
	}
	whenEVByTeam, _, err := computeExpectedValueFromPGO(whenBracket, map[matchupKey]float64{}, givenScoringRules, map[string]float64{}, 10.0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	whenMax := 0.0
	for _, ev := range whenEVByTeam {
		if ev > whenMax {
			whenMax = ev
		}
	}

	thenGot := whenMax
	if thenGot <= 0 {
		t.Fatalf("expected some team to have positive expected value")
	}
}

func TestThatPGODPFavorsHigherKenPomNetInChampionshipProbability(t *testing.T) {
	givenTournamentID := "t"
	givenTeams := createFullTournamentTeams(givenTournamentID)
	givenFinalFour := &models.FinalFourConfig{TopLeftRegion: "East", BottomLeftRegion: "West", TopRightRegion: "South", BottomRightRegion: "Midwest"}
	givenScoringRules := []scoringRule{{WinIndex: 7, PointsAwarded: 100}}

	whenNetByTeamID := make(map[string]float64)
	whenBestID := ""
	whenBestNet := math.Inf(-1)
	whenWorstID := ""
	whenWorstNet := math.Inf(1)
	for _, tm := range givenTeams {
		net := float64(20 - tm.Seed)
		whenNetByTeamID[tm.ID] = net
		if net > whenBestNet {
			whenBestNet = net
			whenBestID = tm.ID
		}
		if net < whenWorstNet {
			whenWorstNet = net
			whenWorstID = tm.ID
		}
	}

	whenBuilder := appbracket.NewBracketBuilder()
	whenBracket, err := whenBuilder.BuildBracket(givenTournamentID, givenTeams, givenFinalFour)
	if err != nil {
		t.Fatalf("failed to build bracket: %v", err)
	}
	_, whenReachByTeam, err := computeExpectedValueFromPGO(whenBracket, map[matchupKey]float64{}, givenScoringRules, whenNetByTeamID, 10.0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	thenBest := whenReachByTeam[whenBestID].WinChamp
	thenWorst := whenReachByTeam[whenWorstID].WinChamp
	if thenBest <= thenWorst {
		t.Fatalf("expected best-net team to have higher win championship prob than worst-net team; best=%.6f worst=%.6f", thenBest, thenWorst)
	}
}

func createRegionTeams(tournamentID, region string, duplicateSeeds []int) []*models.TournamentTeam {
	teams := make([]*models.TournamentTeam, 0, 16+len(duplicateSeeds))
	for seed := 1; seed <= 16; seed++ {
		teams = append(teams, createTeam(tournamentID, region, seed, ""))
	}
	for _, seed := range duplicateSeeds {
		teams = append(teams, createTeam(tournamentID, region, seed, "b"))
	}
	return teams
}

func createFullTournamentTeams(tournamentID string) []*models.TournamentTeam {
	out := make([]*models.TournamentTeam, 0, 68)
	out = append(out, createRegionTeams(tournamentID, "East", []int{11, 16})...)
	out = append(out, createRegionTeams(tournamentID, "West", nil)...)
	out = append(out, createRegionTeams(tournamentID, "South", []int{11})...)
	out = append(out, createRegionTeams(tournamentID, "Midwest", []int{11})...)
	return out
}

func createTeam(tournamentID, region string, seed int, suffix string) *models.TournamentTeam {
	id := fmt.Sprintf("%s-%02d%s", region, seed, suffix)
	schoolID := fmt.Sprintf("school-%s", id)
	name := fmt.Sprintf("%s-%02d%s", region, seed, suffix)
	return &models.TournamentTeam{
		ID:           id,
		SchoolID:     schoolID,
		TournamentID: tournamentID,
		Seed:         seed,
		Region:       region,
		School:       &models.School{ID: schoolID, Name: name},
	}
}
