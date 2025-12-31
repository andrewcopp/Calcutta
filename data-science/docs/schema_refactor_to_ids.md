# Schema Refactor: String Keys → Integer IDs

## Current Schema (String Keys)

```
bronze_tournaments
  - tournament_key (varchar PK) ← "ncaa-tournament-2025"
  
bronze_teams
  - team_key (varchar PK) ← "ncaa-tournament-2025:duke"
  - tournament_key (varchar FK)
  
bronze_simulated_tournaments
  - tournament_key (varchar FK)
  - team_key (varchar FK)
  
silver_predicted_game_outcomes
  - tournament_key (varchar FK)
  - team1_key (varchar FK)
  - team2_key (varchar FK)
  
bronze_calcuttas
  - calcutta_key (varchar PK)
  - tournament_key (varchar FK)
  
bronze_entry_bids
  - calcutta_key (varchar FK)
  - team_key (varchar FK)
```

## New Schema (Integer IDs)

```
bronze_tournaments
  - id (bigserial PK)
  - season (int NOT NULL)
  - tournament_name (varchar NOT NULL)
  - UNIQUE(season, tournament_name)
  
bronze_teams
  - id (bigserial PK)
  - tournament_id (bigint FK → bronze_tournaments.id)
  - school_slug (varchar NOT NULL)
  - school_name (varchar NOT NULL)
  - seed (int NOT NULL)
  - region (varchar NOT NULL)
  - byes (int)
  - kenpom_net, kenpom_o, kenpom_d, kenpom_adj_t (numeric)
  - UNIQUE(tournament_id, school_slug)
  
bronze_simulated_tournaments
  - id (bigserial PK)
  - tournament_id (bigint FK → bronze_tournaments.id)
  - sim_id (int NOT NULL)
  - team_id (bigint FK → bronze_teams.id)
  - wins (int NOT NULL)
  - byes (int NOT NULL)
  - eliminated (boolean NOT NULL)
  - INDEX(tournament_id, sim_id)
  - INDEX(team_id)
  
silver_predicted_game_outcomes
  - id (bigserial PK)
  - tournament_id (bigint FK → bronze_tournaments.id)
  - game_id (varchar NOT NULL) ← Keep for bracket structure
  - round (int NOT NULL)
  - team1_id (bigint FK → bronze_teams.id)
  - team2_id (bigint FK → bronze_teams.id)
  - p_team1_wins (numeric(10,6) NOT NULL)
  - p_matchup (numeric(10,6) NOT NULL DEFAULT 1.0)
  - model_version (varchar)
  - UNIQUE(tournament_id, game_id)
  
bronze_calcuttas
  - id (bigserial PK)
  - tournament_id (bigint FK → bronze_tournaments.id)
  - calcutta_name (varchar NOT NULL)
  - budget_points (int NOT NULL DEFAULT 100)
  - UNIQUE(tournament_id, calcutta_name)
  
bronze_entry_bids
  - id (bigserial PK)
  - calcutta_id (bigint FK → bronze_calcuttas.id)
  - team_id (bigint FK → bronze_teams.id)
  - entry_name (varchar NOT NULL)
  - bid_amount_points (int NOT NULL)
  - INDEX(calcutta_id)
  - INDEX(team_id)
```

## Benefits of ID-based Schema

1. **Performance**: Integer joins are faster than string joins
2. **Storage**: Smaller indexes and foreign keys
3. **Simplicity**: No need to construct/parse composite keys
4. **Flexibility**: Can change naming without breaking references
5. **Standard Practice**: Aligns with typical database design

## Migration Strategy

### Phase 1: Add ID Columns (Non-Breaking)
- Add `id` columns to all tables as `bigserial`
- Keep existing `*_key` columns temporarily
- Populate IDs for existing data

### Phase 2: Add New FK Columns
- Add `tournament_id`, `team_id`, `calcutta_id` columns
- Populate from existing `*_key` columns via joins
- Add indexes on new ID columns

### Phase 3: Update Constraints
- Add foreign key constraints on new ID columns
- Add unique constraints on natural keys (season+name, tournament_id+slug)

### Phase 4: Drop Old Columns (Breaking)
- Drop old `*_key` foreign key constraints
- Drop old `*_key` columns
- Update primary keys if needed

### Phase 5: Update Application Code
- Update Python writers to use IDs
- Update Go repositories to use IDs
- Update API handlers to use IDs

## Natural Key Lookups

Since we're removing string keys, we need efficient lookups:

```sql
-- Find tournament by season
SELECT id FROM bronze_tournaments WHERE season = 2025;

-- Find team by tournament and school
SELECT id FROM bronze_teams 
WHERE tournament_id = ? AND school_slug = 'duke';

-- Find calcutta by tournament and name
SELECT id FROM bronze_calcuttas
WHERE tournament_id = ? AND calcutta_name = 'Main Pool';
```

## Writer Pattern

Python writers will need to:
1. Lookup or insert tournament → get `tournament_id`
2. Lookup or insert teams → get `team_id`s
3. Use IDs for all subsequent inserts

```python
# Example: Write simulated tournaments
def write_simulated_tournaments(season, simulations_df):
    # 1. Get or create tournament
    tournament_id = get_or_create_tournament(season)
    
    # 2. Get team IDs (assumes teams already written)
    team_ids = get_team_ids_by_slug(tournament_id, simulations_df['team_slug'])
    
    # 3. Write simulations with IDs
    simulations_df['tournament_id'] = tournament_id
    simulations_df['team_id'] = team_ids
    
    # Insert using IDs
    insert_simulations(simulations_df[['tournament_id', 'sim_id', 'team_id', 'wins', ...]])
```

## Backward Compatibility

To support gradual migration:
- Keep `*_key` columns initially as computed/generated columns
- Or create views with old schema for backward compatibility
- Once all code is updated, drop compatibility layer

## Implementation Order

1. ✅ Design new schema (this document)
2. Create migration SQL files
3. Update Python writers with ID-based logic
4. Update Go repositories and queries
5. Run migration on database
6. Test pipeline end-to-end
7. Remove old `*_key` columns

## Open Questions

1. **Keep game_id as string?** Yes - it encodes bracket structure (e.g., "East-round_of_64-1")
2. **Keep entry_name in bids table?** Yes - denormalized for convenience
3. **Add lookup tables?** Could add `schools` table with global school IDs, but defer for now
