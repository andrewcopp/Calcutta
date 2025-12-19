# Bracket Test Structure

## Overview

The bracket tests follow two key principles to manage complexity:

1. **Helper/Factory Pattern**: Centralized test data creation in `bracket_test_helpers.go`
2. **Function-Based File Organization**: Tests split into focused files per function being tested

## Test Helper (`bracket_test_helpers.go`)

### BracketTestHelper

Central factory for creating test data with minimal boilerplate.

**Key Methods:**

```go
// Tournament creation
CreateTournament68Teams() []*models.TournamentTeam
CreateStandardRegionTeams(region string, count int) []*models.TournamentTeam
CreateRegionWithDuplicateSeeds(region string, duplicateSeeds ...int) []*models.TournamentTeam
CreateTeam(id, region string, seed int) *models.TournamentTeam
CreateFinalFourConfig() *models.FinalFourConfig

// Bracket creation
CreateEmptyBracket() *models.BracketStructure
CreateBracketTeam(teamID string, seed int) *models.BracketTeam
CreateGame(gameID string, round models.BracketRound, team1, team2 *models.BracketTeam) *models.BracketGame
LinkGames(fromGame *models.BracketGame, toGameID string, slot int)

// Game manipulation
SelectWinner(bracket *models.BracketStructure, gameID, teamID string) error
UnselectWinner(bracket *models.BracketStructure, gameID string) error

// Preset scenarios
CreateBracketWithRound64Game() *models.BracketStructure
CreateBracketWithLinkedGames() *models.BracketStructure
CreateBracketWithFirstFourGame() *models.BracketStructure
```

## Test File Organization

### Principle: One File Per Function

Tests are organized by the function they test, not by the class/struct. This keeps files focused and manageable.

### BracketBuilder Tests

**`build_bracket_test.go`**
- Tests for `BracketBuilder.BuildBracket()`
- Example: `TestThatBracketBuilderGeneratesSameGameIDsAcrossMultipleBuilds`

**`build_regional_bracket_first_four_test.go`**
- Tests for `BracketBuilder.buildRegionalBracket()` - First Four game creation
- 7 tests covering duplicate seed handling and deterministic IDs
- Examples:
  - `TestThatRegionWithNoDuplicateSeedsCreatesZeroFirstFourGames`
  - `TestThatFirstFourGameForElevenSeedHasDeterministicID`
  - `TestThatFirstFourGameLinksToCorrectRoundOfSixtyFourGame`

**`build_regional_bracket_round64_test.go`**
- Tests for `BracketBuilder.buildRegionalBracket()` - Round of 64 matchups
- 6 tests covering seed matchups and game creation
- Examples:
  - `TestThatRoundOfSixtyFourMatchesOneSeedAgainstSixteenSeed`
  - `TestThatRoundOfSixtyFourSeedsSumToSeventeen`
  - `TestThatRoundOfSixtyFourCreatesExactlyEightGamesPerRegion`

**`build_regional_bracket_regional_rounds_test.go`**
- Tests for `BracketBuilder.buildRegionalBracket()` - Regional rounds (Round of 32, Sweet 16, Elite 8)
- 3 tests covering deterministic IDs for later rounds
- Examples:
  - `TestThatRoundOfThirtyTwoGameHasDeterministicIDBasedOnLowestSeed`
  - `TestThatSweetSixteenGameHasDeterministicIDBasedOnLowestSeed`

**`build_final_four_test.go`**
- Tests for `BracketBuilder.buildFinalFour()`
- 8 tests covering Final Four structure and linking
- Examples:
  - `TestThatFinalFourSemifinalOneHasDeterministicID`
  - `TestThatEastRegionalChampionLinksToFinalFourSemifinalOne`
  - `TestThatBothFinalFourSemifinalsLinkToChampionshipGame`

**`to_bracket_team_test.go`**
- Tests for `BracketBuilder.toBracketTeam()`
- 1 test covering LowestSeedSeen initialization
- Example: `TestThatBracketTeamInitiallyHasLowestSeedSeenEqualToOwnSeed`

### BracketService Tests (To Be Created)

Following the same pattern:

**`select_winner_test.go`**
- Tests for `BracketService.SelectWinner()`
- Winner selection, progression, LowestSeedSeen updates

**`unselect_winner_test.go`**
- Tests for `BracketService.UnselectWinner()`
- Winner clearing and downstream game clearing

**`calculate_wins_and_byes_test.go`**
- Tests for wins/byes calculation logic
- First Four teams, bye teams, winners, losers, elimination

## Benefits of This Structure

### 1. Reduced Boilerplate

**Before (without helpers):**
```go
func TestSomething(t *testing.T) {
    // 20 lines of setup
    team1 := &models.TournamentTeam{
        ID: "team1",
        TournamentID: "test",
        SchoolID: "school1",
        Seed: 1,
        Region: "East",
        School: &models.School{
            ID: "school1",
            Name: "School 1",
        },
    }
    // ... repeat for more teams
    
    // Actual test logic
}
```

**After (with helpers):**
```go
func TestSomething(t *testing.T) {
    // GIVEN
    helper := NewBracketTestHelper()
    teams := helper.CreateStandardRegionTeams("East", 16)
    bracket := helper.CreateEmptyBracket()
    
    // WHEN & THEN
    // Actual test logic
}
```

### 2. Focused Files

- Each file tests one function
- Easy to find tests for a specific function
- Files stay under 200 lines
- Clear separation of concerns

### 3. Easy to Extend

Adding a new test:
1. Identify the function being tested
2. Open the corresponding test file
3. Add a new `TestThat{Scenario}` function
4. Use helpers to minimize setup

### 4. Maintainability

- Changes to test data creation happen in one place (`bracket_test_helpers.go`)
- Tests remain focused on behavior, not setup
- Single-assert principle easier to maintain with less boilerplate

## Running Tests

**All tests:**
```bash
cd backend
go test ./pkg/services/...
```

**Specific function tests:**
```bash
go test ./pkg/services/build_bracket_test.go ./pkg/services/bracket_builder.go ./pkg/services/bracket_test_helpers.go
```

**Single test:**
```bash
go test -run TestThatFirstFourGameForElevenSeedHasDeterministicID ./pkg/services/...
```

## Test Count Summary

**BracketBuilder Tests: 25 tests**
- BuildBracket: 1 test
- buildRegionalBracket (First Four): 7 tests
- buildRegionalBracket (Round of 64): 6 tests
- buildRegionalBracket (Regional rounds): 3 tests
- buildFinalFour: 8 tests
- toBracketTeam: 1 test

**BracketService Tests: TBD**
- SelectWinner: ~6 tests
- UnselectWinner: ~3 tests
- Calculate wins/byes: ~7 tests

**Total: ~40+ tests** covering all bracket logic with single assertions and clear naming.
