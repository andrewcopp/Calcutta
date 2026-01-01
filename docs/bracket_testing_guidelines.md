# Bracket Logic Testing Guidelines

## Core Testing Principles (from engineering.md)

### 1. Single Assert Per Test
Each test should have **exactly one reason to fail**. This means:
- One assertion per test function
- No multiple THEN statements checking different things
- Avoids whack-a-mole debugging where fixing one assertion reveals another failure

### 2. TestThat{Scenario} Naming Convention
Test names must follow the format: `TestThat{Scenario}`

**Why this matters:**
- Forces you to write descriptive test names
- Makes it clear what behavior is being tested
- The wording is deliberate and should not be ignored

**Examples:**

❌ **Wrong:**
```go
func TestBracketBuilder_BuildBracket_DeterministicGameIDs(t *testing.T)
func TestBracketBuilder_FirstFourGames(t *testing.T)
func TestRoundOf64Matchups(t *testing.T)
```

✅ **Correct:**
```go
func TestThatBracketBuilderGeneratesSameGameIDsAcrossMultipleBuilds(t *testing.T)
func TestThatRegionWithTwoDuplicateSeedsCreatesExactlyTwoFirstFourGames(t *testing.T)
func TestThatRoundOfSixtyFourMatchesOneSeedAgainstSixteenSeed(t *testing.T)
func TestThatFirstFourGameForElevenSeedHasDeterministicID(t *testing.T)
func TestThatRoundOfSixtyFourSeedsSumToSeventeen(t *testing.T)
```

### 3. GIVEN / WHEN / THEN Structure
All tests follow this structure:
- **GIVEN**: Setup with clear initial state
- **WHEN**: Single action being tested
- **THEN**: Single assertion (one reason to fail)

### 4. Unit Tests Cover Pure Logic
Unit tests should cover pure business logic (deterministic inputs → outputs).

- Do not write unit tests that assert filesystem/DB/HTTP behavior.
- Unit tests must not read/write external state (DB, network/API calls, filesystem, or any external storage).
- Keep I/O in thin wrappers (handlers/CLIs/runners) and keep core logic in pure
  functions so it is easy to test.
- If you need a dependency for unit tests, inject an interface and use a fake/in-memory implementation.

## Bracket Logic Test Coverage Needed

### Deterministic Game IDs
- `TestThatBracketBuilderGeneratesSameGameIDsAcrossMultipleBuilds`
- `TestThatFirstFourGameForElevenSeedHasDeterministicID`
- `TestThatFirstFourGameForSixteenSeedHasDeterministicID`
- `TestThatRoundOfSixtyFourGameForOneSeedHasDeterministicID`
- `TestThatRoundOfThirtyTwoGameHasDeterministicIDBasedOnLowestSeed`
- `TestThatSweetSixteenGameHasDeterministicIDBasedOnLowestSeed`
- `TestThatEliteEightGameHasDeterministicIDForRegionalFinal`
- `TestThatFinalFourSemifinalOneHasDeterministicID`
- `TestThatFinalFourSemifinalTwoHasDeterministicID`
- `TestThatChampionshipGameHasDeterministicID`

### First Four (Play-in Games)
- `TestThatRegionWithNoDuplicateSeedsCreatesZeroFirstFourGames`
- `TestThatRegionWithOneDuplicateSeedCreatesOneFirstFourGame`
- `TestThatRegionWithTwoDuplicateSeedsCreatesExactlyTwoFirstFourGames`
- `TestThatFirstFourGameMatchesBothTeamsWithSameSeed`
- `TestThatFirstFourGameLinksToCorrectRoundOfSixtyFourGame`

### Round of 64 Matchups
- `TestThatRoundOfSixtyFourMatchesOneSeedAgainstSixteenSeed`
- `TestThatRoundOfSixtyFourMatchesTwoSeedAgainstFifteenSeed`
- `TestThatRoundOfSixtyFourMatchesEightSeedAgainstNineSeed`
- `TestThatRoundOfSixtyFourSeedsSumToSeventeen`
- `TestThatRoundOfSixtyFourCreatesExactlyEightGamesPerRegion`

### Regional Round Matchups (Lowest Seed Seen Logic)
- `TestThatRoundOfThirtyTwoMatchesLowestSeedsSummingToNine`
- `TestThatSweetSixteenMatchesLowestSeedsSummingToFive`
- `TestThatEliteEightMatchesLowestSeedsSummingToThree`

### Final Four Structure
- `TestThatEastRegionalChampionLinksToFinalFourSemifinalOne`
- `TestThatSouthRegionalChampionLinksToFinalFourSemifinalOne`
- `TestThatWestRegionalChampionLinksToFinalFourSemifinalTwo`
- `TestThatMidwestRegionalChampionLinksToFinalFourSemifinalTwo`
- `TestThatBothFinalFourSemifinalsLinkToChampionshipGame`

### Lowest Seed Seen Tracking
- `TestThatBracketTeamInitiallyHasLowestSeedSeenEqualToOwnSeed`
- `TestThatOneSeedWinningCarriesLowestSeedSeenOfOne`
- `TestThatNineSeedUpset tingEightSeedCarriesLowestSeedSeenOfEight`
- `TestThatWinnerInheritsMinimumLowestSeedSeenFromBothTeams`

### Winner Selection & Progression
- `TestThatSelectingWinnerSetsWinnerFieldInGame`
- `TestThatSelectingWinnerProgressesTeamToNextGame`
- `TestThatSelectingWinnerUpdatesLowestSeedSeenInNextGame`
- `TestThatUnselectingWinnerClearsWinnerField`
- `TestThatUnselectingWinnerClearsTeamFromNextGame`
- `TestThatUnselectingWinnerClearsAllDownstreamGames`

### Wins & Byes Calculation
- `TestThatTeamWithByeHasOneByeBeforeFirstGame`
- `TestThatFirstFourTeamHasZeroByes`
- `TestThatFirstFourWinnerHasOneWin`
- `TestThatTeamWinningRoundOfSixtyFourHasOneWin`
- `TestThatTeamWinningTwoGamesHasTwoWins`
- `TestThatTeamLosingGameIsMarkedEliminated`
- `TestThatTeamNotYetPlayingIsNotEliminated`

### Validation
- `TestThatBracketWithSixtyEightTeamsPassesValidation`
- `TestThatBracketWithSixtySevenTeamsFailsValidation`
- `TestThatRegionMissingSeedFiveFailsValidation`
- `TestThatRegionWithThreeDuplicateSeedsFailsValidation`
- `TestThatSelectingNonParticipantAsWinnerFails`
- `TestThatSelectingWinnerForGameWithMissingTeamFails`

## Example: Refactoring Multi-Assert Test

### Before (WRONG - Multiple Asserts)
```go
func TestBracketBuilder_FirstFourGames(t *testing.T) {
	// GIVEN a region with two 11-seeds and two 16-seeds
	teams := createRegionWithDuplicateSeeds("East", 11, 16)
	builder := NewBracketBuilder()
	bracket := &models.BracketStructure{
		TournamentID: "test",
		Games:        make(map[string]*models.BracketGame),
		Regions:      []string{"East"},
	}

	// WHEN building the regional bracket
	_, err := builder.buildRegionalBracket(bracket, "East", teams)
	if err != nil {
		t.Fatalf("Failed to build regional bracket: %v", err)
	}

	// THEN there should be exactly 2 First Four games
	firstFourCount := 0
	for _, game := range bracket.Games {
		if game.Round == models.RoundFirstFour {
			firstFourCount++
		}
	}
	if firstFourCount != 2 {
		t.Errorf("Expected 2 First Four games, got %d", firstFourCount)
	}

	// THEN First Four games should have deterministic IDs
	game11 := bracket.Games["East-first_four-11"]
	if game11 == nil {
		t.Error("Expected First Four game for 11-seeds")
	}

	// THEN teams should match
	if game11.Team1.Seed != 11 || game11.Team2.Seed != 11 {
		t.Errorf("Wrong seeds: %d vs %d", game11.Team1.Seed, game11.Team2.Seed)
	}
}
```

### After (CORRECT - Single Assert Per Test)
```go
func TestThatRegionWithTwoDuplicateSeedsCreatesExactlyTwoFirstFourGames(t *testing.T) {
	// GIVEN a region with two 11-seeds and two 16-seeds
	teams := createRegionWithDuplicateSeeds("East", 11, 16)
	builder := NewBracketBuilder()
	bracket := &models.BracketStructure{
		TournamentID: "test",
		Games:        make(map[string]*models.BracketGame),
		Regions:      []string{"East"},
	}

	// WHEN building the regional bracket
	_, err := builder.buildRegionalBracket(bracket, "East", teams)
	if err != nil {
		t.Fatalf("Failed to build regional bracket: %v", err)
	}

	// THEN there should be exactly 2 First Four games
	firstFourCount := 0
	for _, game := range bracket.Games {
		if game.Round == models.RoundFirstFour {
			firstFourCount++
		}
	}
	if firstFourCount != 2 {
		t.Errorf("Expected 2 First Four games, got %d", firstFourCount)
	}
}

func TestThatFirstFourGameForElevenSeedHasDeterministicID(t *testing.T) {
	// GIVEN a region with two 11-seeds
	teams := createRegionWithDuplicateSeeds("East", 11, 16)
	builder := NewBracketBuilder()
	bracket := &models.BracketStructure{
		TournamentID: "test",
		Games:        make(map[string]*models.BracketGame),
		Regions:      []string{"East"},
	}

	// WHEN building the regional bracket
	_, err := builder.buildRegionalBracket(bracket, "East", teams)
	if err != nil {
		t.Fatalf("Failed to build regional bracket: %v", err)
	}

	// THEN First Four game for 11-seeds has ID 'East-first_four-11'
	if bracket.Games["East-first_four-11"] == nil {
		t.Error("Expected First Four game with ID 'East-first_four-11'")
	}
}

func TestThatFirstFourGameMatchesBothTeamsWithSameSeed(t *testing.T) {
	// GIVEN a region with two 11-seeds
	teams := createRegionWithDuplicateSeeds("East", 11, 16)
	builder := NewBracketBuilder()
	bracket := &models.BracketStructure{
		TournamentID: "test",
		Games:        make(map[string]*models.BracketGame),
		Regions:      []string{"East"},
	}

	// WHEN building the regional bracket
	_, err := builder.buildRegionalBracket(bracket, "East", teams)
	if err != nil {
		t.Fatalf("Failed to build regional bracket: %v", err)
	}

	// THEN both teams in the 11-seed First Four game have seed 11
	game11 := bracket.Games["East-first_four-11"]
	if game11 == nil {
		t.Fatal("First Four game not found")
	}
	if game11.Team1.Seed != 11 || game11.Team2.Seed != 11 {
		t.Errorf("Expected both teams to have seed 11, got %d vs %d", 
			game11.Team1.Seed, game11.Team2.Seed)
	}
}
```

## Benefits of Single-Assert Tests

1. **Precise Failure Identification**: When a test fails, you know exactly what broke
2. **No Cascading Failures**: Fixing one issue doesn't reveal hidden failures
3. **Better Test Names**: Forces descriptive naming that documents behavior
4. **Easier Debugging**: Small, focused tests are easier to understand and fix
5. **Better Coverage Visibility**: Can see exactly which scenarios are tested

## Next Steps

1. Refactor all existing bracket tests to follow these guidelines
2. Write comprehensive tests covering all scenarios listed above
3. Ensure each test has exactly one assertion
4. Use `TestThat{Scenario}` naming for all tests
5. Run tests to verify coverage: `go test -v ./pkg/services/...`
