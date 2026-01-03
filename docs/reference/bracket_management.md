# Tournament Bracket Management System

## Overview

This document describes the bracket-based tournament management system that replaces manual wins/byes updates with a zero-sum, validated bracket progression system.

## Problem Statement

The previous admin system allowed manual updates to team wins, byes, and elimination status without validation. This created several issues:
- **No zero-sum enforcement**: Admins could give everyone 8 wins
- **No structural validation**: Updates didn't respect tournament bracket structure
- **Error-prone**: Manual tracking of 68 teams across 7 rounds is tedious
- **No safeguards**: Easy to make mistakes that break the tournament state

## Solution Architecture

### Core Principles

1. **Tournament structure lives in code, not SQL**: The 68-team NCAA tournament has well-defined rules that are better expressed as business logic than database constraints
2. **Zero-sum by design**: Winner selection automatically updates bracket state, ensuring only valid progressions
3. **Declarative state management**: Wins/byes/eliminated are calculated from bracket state, not manually set
4. **Validation at every step**: All operations validate bracket integrity before persisting

### Components

#### 1. Bracket Models (`pkg/models/bracket.go`)

**BracketStructure**: Complete tournament bracket representation
- Maps game IDs to game objects
- Tracks regional structure
- Stores Final Four configuration

**BracketGame**: Individual game in the bracket
- Teams (nullable until determined)
- Winner (nullable until selected)
- Next game reference (bracket progression)
- Round and region metadata

**BracketTeam**: Team reference in bracket context
- Links to TournamentTeam
- Includes seed and region for display

**BracketValidator**: Validates bracket operations
- Winner selection validation
- Bracket progression validation
- Ensures structural integrity

#### 2. Bracket Builder (`pkg/services/bracket_builder.go`)

Generates complete bracket structure from tournament teams following NCAA rules:

**First Four (variable games per region)**:
- Rule: If two teams from the same region have the same seed, that's a play-in game
- Each region can have 0-4 play-in games depending on seed distribution
- Winners advance to Round of 64
- Teams without play-in games get a bye

**Regional Brackets (4 regions, 16+ teams each)**:
- Each region must have seeds 1-16 represented (at least 16 teams)
- Additional teams beyond 16 create play-in games
- Round of 64: Seeds matched where sum = 17 (1 vs 16, 2 vs 15, etc.)
- Round of 32: Winners advance
- Sweet 16: Regional semifinals
- Elite 8: Regional finals

**Final Four**:
- Configurable regional matchups (e.g., East vs West, South vs Midwest)
- Semifinals → Championship

#### 3. Bracket Service (`pkg/services/bracket_service.go`)

Manages bracket state and operations:

**GetBracket**: Retrieves current bracket state
- Builds bracket structure from teams
- Applies current results
- Returns complete bracket

**SelectWinner**: Selects game winner and progresses bracket
- Validates winner is a participant
- Sets winner in current game
- Progresses winner to next game
- Recalculates all team stats (wins/byes/eliminated)
- Validates bracket integrity

**UnselectWinner**: Removes winner selection
- Clears winner from game
- Recursively clears downstream games
- Recalculates team stats

**ValidateBracketSetup**: Validates tournament is ready for bracket
- Exactly 68 teams
- 17 teams per region
- Correct seed distribution (4 of each seed, except 6 each for seeds 11 and 16)

#### 4. Tournament Model Extension

Added Final Four configuration to `Tournament` model:
```go
FinalFourTopLeft     string // e.g., "East"
FinalFourBottomLeft  string // e.g., "West"
FinalFourTopRight    string // e.g., "South"
FinalFourBottomRight string // e.g., "Midwest"
```

This allows different Final Four matchups each year.

## API Endpoints

### GET `/api/tournaments/{id}/bracket`
Returns complete bracket structure with current state.

**Response**:
```json
{
  "tournamentId": "...",
  "regions": ["East", "West", "South", "Midwest"],
  "games": [
    {
      "gameId": "...",
      "round": "round_of_64",
      "region": "East",
      "team1": { "teamId": "...", "name": "UConn", "seed": 1 },
      "team2": { "teamId": "...", "name": "Stetson", "seed": 16 },
      "winner": null,
      "nextGameId": "...",
      "nextGameSlot": 1,
      "canSelect": true
    }
  ],
  "finalFour": {
    "topLeftRegion": "East",
    "bottomLeftRegion": "West",
    "topRightRegion": "South",
    "bottomRightRegion": "Midwest"
  }
}
```

### POST `/api/tournaments/{tournamentId}/bracket/games/{gameId}/winner`
Selects a winner for a game.

**Request**:
```json
{
  "winnerTeamId": "team-uuid"
}
```

**Response**: Updated bracket structure

**Validation**:
- Winner must be one of the two participating teams
- Both teams must be present before selecting winner
- Bracket progression must remain valid

### DELETE `/api/tournaments/{tournamentId}/bracket/games/{gameId}/winner`
Removes winner selection and clears downstream games.

**Response**: Updated bracket structure

### GET `/api/tournaments/{id}/bracket/validate`
Validates tournament setup is ready for bracket generation.

**Response**:
```json
{
  "valid": true,
  "errors": []
}
```

## Zero-Sum Enforcement

The system enforces zero-sum properties through:

1. **Structural validation**: Winners can only be selected from participating teams
2. **Automatic progression**: Winner selection automatically places team in next game
3. **Calculated stats**: Wins/byes/eliminated are derived from bracket state, not manually set
4. **Cascade clearing**: Unselecting a winner clears all downstream games

This makes it **impossible** to create invalid states like:
- Giving everyone 8 wins
- Having a team win without playing
- Having inconsistent bracket progression
- Having teams advance without winning

## Future Enhancements

### 1. Scenario Exploration

The same bracket logic can be used for "what-if" analysis:

```go
// Explore different outcomes without persisting
func (s *BracketService) ExploreScenario(
    ctx context.Context,
    tournamentID string,
    hypotheticalWinners map[string]string, // gameID -> winnerTeamID
) (*models.BracketStructure, error)
```

Use cases:
- **Entry-focused**: "What needs to happen for my entry to win?"
- **Calcutta-focused**: "What are the possible leaderboard outcomes?"
- **User exploration**: Interactive bracket picker showing portfolio impact

### 2. Backtracking Algorithm

Find all paths to victory for an entry:

```go
func FindPathsToVictory(
    bracket *BracketStructure,
    entryID string,
    teams []string,
) [][]GameOutcome
```

### 3. Webhook Integration

Replace manual admin updates with automatic game result ingestion:
- Subscribe to tournament API webhooks
- Automatically call `SelectWinner` when games complete
- Admin UI becomes read-only monitoring tool

### 4. Optimistic Locking

Add version field to prevent concurrent updates:
```go
type BracketStructure struct {
    Version int `json:"version"`
    // ...
}
```

## Database Migrations

The Tournament model extension requires a migration:

```sql
ALTER TABLE tournaments 
ADD COLUMN final_four_top_left VARCHAR(50),
ADD COLUMN final_four_bottom_left VARCHAR(50),
ADD COLUMN final_four_top_right VARCHAR(50),
ADD COLUMN final_four_bottom_right VARCHAR(50);
```

Default values for existing tournaments:
```sql
UPDATE tournaments SET
    final_four_top_left = 'East',
    final_four_bottom_left = 'West',
    final_four_top_right = 'South',
    final_four_bottom_right = 'Midwest'
WHERE final_four_top_left IS NULL;
```

## Testing

Unit tests validate core bracket logic:
- Seed matchup rules (1 vs 16, 2 vs 15, etc.)
- Winner selection validation
- Bracket progression validation
- Wins/byes calculation from bracket state

Run tests:
```bash
cd backend
go test ./pkg/models -v -run TestBracket
```

## Frontend Integration

The frontend admin UI should:

1. **Display visual bracket**: Show all games organized by round and region
2. **Enable winner selection**: Click on a team to select as winner
3. **Show progression**: Highlight how winners flow through bracket
4. **Undo capability**: Allow unselecting winners to fix mistakes
5. **Validation feedback**: Show which games can have winners selected
6. **Real-time updates**: Recalculate portfolio scores as bracket updates

Example UI flow:
```
[First Four] → [Round of 64] → [Round of 32] → [Sweet 16] → [Elite 8] → [Final Four] → [Championship]

East Region:
  (1) UConn vs (16) Stetson [Select Winner]
  (8) FAU vs (9) Northwestern [Select Winner]
  ...

Winner: UConn → advances to play winner of (8)/(9) game
```

## Migration Path

1. **Phase 1**: Deploy backend with new bracket endpoints (✅ Complete)
2. **Phase 2**: Build admin UI for bracket management
3. **Phase 3**: Test with historical tournament data
4. **Phase 4**: Use for live tournament
5. **Phase 5**: Add scenario exploration features
6. **Phase 6**: Integrate with external tournament APIs

## Conclusion

This bracket-based system provides:
- ✅ Zero-sum enforcement through structural validation
- ✅ Reduced admin burden (click winners vs manual stat updates)
- ✅ Impossible to create invalid states
- ✅ Foundation for scenario exploration
- ✅ Clean separation of concerns (rules in code, state in DB)
- ✅ Comprehensive validation and testing
