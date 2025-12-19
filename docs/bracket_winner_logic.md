# Bracket Winner Determination Logic

## Problem Solved

When selecting a play-in game winner, they were automatically being marked as the winner of their next Round of 64 game. This was incorrect - winning a game should progress you to the next game, but not automatically win it.

## Root Cause

The original logic compared absolute wins between opponents:
```go
if team1.Wins > team2.Wins {
    winner = game.Team1
}
```

This failed because:
- Play-in winner: 1 win, 0 byes (progress = 1)
- Round of 64 opponent: 0 wins, 1 bye (progress = 1)
- System saw 1 win > 0 wins → incorrectly marked play-in winner as Round of 64 winner

## Solution: Minimum Required Progress

Each round requires a minimum `(wins + byes)` to participate. To **win** a game, a team must have **more** than the minimum (meaning they advanced past that round).

### Minimum Progress Requirements

```go
func (s *BracketService) getMinProgressForRound(round models.BracketRound) int {
    switch round {
    case models.RoundFirstFour:
        return 0 // Starting point for play-in teams
    case models.RoundOf64:
        return 1 // Need 1 (either 1 bye or 1 win from First Four)
    case models.RoundOf32:
        return 2 // Need 2 total progress
    case models.RoundSweet16:
        return 3 // Need 3 total progress
    case models.RoundElite8:
        return 4 // Need 4 total progress
    case models.RoundFinalFour:
        return 5 // Need 5 total progress
    case models.RoundChampionship:
        return 6 // Need 6 total progress
    }
}
```

### Winner Determination Logic

```go
minRequired := s.getMinProgressForRound(round)
team1Progress := team1.Wins + team1.Byes
team2Progress := team2.Wins + team2.Byes

var winner *models.BracketTeam
if team1Progress > minRequired && team2Progress >= minRequired {
    // team1 won this game and advanced
    winner = game.Team1
} else if team2Progress > minRequired && team1Progress >= minRequired {
    // team2 won this game and advanced
    winner = game.Team2
}
// If both are exactly at minRequired, they just arrived - game not played yet
```

## Examples

### Example 1: Play-in Winner Arrives at Round of 64

**Setup:**
- Team 11a: 1 win, 0 byes → progress = 1
- Team 6: 0 wins, 1 bye → progress = 1
- Round of 64 requires minProgress = 1

**Logic:**
- team11a progress (1) is NOT > minRequired (1) ✗
- team6 progress (1) is NOT > minRequired (1) ✗
- **Result: No winner** ✓

Both teams just arrived at the game. Game hasn't been played yet.

### Example 2: Team Wins Round of 64

**Setup:**
- Team 1: 1 win, 1 bye → progress = 2
- Team 16: 0 wins, 1 bye → progress = 1
- Round of 64 requires minProgress = 1

**Logic:**
- team1 progress (2) > minRequired (1) ✓
- team16 progress (1) >= minRequired (1) ✓
- **Result: Team 1 is winner** ✓

Team 1 has progressed beyond this round (won the game).

### Example 3: Play-in Winner Wins Round of 64

**Setup:**
- Team 11a: 2 wins, 0 byes → progress = 2
- Team 6: 0 wins, 1 bye → progress = 1
- Round of 64 requires minProgress = 1

**Logic:**
- team11a progress (2) > minRequired (1) ✓
- team6 progress (1) >= minRequired (1) ✓
- **Result: Team 11a is winner** ✓

Team 11a won First Four (1 win) then won Round of 64 (2 wins total).

### Example 4: Both Teams at Round of 32

**Setup:**
- Team 1: 1 win, 1 bye → progress = 2
- Team 8: 1 win, 1 bye → progress = 2
- Round of 32 requires minProgress = 2

**Logic:**
- team1 progress (2) is NOT > minRequired (2) ✗
- team8 progress (2) is NOT > minRequired (2) ✗
- **Result: No winner** ✓

Both teams just arrived at Round of 32. Game hasn't been played yet.

### Example 5: Team Wins Round of 32

**Setup:**
- Team 1: 2 wins, 1 bye → progress = 3
- Team 8: 1 win, 1 bye → progress = 2
- Round of 32 requires minProgress = 2

**Logic:**
- team1 progress (3) > minRequired (2) ✓
- team8 progress (2) >= minRequired (2) ✓
- **Result: Team 1 is winner** ✓

Team 1 has progressed to Sweet 16.

## Key Insights

1. **Arrival vs Victory**: Having exactly `minRequired` progress means you've arrived at the game. Having MORE means you've won it and moved on.

2. **Progress Tracking**: `wins + byes` tracks total tournament progress, regardless of path (play-in or bye).

3. **Path Independence**: A play-in team with 2 wins and a bye team with 1 win + 1 bye both have progress = 2, so they can face each other in Round of 32.

4. **No Rewind Needed**: We don't need to track historical state - current progress tells us everything.

## Test Coverage

Created comprehensive tests in `apply_current_results_test.go`:

1. ✅ `TestThatPlayInWinnerAppearsInRoundOfSixtyFourButDoesNotAutoWin`
2. ✅ `TestThatTeamsWithEqualWinsHaveNoWinnerDetermined`
3. ✅ `TestThatTeamWithMoreWinsIsMarkedAsWinner`
4. ✅ `TestThatWinnerProgressesToNextGame`
5. ✅ `TestThatLowestSeedSeenIsUpdatedWhenWinnerProgresses`
6. ✅ `TestThatFirstFourWinnerHasLowestSeedSeenEqualToSeed`
7. ✅ `TestThatMultipleRoundsAreProcessedInOrder`

All tests follow `engineering.md` guidelines:
- Single assertion per test
- `TestThat{Scenario}` naming
- GIVEN/WHEN/THEN structure
- Deterministic outcomes

## Benefits

✅ **Correct play-in handling** - Winners progress but don't auto-win next game  
✅ **Clear mental model** - Minimum progress thresholds per round  
✅ **No state rewinding** - Current progress is sufficient  
✅ **Path independent** - Play-in and bye teams handled uniformly  
✅ **Comprehensive tests** - All edge cases covered
