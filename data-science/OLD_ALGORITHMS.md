# Old Algorithms from calcutta_ds

This document catalogs algorithms from the pre-refactor `calcutta_ds` package that may be valuable for testing or comparison purposes.

## 1. Waterfill Equal Allocation (`allocation.py`)

**Algorithm**: Equal distribution with max-per-team caps
**Status**: Replaced by greedy optimizer in `recommended_entry_bids.py`
**Preserve for**: Baseline comparison testing

```python
def waterfill_equal(k: int, budget: float, max_per_team: float) -> List[float]:
    """
    Distributes budget equally across k teams, respecting max_per_team cap.
    
    Algorithm:
    - Initialize all bids to 0
    - Iterate through teams, incrementing each by 1 until budget exhausted
    - Skip teams that have reached max_per_team cap
    
    Use case: Simple baseline for testing greedy optimizer performance
    """
```

**Key characteristics**:
- Deterministic
- Simple round-robin allocation
- No optimization for expected value
- Good for testing that greedy optimizer outperforms naive approaches

---

## 2. Old Simulation with Historical Winners (`sim.py::expected_simulation`)

**Algorithm**: Monte Carlo simulation with optional historical winner override
**Status**: Replaced by `simulated_tournaments.py` + `simulated_entry_outcomes.py`
**Preserve for**: Testing competitor filtering and historical replay

```python
def expected_simulation(
    ...
    use_historical_winners: bool,
    competitor_entry_keys: Optional[List[str]] = None,
) -> Dict[str, object]:
    """
    Old simulation that supports:
    1. Historical winner override (use actual tournament results)
    2. Competitor filtering (exclude specific entries from standings)
    
    Use cases:
    - Replaying historical tournaments exactly
    - Testing with subset of competitors
    """
```

**Key features**:
- `use_historical_winners`: Forces games to use actual results from `winner_team_key`
- `competitor_entry_keys`: Filters standings to only include specified entries
- Returns dict with metrics (not DataFrame)

**Preserve for**:
- Testing historical replay accuracy
- Validating competitor exclusion logic (avoiding leakage when author is in pool)

---

## 3. Team Points Scenarios Matrix (`sim.py::simulate_team_points_scenarios`)

**Algorithm**: Returns numpy matrix of team points across simulations
**Status**: Replaced by cached tournament approach
**Preserve for**: Matrix-based analysis and vectorized operations

```python
def simulate_team_points_scenarios(
    ...
) -> Tuple[List[str], np.ndarray]:
    """
    Returns:
    - team_keys: List of team identifiers
    - matrix: (n_sims x n_teams) numpy array of team points
    
    Use case: Vectorized portfolio optimization or covariance analysis
    """
```

**Key characteristics**:
- Returns numpy array instead of DataFrame
- Efficient for matrix operations
- Good for correlation/covariance analysis between teams

**Preserve for**:
- Testing alternative portfolio optimization approaches
- Covariance-based diversification strategies

---

## 4. Old Portfolio Builder with Equal Allocation (`backtest_scaffold_runner.py::_build_portfolio_equal`)

**Algorithm**: Top-k selection + waterfill allocation
**Status**: Replaced by greedy optimizer
**Preserve for**: Baseline comparison

```python
def _build_portfolio_equal(
    df: pd.DataFrame,
    score_col: str,
    k: int,
    budget: float,
    max_per_team: float,
    min_bid: float,
) -> Tuple[pd.DataFrame, List[Dict[str, object]]]:
    """
    Simple portfolio construction:
    1. Rank teams by score_col
    2. Select top k teams
    3. Allocate budget equally (waterfill)
    
    Use case: Baseline for testing greedy optimizer improvements
    """
```

**Preserve for**:
- A/B testing against greedy optimizer
- Verifying that optimization provides value over naive top-k selection

---

## 5. Realized Results Calculator (`backtest_scaffold_runner.py::_realized_results`)

**Algorithm**: Calculates actual performance using historical data
**Status**: Partially replaced by investment_report
**Preserve for**: Historical validation

```python
def _realized_results(
    tables: Dict[str, pd.DataFrame],
    calcutta_key: str,
    points_mode: str,
    portfolio_rows: List[Dict[str, object]],
    sim_entry_key: str,
    budget: float,
) -> Tuple[Dict[str, object], pd.DataFrame]:
    """
    Computes what actually happened with a given portfolio.
    Uses real tournament results (wins/byes from teams.parquet).
    
    Use case: Backtesting - compare predicted vs actual performance
    """
```

**Preserve for**:
- Backtesting framework
- Validating predictions against actual outcomes

---

## Recommendations

### Keep as Test Fixtures

1. **`waterfill_equal`** - Simple baseline for optimizer testing
2. **`use_historical_winners` flag** - For exact historical replay
3. **`competitor_entry_keys` filtering** - For leakage-free backtesting
4. **`simulate_team_points_scenarios`** - For matrix-based analysis

### Can Remove

1. **Old CLI** (`cli_backtest_scaffold.py`) - Fully replaced by `moneyball/cli.py`
2. **Old runner** (`backtest_scaffold_runner.py`) - Fully replaced by `moneyball/pipeline/runner.py`
3. **Duplicate simulation logic** - Core simulation now in `simulated_tournaments.py`

### Migration Strategy

1. Create test fixtures in `tests/fixtures/` for preserved algorithms
2. Add comparison tests: greedy vs waterfill, cached vs old simulation
3. Remove obsolete modules after extracting valuable algorithms
4. Keep `calcutta_ds` core utilities (bracket, points, standings, io) as shared library
