# Critical Bug Fix: Simulated Tournaments Data Corruption

## Date
December 29, 2025

## Summary
Fixed a critical bug in `simulate_tournaments()` that was causing teams to accumulate impossibly high win counts (e.g., Duke averaging 98 wins per simulation instead of ~4). This bug corrupted all investment reports and made the 2023-2025 performance appear catastrophically bad when it was actually reasonable.

## Root Cause

### The Bug
In `moneyball/models/simulated_tournaments.py`, the merge operation at lines 98-108 was only matching on `game_id`:

```python
# BEFORE (BUGGY):
games_graph = games.merge(
    predicted_game_outcomes[
        [
            "game_id",
            "p_team1_wins_given_matchup",
            "p_team2_wins_given_matchup",
        ]
    ],
    on="game_id",
    how="left",
)
```

This created a **cartesian product** when `predicted_game_outcomes` contained multiple possible matchups for the same `game_id` (which it does for all games beyond Round 1, since we don't know which teams will advance).

For example:
- `games.parquet`: 67 rows (actual bracket structure)
- `predicted_game_outcomes.parquet`: 935 rows (all possible matchups)
- After merge: 935 rows (one per possible matchup, not per actual game)

The simulation then iterated over all 935 rows, treating each as a separate game and incrementing win counts for every possible matchup.

### The Fix
Changed the merge to match on `game_id` + `team1_key` + `team2_key`:

```python
# AFTER (FIXED):
games_graph = games.merge(
    predicted_game_outcomes[
        [
            "game_id",
            "team1_key",
            "team2_key",
            "p_team1_wins_given_matchup",
            "p_team2_wins_given_matchup",
        ]
    ],
    on=["game_id", "team1_key", "team2_key"],
    how="left",
)
```

This ensures exactly one row per game in the bracket, matching the specific teams that are actually playing.

## Impact

### Before Fix (Corrupted Data)
- Duke 2025: 79-116 wins per simulation (avg 98)
- Iowa State 2022: 2-19 wins per simulation (avg 10.7)
- Expected points calculations were completely wrong
- Portfolio performance metrics were meaningless

### After Fix (Correct Data)
- Duke 2025: 0-5 wins per simulation (avg 3.87)
- Iowa State 2022: 0-6 wins per simulation (reasonable distribution)
- Expected points match predicted values within Monte Carlo variance

### Performance Comparison

| Year | P(1st) Before | P(1st) After | Change |
|------|---------------|--------------|--------|
| 2017 | 31.9% | 0.04% | -31.9% |
| 2018 | 0.0% | 66.8% | +66.8% |
| 2019 | 94.9% | 9.7% | -85.2% |
| 2021 | 21.7% | 6.2% | -15.5% |
| 2022 | 47.4% | 4.3% | -43.1% |
| 2023 | 0.2% | 0.5% | +0.3% |
| 2024 | 0.0% | 8.1% | +8.1% |
| 2025 | 0.0% | 1.1% | +1.1% |

**Key Insight**: The "before" numbers were completely unreliable due to corrupted simulation data. The "after" numbers represent actual expected performance based on correct tournament simulations.

## Unit Tests Added

Added comprehensive unit tests to `tests/test_simulated_tournaments.py`:

1. **`TestThatSimulatedTournamentsWinsAreBounded`**: Ensures no team can win more games than exist in the bracket
2. **`TestThatSimulatedTournamentsMergeCorrectly`**: Specifically tests that multiple matchups per game_id don't create duplicate games
3. **`TestThatSimulatedTournamentsProducesReasonableWinDistribution`**: Validates that high-probability teams win more often

These tests will prevent this bug from recurring.

## Lessons Learned

1. **Always validate output ranges**: Win counts should be bounded by the number of games
2. **Be careful with merges**: When merging DataFrames with duplicate keys, specify all join columns
3. **Unit test critical calculations**: Expected points and win distributions should be tested
4. **Sanity check simulation outputs**: 98 wins per simulation for a team that can play at most 6 games is clearly wrong

## Files Modified

- `moneyball/models/simulated_tournaments.py`: Fixed merge operation (lines 98-111)
- `tests/test_simulated_tournaments.py`: Added 3 new unit tests
- All tournament simulation files regenerated for 2017-2025
- All investment reports regenerated with correct data

## Verification

```bash
# Regenerate all tournaments
for year in 2017 2018 2019 2021 2022 2023 2024 2025; do
  rm -f out/$year/derived/tournaments.parquet
  python -m moneyball.cli simulate-tournaments out/$year --regenerate
done

# Regenerate all investment reports
bash scripts/regenerate_all_reports.sh

# Run unit tests
python -m unittest tests.test_simulated_tournaments -v
```

All tests pass. Performance metrics are now based on correct tournament simulations.
