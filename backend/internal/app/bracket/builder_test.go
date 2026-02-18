package bracket

import (
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

type bracketLink struct {
	NextGameID   string
	NextGameSlot int
}

func TestThatBuildBracketStructureReturnsErrorWhenNotSixtyEightTeams(t *testing.T) {
	ff := &models.FinalFourConfig{TopLeftRegion: "East", BottomLeftRegion: "West", TopRightRegion: "South", BottomRightRegion: "Midwest"}

	_, err := BuildBracketStructure("t", nil, ff)

	if err == nil {
		t.Errorf("expected error")
	}
}

func TestThatBuildBracketStructureSetsTournamentID(t *testing.T) {
	teams := createFullTournamentTeams("t")
	ff := &models.FinalFourConfig{TopLeftRegion: "East", BottomLeftRegion: "West", TopRightRegion: "South", BottomRightRegion: "Midwest"}

	bracket, err := BuildBracketStructure("t", teams, ff)
	if err != nil {
		t.Fatalf("failed to build bracket: %v", err)
	}

	got := ""
	if bracket != nil {
		got = bracket.TournamentID
	}
	if got != "t" {
		t.Errorf("expected tournament id to be t")
	}
}

func TestThatBuildBracketStructurePreservesFinalFourConfigPointer(t *testing.T) {
	teams := createFullTournamentTeams("t")
	ff := &models.FinalFourConfig{TopLeftRegion: "East", BottomLeftRegion: "West", TopRightRegion: "South", BottomRightRegion: "Midwest"}

	bracket, err := BuildBracketStructure("t", teams, ff)
	if err != nil {
		t.Fatalf("failed to build bracket: %v", err)
	}

	if bracket.FinalFour != ff {
		t.Errorf("expected final four config pointer to be preserved")
	}
}

func TestThatRegionWithNoDuplicateSeedsCreatesZeroFirstFourGames(t *testing.T) {
	bracket := &models.BracketStructure{TournamentID: "t", Regions: []string{"West"}, Games: make(map[string]*models.BracketGame)}
	teams := createRegionTeams("t", "West", nil)

	_, err := buildRegionalBracket(bracket, "West", teams)
	if err != nil {
		t.Fatalf("failed to build regional bracket: %v", err)
	}

	got := countGamesByRound(bracket, models.RoundFirstFour)
	want := 0
	if got != want {
		t.Errorf("expected %d first four games, got %d", want, got)
	}
}

func TestThatRegionWithOneDuplicateSeedCreatesOneFirstFourGame(t *testing.T) {
	bracket := &models.BracketStructure{TournamentID: "t", Regions: []string{"East"}, Games: make(map[string]*models.BracketGame)}
	teams := createRegionTeams("t", "East", []int{11})

	_, err := buildRegionalBracket(bracket, "East", teams)
	if err != nil {
		t.Fatalf("failed to build regional bracket: %v", err)
	}

	got := countGamesByRound(bracket, models.RoundFirstFour)
	want := 1
	if got != want {
		t.Errorf("expected %d first four games, got %d", want, got)
	}
}

func TestThatFirstFourGameForElevenSeedHasDeterministicID(t *testing.T) {
	bracket := &models.BracketStructure{TournamentID: "t", Regions: []string{"East"}, Games: make(map[string]*models.BracketGame)}
	teams := createRegionTeams("t", "East", []int{11})

	_, err := buildRegionalBracket(bracket, "East", teams)
	if err != nil {
		t.Fatalf("failed to build regional bracket: %v", err)
	}

	exists := bracket.Games["East-first_four-11"] != nil
	if !exists {
		t.Errorf("expected first four game with deterministic id East-first_four-11")
	}
}

func TestThatFirstFourGameLinksToCorrectRoundOfSixtyFourGame(t *testing.T) {
	bracket := &models.BracketStructure{TournamentID: "t", Regions: []string{"East"}, Games: make(map[string]*models.BracketGame)}
	teams := createRegionTeams("t", "East", []int{11})

	_, err := buildRegionalBracket(bracket, "East", teams)
	if err != nil {
		t.Fatalf("failed to build regional bracket: %v", err)
	}
	game := bracket.Games["East-first_four-11"]

	got := bracketLink{}
	if game != nil {
		got = bracketLink{NextGameID: game.NextGameID, NextGameSlot: game.NextGameSlot}
	}
	want := bracketLink{NextGameID: "East-round_of_64-6", NextGameSlot: 2}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("expected first four link %+v, got %+v", want, got)
	}
}

func TestThatRoundOfSixtyFourGameForOneSeedHasDeterministicID(t *testing.T) {
	bracket := &models.BracketStructure{TournamentID: "t", Regions: []string{"West"}, Games: make(map[string]*models.BracketGame)}
	teams := createRegionTeams("t", "West", nil)

	_, err := buildRegionalBracket(bracket, "West", teams)
	if err != nil {
		t.Fatalf("failed to build regional bracket: %v", err)
	}

	exists := bracket.Games["West-round_of_64-1"] != nil
	if !exists {
		t.Errorf("expected round of 64 game with deterministic id West-round_of_64-1")
	}
}

func TestThatRoundOfSixtyFourMatchesOneSeedAgainstSixteenSeed(t *testing.T) {
	bracket := &models.BracketStructure{TournamentID: "t", Regions: []string{"West"}, Games: make(map[string]*models.BracketGame)}
	teams := createRegionTeams("t", "West", nil)

	_, err := buildRegionalBracket(bracket, "West", teams)
	if err != nil {
		t.Fatalf("failed to build regional bracket: %v", err)
	}
	game := bracket.Games["West-round_of_64-1"]

	got := matchupSeeds{}
	if game != nil && game.Team1 != nil && game.Team2 != nil {
		got = matchupSeeds{Team1Seed: game.Team1.Seed, Team2Seed: game.Team2.Seed}
	}
	want := matchupSeeds{Team1Seed: 1, Team2Seed: 16}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("expected matchup %+v, got %+v", want, got)
	}
}

type matchupSeeds struct {
	Team1Seed int
	Team2Seed int
}

func TestThatRoundOfSixtyFourSeedsSumToSeventeen(t *testing.T) {
	bracket := &models.BracketStructure{TournamentID: "t", Regions: []string{"West"}, Games: make(map[string]*models.BracketGame)}
	teams := createRegionTeams("t", "West", nil)

	_, err := buildRegionalBracket(bracket, "West", teams)
	if err != nil {
		t.Fatalf("failed to build regional bracket: %v", err)
	}

	satisfied := true
	for _, game := range bracket.Games {
		if game == nil || game.Round != models.RoundOf64 || game.Region != "West" {
			continue
		}
		if game.Team1 == nil || game.Team2 == nil {
			satisfied = false
			break
		}
		if game.Team1.Seed+game.Team2.Seed != 17 {
			satisfied = false
			break
		}
	}

	if !satisfied {
		t.Errorf("expected all round of 64 games in West to have seeds summing to 17")
	}
}

func TestThatRoundOfThirtyTwoGameHasDeterministicIDBasedOnLowestSeed(t *testing.T) {
	bracket := &models.BracketStructure{TournamentID: "t", Regions: []string{"West"}, Games: make(map[string]*models.BracketGame)}
	teams := createRegionTeams("t", "West", nil)

	_, err := buildRegionalBracket(bracket, "West", teams)
	if err != nil {
		t.Fatalf("failed to build regional bracket: %v", err)
	}

	exists := bracket.Games["West-round_of_32-1"] != nil
	if !exists {
		t.Errorf("expected round of 32 game with deterministic id West-round_of_32-1")
	}
}

func TestThatSweetSixteenGameHasDeterministicIDBasedOnLowestSeed(t *testing.T) {
	bracket := &models.BracketStructure{TournamentID: "t", Regions: []string{"West"}, Games: make(map[string]*models.BracketGame)}
	teams := createRegionTeams("t", "West", nil)

	_, err := buildRegionalBracket(bracket, "West", teams)
	if err != nil {
		t.Fatalf("failed to build regional bracket: %v", err)
	}

	exists := bracket.Games["West-sweet_16-1"] != nil
	if !exists {
		t.Errorf("expected sweet 16 game with deterministic id West-sweet_16-1")
	}
}

func TestThatEliteEightGameHasDeterministicIDForRegionalFinal(t *testing.T) {
	bracket := &models.BracketStructure{TournamentID: "t", Regions: []string{"West"}, Games: make(map[string]*models.BracketGame)}
	teams := createRegionTeams("t", "West", nil)

	_, err := buildRegionalBracket(bracket, "West", teams)
	if err != nil {
		t.Fatalf("failed to build regional bracket: %v", err)
	}

	exists := bracket.Games["West-elite_8-1"] != nil
	if !exists {
		t.Errorf("expected elite 8 game with deterministic id West-elite_8-1")
	}
}

func TestThatBracketTeamInitiallyHasLowestSeedSeenEqualToOwnSeed(t *testing.T) {
	team := &models.TournamentTeam{ID: "x", SchoolID: "s", TournamentID: "t", Seed: 5, Region: "West"}

	got := toBracketTeam(team)

	lowest := -1
	if got != nil {
		lowest = got.LowestSeedSeen
	}
	if lowest != 5 {
		t.Errorf("expected lowest seed seen to equal 5")
	}
}

func TestThatBracketBuilderGeneratesSameGameIDsAcrossMultipleBuilds(t *testing.T) {
	teams := createFullTournamentTeams("t")
	ff := &models.FinalFourConfig{TopLeftRegion: "East", BottomLeftRegion: "West", TopRightRegion: "South", BottomRightRegion: "Midwest"}

	bracket1, err1 := BuildBracketStructure("t", teams, ff)
	bracket2, err2 := BuildBracketStructure("t", teams, ff)

	ids1 := sortedGameIDs(bracket1)
	ids2 := sortedGameIDs(bracket2)

	if err1 != nil {
		t.Fatalf("failed to build bracket1: %v", err1)
	}
	if err2 != nil {
		t.Fatalf("failed to build bracket2: %v", err2)
	}
	if !reflect.DeepEqual(ids1, ids2) {
		t.Errorf("expected deterministic game ids across builds")
	}
}

func TestThatFinalFourSemifinalOneHasDeterministicID(t *testing.T) {
	teams := createFullTournamentTeams("t")
	ff := &models.FinalFourConfig{TopLeftRegion: "East", BottomLeftRegion: "West", TopRightRegion: "South", BottomRightRegion: "Midwest"}

	bracket, err := BuildBracketStructure("t", teams, ff)
	if err != nil {
		t.Fatalf("failed to build bracket: %v", err)
	}

	exists := bracket.Games["final_four-1"] != nil
	if !exists {
		t.Errorf("expected semifinal one to have deterministic id final_four-1")
	}
}

func TestThatFinalFourSemifinalTwoHasDeterministicID(t *testing.T) {
	teams := createFullTournamentTeams("t")
	ff := &models.FinalFourConfig{TopLeftRegion: "East", BottomLeftRegion: "West", TopRightRegion: "South", BottomRightRegion: "Midwest"}

	bracket, err := BuildBracketStructure("t", teams, ff)
	if err != nil {
		t.Fatalf("failed to build bracket: %v", err)
	}

	exists := bracket.Games["final_four-2"] != nil
	if !exists {
		t.Errorf("expected semifinal two to have deterministic id final_four-2")
	}
}

func TestThatChampionshipGameHasDeterministicID(t *testing.T) {
	teams := createFullTournamentTeams("t")
	ff := &models.FinalFourConfig{TopLeftRegion: "East", BottomLeftRegion: "West", TopRightRegion: "South", BottomRightRegion: "Midwest"}

	bracket, err := BuildBracketStructure("t", teams, ff)
	if err != nil {
		t.Fatalf("failed to build bracket: %v", err)
	}

	exists := bracket.Games["championship"] != nil
	if !exists {
		t.Errorf("expected championship to have deterministic id championship")
	}
}

func countGamesByRound(bracket *models.BracketStructure, round models.BracketRound) int {
	if bracket == nil {
		return 0
	}
	count := 0
	for _, game := range bracket.Games {
		if game != nil && game.Round == round {
			count++
		}
	}
	return count
}

func sortedGameIDs(bracket *models.BracketStructure) []string {
	if bracket == nil {
		return nil
	}
	ids := make([]string, 0, len(bracket.Games))
	for id := range bracket.Games {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func createRegionTeams(tournamentID, region string, duplicateSeeds []int) []*models.TournamentTeam {
	teams := make([]*models.TournamentTeam, 0, 16+len(duplicateSeeds))
	for seed := 1; seed <= 16; seed++ {
		teams = append(teams, createTeam(tournamentID, region, seed, ""))
	}
	for _, seed := range duplicateSeeds {
		teams = append(teams, createTeam(tournamentID, region, seed, "b"))
	}

	firstFourSeeds := make(map[int]bool)
	for _, seed := range duplicateSeeds {
		firstFourSeeds[seed] = true
	}
	for _, team := range teams {
		if firstFourSeeds[team.Seed] {
			team.Byes = 0
		} else {
			team.Byes = 1
		}
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
