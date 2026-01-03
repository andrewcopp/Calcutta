# Bracket State Management

## Problem

The original implementation tried to maintain ephemeral bracket state by progressing winners through games in memory. This approach was fragile when teams advanced through multiple rounds, especially after play-in games, causing 500 errors when selecting winners in subsequent rounds.

## Solution

**Store wins in the database and rebuild the bracket from that state on every request.**

This approach treats the bracket structure as a **derived view** of the underlying tournament state (team wins), rather than trying to maintain it as mutable state.

## Architecture

### Data Flow

```
User selects winner
    ↓
Increment team.wins in database
    ↓
Rebuild entire bracket from scratch
    ↓
Apply wins to determine game winners
    ↓
Progress winners through bracket structure
    ↓
Return updated bracket to user
```

### Key Components

#### 1. **Tournament Teams Table** (Source of Truth)

```sql
CREATE TABLE tournament_teams (
    id UUID PRIMARY KEY,
    tournament_id UUID NOT NULL,
    school_id UUID NOT NULL,
    seed INTEGER NOT NULL,
    region VARCHAR(50) NOT NULL,
    byes INTEGER NOT NULL DEFAULT 0,
    wins INTEGER NOT NULL DEFAULT 0,        -- ← Source of truth
    eliminated BOOLEAN NOT NULL DEFAULT FALSE,
    ...
);
```

The `wins` column is the **single source of truth** for bracket state.

#### 2. **SelectWinner** (Simplified)

**Before:**
```go
// Complex ephemeral state management
game.Winner = winner
nextGame.Team1 = winnerCopy
// ... lots of mutation logic
```

**After:**
```go
func (s *BracketService) SelectWinner(ctx, tournamentID, gameID, winnerTeamID) {
    // 1. Validate the selection
    bracket := s.GetBracket(ctx, tournamentID)
    game := bracket.Games[gameID]
    s.validator.ValidateWinnerSelection(game, winnerTeamID)
    
    // 2. Increment wins in database
    team := s.tournamentRepo.GetTournamentTeam(ctx, winnerTeamID)
    team.Wins++
    s.tournamentRepo.UpdateTournamentTeam(ctx, team)
    
    // 3. Rebuild bracket from updated state
    return s.GetBracket(ctx, tournamentID)
}
```

#### 3. **GetBracket** (Reconstruction)

```go
func (s *BracketService) GetBracket(ctx, tournamentID) {
    // 1. Get tournament and teams from database
    tournament := s.tournamentRepo.GetByID(ctx, tournamentID)
    teams := s.tournamentRepo.GetTeams(ctx, tournamentID)
    
    // 2. Build fresh bracket structure (deterministic)
    bracket := s.builder.BuildBracket(tournamentID, teams, finalFour)
    
    // 3. Apply current results from team wins
    s.applyCurrentResults(ctx, bracket, teams)
    
    return bracket
}
```

#### 4. **applyCurrentResults** (State Reconstruction)

This function walks through the bracket in round order and determines winners based on stored wins:

```go
func (s *BracketService) applyCurrentResults(ctx, bracket, teams) {
    teamMap := makeTeamMap(teams)
    
    // Process rounds in order
    for _, round := range [FirstFour, Round64, Round32, Sweet16, Elite8, FinalFour, Championship] {
        for _, game := range bracket.Games {
            if game.Round != round || game.Team1 == nil || game.Team2 == nil {
                continue
            }
            
            team1 := teamMap[game.Team1.TeamID]
            team2 := teamMap[game.Team2.TeamID]
            
            // Determine winner based on wins
            winsNeeded := getWinsNeededForRound(round)
            
            if team1.Wins > winsNeeded {
                game.Winner = game.Team1
                progressToNextGame(game, bracket)
            } else if team2.Wins > winsNeeded {
                game.Winner = game.Team2
                progressToNextGame(game, bracket)
            }
        }
    }
}
```

**Wins Required Per Round:**
- First Four: 0 wins (starting point for play-in teams)
- Round of 64: 0 wins (starting point for non-play-in teams)
- Round of 32: 1 win
- Sweet 16: 2 wins
- Elite 8: 3 wins
- Final Four: 4 wins
- Championship: 5 wins

## Benefits

### 1. **Eliminates Ephemeral State Issues**

No more tracking which team is in which game slot across multiple rounds. The bracket is always rebuilt from the source of truth.

### 2. **Handles Play-In Games Correctly**

Play-in winners start with 1 win, so they correctly appear in Round of 64 games and progress normally through the bracket.

### 3. **Idempotent Operations**

Calling `GetBracket` multiple times with the same database state produces identical results.

### 4. **Simplified Logic**

- `SelectWinner`: Just increment wins
- `UnselectWinner`: Just decrement wins
- No complex downstream clearing logic needed

### 5. **Easier to Debug**

State is always in the database. No need to track ephemeral mutations through multiple function calls.

### 6. **Supports Undo/Redo**

Since wins are stored, you can easily implement undo by decrementing wins, or even store a history of win changes.

## Trade-offs

### Performance

**Cost:** Rebuilding the bracket on every request requires:
- Database query for teams
- Bracket builder execution
- State reconstruction logic

**Mitigation:** 
- Bracket building is fast (68 teams, ~67 games)
- Can add caching layer if needed
- Database queries are indexed

**Reality:** For a tournament bracket admin UI with occasional updates, this is negligible.

### Complexity

**Before:** Complex ephemeral state management with many edge cases
**After:** Simple database updates + deterministic reconstruction

The reconstruction logic is straightforward and testable.

## Testing Strategy

### Unit Tests

Test the reconstruction logic in isolation:

```go
func TestThatTeamWithOneWinAppearsInRoundOf32(t *testing.T) {
    // GIVEN a team with 1 win
    team := createTeamWithWins(1)
    bracket := buildBracketWithTeam(team)
    
    // WHEN applying current results
    applyCurrentResults(bracket, []*TournamentTeam{team})
    
    // THEN the team appears in Round of 32
    game32 := findGameInRound(bracket, RoundOf32, team.ID)
    assert.NotNil(t, game32)
}
```

### Integration Tests

Test the full flow:

```go
func TestThatSelectingPlayInWinnerProgressesToRound64(t *testing.T) {
    // GIVEN a tournament with a play-in game
    tournament := createTournamentWithPlayIn()
    
    // WHEN selecting the play-in winner
    bracket := service.SelectWinner(ctx, tournament.ID, "East-first_four-11", team11a.ID)
    
    // THEN the winner appears in Round of 64
    game64 := bracket.Games["East-round_of_64-6"]
    assert.Equal(t, team11a.ID, game64.Team2.TeamID)
}
```

## Migration Notes

### Database

No schema changes required. The `wins` column already exists.

### API

No API changes required. The endpoints remain the same:
- `POST /api/tournaments/:id/bracket/winner` - Select winner
- `DELETE /api/tournaments/:id/bracket/games/:gameId/winner` - Unselect winner
- `GET /api/tournaments/:id/bracket` - Get bracket

### Frontend

No frontend changes required. The bracket structure returned is identical.

## Future Enhancements

### 1. Caching

Add a cache layer to avoid rebuilding on every request:

```go
func (s *BracketService) GetBracket(ctx, tournamentID) {
    cacheKey := fmt.Sprintf("bracket:%s:%d", tournamentID, getTournamentVersion(ctx, tournamentID))
    
    if cached := cache.Get(cacheKey); cached != nil {
        return cached
    }
    
    bracket := s.buildBracketFromDatabase(ctx, tournamentID)
    cache.Set(cacheKey, bracket, 5*time.Minute)
    return bracket
}
```

### 2. Event Sourcing

Store a history of winner selections:

```sql
CREATE TABLE bracket_events (
    id UUID PRIMARY KEY,
    tournament_id UUID NOT NULL,
    game_id VARCHAR(100) NOT NULL,
    winner_team_id UUID NOT NULL,
    action VARCHAR(20) NOT NULL, -- 'select' or 'unselect'
    created_at TIMESTAMP NOT NULL
);
```

This enables:
- Audit trail
- Undo/redo
- Time-travel debugging

### 3. Optimistic Locking

Add version numbers to prevent concurrent updates:

```sql
ALTER TABLE tournament_teams ADD COLUMN version INTEGER NOT NULL DEFAULT 1;
```

```go
UPDATE tournament_teams 
SET wins = $1, version = version + 1 
WHERE id = $2 AND version = $3
```

## Summary

The new architecture treats the bracket as a **pure function of team wins**:

```
Bracket = f(Teams, Wins)
```

This eliminates ephemeral state issues, simplifies the code, and makes the system more reliable and easier to test.
