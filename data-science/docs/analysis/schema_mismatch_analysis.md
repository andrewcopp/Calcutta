# Schema Mismatch Analysis

## Data Type Mismatches Between DataFrames and Database

### 1. Predicted Game Outcomes (Silver Layer)

**DataFrame Structure:**
```
- game_id: string
- round: string ('first_four', 'round_of_64', 'round_of_32', 'sweet_16', 'elite_8', 'final_four', 'championship')
- round_order: int64
- team1_key: string
- team2_key: string
- p_matchup: float64
- p_team1_wins_given_matchup: float64
- p_team2_wins_given_matchup: float64
```

**Database Schema:**
```sql
- round: integer (NOT NULL)
- p_team1_wins: numeric(10,6)
- p_matchup: numeric(10,6)
```

**Issues:**
- ❌ `round` is a string in DataFrame but integer in database
- ❌ DataFrame has `p_team1_wins_given_matchup` but database expects `p_team1_wins`
- ❌ DataFrame has `p_team2_wins_given_matchup` but database doesn't have `p_team2_wins`

### 2. Simulated Tournaments (Bronze Layer)

**DataFrame Structure:**
```
- sim_id: int64
- team_key: string
- wins: int64
```

**Database Schema:**
```sql
- tournament_key: varchar(100) (NOT NULL)
- sim_id: integer (NOT NULL)
- team_key: varchar(100) (NOT NULL)
- wins: integer (NOT NULL)
- byes: integer (NOT NULL)
- eliminated: boolean (NOT NULL)
```

**Issues:**
- ❌ DataFrame missing `tournament_key` (needs to be added by writer)
- ❌ DataFrame missing `byes` column
- ❌ DataFrame missing `eliminated` column

## Proposed Solutions

### Option A: Transform Data in Writers (Minimal Schema Changes)

**For Predicted Game Outcomes:**
1. Map round strings to integers (inverted bracket approach):
   ```python
   round_mapping = {
       'championship': 0,      # 1 game from championship
       'final_four': 1,        # 2 games from championship
       'elite_8': 2,           # 3 games from championship
       'sweet_16': 3,          # 4 games from championship
       'round_of_32': 4,       # 5 games from championship
       'round_of_64': 5,       # 6 games from championship
       'first_four': 6,        # 7 games from championship
   }
   ```
2. Rename `p_team1_wins_given_matchup` → `p_team1_wins`
3. Drop `p_team2_wins_given_matchup` (can be calculated as 1 - p_team1_wins)

**For Simulated Tournaments:**
1. Add `tournament_key` from context
2. Calculate `byes` from tournament structure (teams that skip first_four)
3. Set `eliminated = (wins < 6)` for 64-team tournament

**Pros:**
- ✅ No schema changes needed
- ✅ Preserves existing database structure
- ✅ Works with current Go API code

**Cons:**
- ❌ Need to calculate derived fields (byes, eliminated)
- ❌ Round mapping is hardcoded for 68-team tournament

### Option B: Modify Bronze Schema (Flexible for Future)

**For Predicted Game Outcomes:**
1. Change `round` from `integer` to `varchar(20)` to store round names
2. Add `round_number` as integer for sorting/filtering
3. Keep both for flexibility

**For Simulated Tournaments:**
1. Make `byes` and `eliminated` nullable or remove them
2. These can be calculated on-demand in the API layer

**Pros:**
- ✅ More flexible for variable tournament sizes
- ✅ Easier to understand data (round names are clearer)
- ✅ Less transformation logic needed

**Cons:**
- ❌ Requires schema migration
- ❌ Need to update Go repository code
- ❌ Need to update API handlers

### Option C: Hybrid Approach (Recommended)

**For Predicted Game Outcomes:**
1. Use inverted round numbering (championship = 0, counting backwards)
2. Store round name in a separate lookup table or as metadata
3. Transform in writer: `round_name → round_number`

**For Simulated Tournaments:**
1. Make `byes` and `eliminated` nullable in schema
2. Calculate and store them if available, otherwise NULL
3. API can calculate on-demand if NULL

**Pros:**
- ✅ Minimal schema changes (just make columns nullable)
- ✅ Inverted numbering handles variable tournament sizes
- ✅ Backward compatible with existing data
- ✅ Can add calculated fields later without breaking

**Cons:**
- ❌ Still requires some schema migration
- ❌ Need to update writers to handle nullable fields

## Inverted Bracket Numbering Benefits

**Current Approach (Forward):**
- first_four = 0
- round_of_64 = 1
- round_of_32 = 2
- ...
- championship = 6

**Problem:** Breaks if tournament size changes (e.g., 64 teams instead of 68)

**Inverted Approach (Backward from Championship):**
- championship = 0
- final_four = 1
- elite_8 = 2
- sweet_16 = 3
- round_of_32 = 4
- round_of_64 = 5
- first_four = 6

**Benefits:**
- ✅ Championship is always round 0
- ✅ Works for any tournament size
- ✅ Easy to calculate "games until championship"
- ✅ Future-proof for NCAA expansion
- ✅ Smaller tournaments just have higher round numbers

## Recommendation

**Implement Option C with inverted bracket numbering:**

1. **Immediate (Writer Changes):**
   - Map round strings to inverted integers in writers
   - Add tournament_key to simulated tournaments
   - Make byes/eliminated nullable in schema (small migration)

2. **Short-term (Schema Enhancement):**
   - Add `round_name` column to predicted_game_outcomes for clarity
   - Keep `round` as integer for performance

3. **Long-term (If Needed):**
   - Create `tournament_rounds` lookup table
   - Store round metadata (name, games_count, etc.)
