# Tournament Data Modeling

## Overview
This document outlines the considerations and potential approaches for modeling tournament data in the Calcutta application. The model needs to support both completed tournaments and in-progress tournaments, as well as enable simulation of potential outcomes.

## Key Requirements

### 1. Tournament States
- States should be inferred from game statuses rather than stored
- Function `getTournamentState(tournamentId)` determines current state:
  - **Completed**: All games have winners
  - **In-Progress**: Some games completed, some in progress or scheduled
  - **Future**: Games scheduled but not started

### 2. Game Representation
- Core model: `TournamentGame`
  - id
  - tournament_id (FK to Tournament)
  - team1_id (FK to TournamentTeam, nullable)
  - team2_id (FK to TournamentTeam, nullable)
  - tipoff_time (datetime)
  - sort_order (integer)
  - team1_score (integer, nullable)
  - team2_score (integer, nullable)
  - next_game_id (FK to TournamentGame, nullable)
  - next_game_slot (integer: 1 or 2)
  - is_final (boolean)

- Game status is inferred from:
  - Future: tipoff_time is in the future
  - In Progress: tipoff_time is in the past, is_final is false
  - Completed: is_final is true

- Winner is determined by comparing team1_score and team2_score when is_final is true

### 3. Tournament Structure
- Support for 68-team tournament (with play-in games)
- Handle byes through single-participant games
- Track progression through rounds
- Support for different tournament formats (if needed)

### 4. Simulation Capabilities
Two distinct types of simulations:

#### A. Predictive Simulations (Backend)
- Use KenPom or similar metrics for probability-based predictions
- Calculate ExpectedScores and PredictedScores
- Store results in `CalcuttaPredictions` table
- Update after each game completion
- Models:
  ```
  CalcuttaPredictions
    ├── TournamentId
    ├── TeamId
    ├── ExpectedWins
    ├── PredictedWins
    ├── KenPomRank
    └── LastUpdated
  ```

#### B. Interactive Scenarios (Frontend)
- User-driven winner selection
- Real-time leaderboard impact calculation
- Store scenario preferences in frontend/local storage
- Focus on "what-if" paths to victory
- Models:
  ```
  UserScenarios
    ├── UserId
    ├── TournamentId
    ├── ScenarioName
    ├── GameResults (JSON)
    └── LastModified
  ```

## Data Models

### Core Tournament Structure
```
Tournament
  ├── id
  ├── name
  ├── year
  ├── rounds
  ├── created_at
  └── updated_at

TournamentTeam
  ├── id
  ├── tournament_id (FK to Tournament)
  ├── school_id (FK to School)
  ├── seed
  ├── byes
  ├── wins
  ├── eliminated
  ├── created_at
  └── updated_at

TournamentGame
  ├── id
  ├── tournament_id (FK to Tournament)
  ├── team1_id (FK to TournamentTeam, nullable)
  ├── team2_id (FK to TournamentTeam, nullable)
  ├── tipoff_time (datetime)
  ├── sort_order (integer)
  ├── team1_score (integer, nullable)
  ├── team2_score (integer, nullable)
  ├── next_game_id (FK to TournamentGame, nullable)
  ├── next_game_slot (integer: 1 or 2)
  ├── is_final (boolean)
  ├── created_at
  └── updated_at
```

### Simulation Support
```
CalcuttaPredictions
  ├── TournamentId
  ├── TeamId
  ├── ExpectedWins
  ├── PredictedWins
  ├── KenPomRank
  └── LastUpdated

UserScenarios
  ├── UserId
  ├── TournamentId
  ├── ScenarioName
  ├── GameResults (JSON)
  └── LastModified
```

## Key Considerations

### 1. Maximum Score Calculation
- Function to calculate maximum possible score for a portfolio
- Examines bracket structure for potential team matchups
- Prevents impossible scenarios (e.g., two portfolio teams winning championship)
- Updates as tournament progresses

### 2. Performance Considerations
- 2^67 - 1 possible tournament outcomes (too many to store)
- Focus on user-driven scenarios rather than exhaustive simulations
- Balance between frontend and backend calculations
- Efficient storage of game and participant data

### 3. Data Integrity
- Maintain clear separation between actual and simulated results
- Ensure accurate historical data preservation
- Handle tournament delays or schedule changes
- Support for different tournament formats

## Testing Strategy

### 1. Game Status Tests
- Test inference of game status based on tipoff_time and is_final
- Test edge cases (e.g., games with no tipoff time, games with no scores)
- Test status transitions as tournament progresses

### 2. Winner Determination Tests
- Test winner determination based on scores when is_final is true
- Test edge cases (e.g., tied scores, missing scores)
- Test winner propagation to next game

### 3. Tournament Progression Tests
- Test bracket structure integrity
- Test team advancement through rounds
- Test maximum score calculation with various portfolio configurations

### 4. Simulation Tests
- Test predictive simulation accuracy
- Test interactive scenario creation and management
- Test leaderboard updates based on simulated results

## Questions to Consider

1. How do we handle reseeding in different regions?
2. What happens if a tournament format changes?
3. How do we maintain data integrity during simulations?
4. How do we handle historical data vs current tournament?
5. What's the best way to represent byes in the data model?
6. How do we handle tournament delays or schedule changes?
7. How do we efficiently calculate and update maximum possible scores?
8. What metrics should we use for predictive simulations?

## Next Steps

1. Define core data models
2. Create sample tournament structures
3. Design simulation interface
4. Plan migration strategy for existing data
5. Consider performance implications
6. Design API endpoints for tournament management
7. Implement maximum score calculation logic
8. Develop predictive simulation algorithms
9. Write comprehensive tests for game status and winner determination 