# Portfolio Optimizer Improvements

## Current State

The greedy optimizer maximizes expected value without considering:
1. **Correlation between teams** (teams in same region have correlated outcomes)
2. **Portfolio variance** (risk of underperformance)
3. **Tail risk** (probability of finishing last)

## Proposed Improvements

### 1. Correlation-Aware Optimization (High Priority)

**Approach**: Use simulated tournament data to calculate actual correlations

```python
# From simulated_tournaments.parquet (5000 simulations):
# For each team pair (i, j):
#   correlation[i,j] = corr(points_i, points_j) across simulations

# Portfolio variance:
#   var(portfolio) = sum_i sum_j (weight_i * weight_j * cov[i,j])
#   where cov[i,j] = correlation[i,j] * std_i * std_j

# Optimization objective:
#   maximize: expected_return - lambda * sqrt(variance)
#   subject to: budget, min/max teams, min/max per team
```

**Key Insights:**
- Teams in **same region** have negative correlation (one eliminates the other)
- Teams in **different regions** have lower correlation (more independent)
- **1-seeds from different regions** are good diversification (all can make Final Four)
- **Mid-seeds in same region** are bad (correlated early exits)

**Expected Impact:**
- Reduce variance by 20-30%
- Improve P(top 3) by preferring diversified portfolios
- May slightly reduce expected value but increase consistency

### 2. Mean-CVaR Optimization (Medium Priority)

**Approach**: Optimize for expected value while limiting downside risk

```python
# CVaR (Conditional Value at Risk) = expected loss in worst 5% of outcomes
# From simulations:
#   CVaR_5% = mean(payout | payout < 5th percentile)

# Optimization objective:
#   maximize: expected_payout - lambda * CVaR_5%
#   subject to: budget, min/max teams, min/max per team
```

**Key Insights:**
- Penalizes portfolios that can catastrophically fail
- Useful for avoiding last place (which pays $0)
- More appropriate than variance for asymmetric payouts

**Expected Impact:**
- Reduce P(bottom 10%) by 30-40%
- Slightly reduce P(1st) but improve P(top 3)
- Better for risk-averse strategies

### 3. Simulation-Based Kelly Criterion (Lower Priority)

**Approach**: Use actual simulation distributions, not assumptions

```python
# Current Kelly uses assumed variance
# Better: calculate from simulations

# For each team:
#   edge = (expected_payout - cost) / cost
#   variance = var(payout) from simulations
#   kelly_fraction = edge / variance

# Then allocate budget proportional to kelly_fractions
```

**Expected Impact:**
- More accurate than current Kelly implementation
- Natural diversification when outcomes are uncertain
- May perform similarly to correlation-aware approach

## Implementation Priority

**Phase 1: Correlation-Aware Greedy** (Recommended First)
- Modify greedy optimizer to consider correlation
- Add correlation penalty to marginal ROI calculation
- Test on historical data

**Phase 2: Feature Engineering** (Parallel Track)
- Implement brand tax, upset chic, within-seed ranking
- These improve market predictions → better team selection
- Independent of optimizer improvements

**Phase 3: Advanced Optimization** (If Phase 1 succeeds)
- Implement full mean-variance or mean-CVaR optimization
- May require quadratic programming solver
- Higher complexity, potentially higher reward

## Validation Approach

For each improvement:
1. Backtest on 2017-2025 data
2. Compare metrics: P(1st), P(top 3), mean payout, variance
3. Check if improvement is consistent across years
4. Validate that simulated performance matches actual outcomes

## Feature Engineering (Parallel Track)

While improving optimizer, also implement market prediction features:

**High Priority:**
1. **Brand tax**: Indicator for blue-blood programs (Duke, UNC, Kentucky, Kansas)
2. **Upset chic**: Indicator for trendy upset seeds (10-12)
3. **Within-seed ranking**: KenPom rank within seed group

**Medium Priority:**
4. **Region strength**: Average KenPom of top-4 teams in region
5. **Path difficulty**: Expected opponent strength to Final Four

**Expected Impact:**
- Improve market prediction MAE by 10-15%
- Better identify undervalued teams
- Compounds with optimizer improvements
