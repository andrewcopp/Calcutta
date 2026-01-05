# Artifact Structure Proposal

## Problem Statement

Current artifacts mix production data with debug data, creating opportunities for:
1. **Cross-contamination**: Wrong fields being used in downstream calculations
2. **Unclear ownership**: What field comes from which step?
3. **Debugging difficulty**: Can't inspect intermediate steps without polluting production flow

Example issues found:
- `expected_team_points` calculation has inconsistent multipliers (30x for Duke, 85x for Liberty)
- `recommended_entry_bids.parquet` contains fields from multiple pipeline steps
- Tournament simulation caps at 5 wins for some teams (Duke, Auburn) but not others (Houston, Florida)

## Proposed Structure

### Step 1: Tournament Performance Prediction

**Production Artifact**: `tournament_value.parquet`
```
Columns:
- team_key (str): Unique team identifier
- expected_points_per_entry (float): Expected points this team will earn for a single entry that owns 100% of this team
```

**Debug Artifact**: `tournament_value_debug.parquet`
```
Columns:
- team_key (str)
- expected_points_per_entry (float)
- variance_points (float)
- std_points (float)
- p_championship (float)
- p_final_four (float)
- avg_wins (float)
- max_wins (int)
- win_distribution (json): {0: count, 1: count, ...}
```

### Step 2: Market Prediction

**Production Artifact**: `market_prediction.parquet`
```
Columns:
- team_key (str): Unique team identifier
- predicted_market_share (float): Predicted fraction of total pool this team will cost (0.0 to 1.0)
```

**Debug Artifact**: `market_prediction_debug.parquet`
```
Columns:
- team_key (str)
- predicted_market_share (float)
- feature_1 (float)
- feature_2 (float)
- ... (all features used in model)
- model_confidence (float)
```

### Step 3: Portfolio Construction

**Production Artifact**: `recommended_bids.parquet`
```
Columns:
- team_key (str): Unique team identifier
- bid_amount_points (int): How many points to bid on this team
```

**Debug Artifact**: `recommended_bids_debug.parquet`
```
Columns:
- team_key (str)
- bid_amount_points (int)
- expected_points_per_entry (float): From tournament_value.parquet
- predicted_market_share (float): From market_prediction.parquet
- predicted_cost_points (float): predicted_market_share * total_pool_points
- ownership_fraction (float): bid_amount_points / predicted_cost_points
- expected_return_points (float): expected_points_per_entry * ownership_fraction
- roi (float): expected_return_points / bid_amount_points
- score (float): Internal optimization score
- strategy (str): Which strategy selected this team
```

### Step 4: Entry Simulation

**Production Artifact**: `simulated_outcomes.parquet`
```
Columns:
- sim_id (int): Simulation identifier
- entry_points (float): Total points earned by our entry in this simulation
- rank (int): Finishing position (1 = first place)
```

**Debug Artifact**: `simulated_outcomes_debug.parquet`
```
Columns:
- sim_id (int)
- entry_points (float)
- rank (int)
- team_key (str): One row per team per sim
- team_wins (int)
- team_points (float)
- ownership_fraction (float)
- contribution_points (float): team_points * ownership_fraction
```

## Key Principles

1. **Production artifacts have ONLY the minimum fields needed for the next step**
   - Prevents accidental use of wrong fields
   - Clear contract between steps

2. **Debug artifacts have ALL the information for inspection**
   - Can trace back any calculation
   - Can validate intermediate steps
   - Can identify bugs

3. **Naming convention**:
   - Production: `{step_name}.parquet` (e.g., `tournament_value.parquet`)
   - Debug: `{step_name}_debug.parquet` (e.g., `tournament_value_debug.parquet`)

4. **Each step reads ONLY from previous production artifacts**
   - Step 2 reads: `tournament_value.parquet`
   - Step 3 reads: `tournament_value.parquet`, `market_prediction.parquet`
   - Step 4 reads: `recommended_bids.parquet`, `tournament_value.parquet`

## Migration Plan

1. **Phase 1**: Create new production artifacts alongside existing ones
2. **Phase 2**: Update each step to write both production and debug artifacts
3. **Phase 3**: Update each step to read from new production artifacts
4. **Phase 4**: Remove old artifacts once validated

## Immediate Bugs to Fix (After Migration)

1. **Tournament simulation**: Duke makes finals 26% but wins 0% (impossible)
2. **Points scoring**: Use actual payout structure (50, 100, 150, 200, 250, 300 incremental)
3. **Ownership calculation**: Fix `expected_team_points` to be consistent across teams
4. **Championship probability**: Verify sums to 100% across all teams
