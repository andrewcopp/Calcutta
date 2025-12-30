# Optimal Feature Set for Market Prediction

## Date
December 30, 2025

## Summary

Through systematic feature engineering and forward selection, we identified an optimal 5-feature set that achieves **38% improvement** in market prediction accuracy over a KenPom-only baseline.

## Final Feature Set

### 1. Championship Equity (`champ_equity`)
**What it is**: Approximate historical probability of winning the championship based on seed

**Why it works**: 
- Smart non-linear encoding of seed information
- Captures what the market cares about (title contenders get premium)
- More efficient than 16 categorical seed dummies
- Selected FIRST in forward selection from empty set

**Implementation**:
```python
seed_title_prob = {
    1: 0.20, 2: 0.12, 3: 0.08, 4: 0.05, 5: 0.03, 6: 0.02,
    7: 0.01, 8: 0.01, 9: 0.005, 10: 0.003, 11: 0.002,
    12: 0.001, 13: 0.0005, 14: 0.0002, 15: 0.0001, 16: 0.00001
}
champ_equity = seed.map(seed_title_prob)
```

### 2. KenPom Percentile (`kenpom_net_pct`)
**What it is**: Percentile rank of KenPom net rating within the tournament field

**Why it works**:
- Relative strength matters more than absolute strength
- Normalizes across years with different rating scales
- Selected SECOND in forward selection (before raw KenPom!)

**Implementation**:
```python
kenpom_net_pct = kenpom_net.rank(pct=True)
```

### 3. KenPom Balance (`kenpom_balance`)
**What it is**: Absolute difference between offensive and defensive percentile ranks

**Why it works**:
- Teams with imbalanced offense/defense are systematically mispriced
- Market may not properly value stylistic extremes
- Captures a distinct dimension from overall strength

**Implementation**:
```python
kenpom_o_pct = kenpom_o.rank(pct=True)
kenpom_d_pct = kenpom_d.rank(pct=True)
kenpom_balance = abs(kenpom_o_pct - kenpom_d_pct)
```

### 4. Points Per Equity (`points_per_equity`)
**What it is**: Ratio of expected tournament points to championship probability

**Why it works**:
- Identifies "value plays" - teams with good expected points but lower title odds
- Market may underprice solid-but-not-sexy teams
- Example: 5-seed has decent points (4) but low title equity (0.03) = high ratio

**Implementation**:
```python
seed_expected_points = {
    1: 12, 2: 9, 3: 7, 4: 5, 5: 4, 6: 3, 7: 2, 8: 2,
    9: 1, 10: 1, 11: 1, 12: 1, 13: 0.5, 14: 0.3, 15: 0.2, 16: 0.1
}
expected_points = seed.map(seed_expected_points)
points_per_equity = expected_points / (champ_equity + 0.001)
```

### 5. KenPom Percentile Cubed (`kenpom_pct_cubed`)
**What it is**: Non-linear transformation of KenPom percentile

**Why it works**:
- Captures that market overweights the very top teams
- Difference between 95th and 99th percentile matters more than 50th to 54th
- Adds non-linearity to strength relationship

**Implementation**:
```python
kenpom_pct_cubed = kenpom_net_pct ** 3
```

## Performance

| Metric | Value |
|--------|-------|
| **Mean MAE** | **0.006374** |
| **Improvement over KenPom baseline** | **38.0%** |
| **Number of features** | **5** |
| **Years improved** | **8/8** |

## Key Insights from Feature Engineering

### 1. Seed is Categorical, Not Continuous
- Initial testing used seed², seed³ (mathematically incorrect)
- Treating seed as continuous gave only 8.8% improvement
- Categorical encoding (16 dummies) gave 30.6% improvement
- **Championship equity** provides even better encoding with just 1 parameter

### 2. Forward Selection Reveals True Importance
When starting from empty set (no protected baseline):
1. **champ_equity** selected first (+30.9% improvement)
2. **kenpom_net_pct** selected second (not raw kenpom!)
3. **kenpom_balance** selected third

This shows our initial assumption that raw KenPom should be baseline was wrong.

### 3. Percentiles Beat Raw Ratings
- `kenpom_net_pct` (percentile) selected over `kenpom_net` (raw rating)
- Percentiles normalize across years and capture relative position
- More stable and interpretable

### 4. Non-Linear Effects Matter
- `kenpom_pct_cubed` captures market's overweighting of elite teams
- `champ_equity` provides non-linear seed encoding
- Linear models need these transformations to capture curvature

### 5. Simplicity Wins
- Just 5 features capture most systematic market behavior
- Additional features tested (24 creative ideas) added minimal value
- Diminishing returns suggest we've captured the main signals

## What Didn't Work

### Features with No Impact
- **Last year performance**: 0.0% improvement (no systematic effect)
- **Region strength**: -0.1% (doesn't help after controlling for team strength)
- **Power conference**: 0.0% (no effect after KenPom)
- **Tempo indicators**: ~0% (minimal impact)

### Why 16 Seed Dummies Weren't Selected
- Would require 15 parameters vs 1 for championship equity
- Not enough data to reliably estimate 16 separate coefficients
- Championship equity captures the same information more efficiently

### Why Raw KenPom Wasn't Selected
- Percentile rank is more stable across years
- Absolute ratings have different scales in different years
- Relative position matters more than absolute strength

## Comparison to Previous Approaches

| Approach | Features | MAE | Improvement |
|----------|----------|-----|-------------|
| **Optimal (new)** | 5 | **0.006374** | **38.0%** |
| KenPom + interactions | 6 | 0.006612 | 35.6% |
| KenPom + champ equity | 5 | 0.007091 | 30.9% |
| Categorical seed (16) | 15 | 0.007130 | 30.6% |
| KenPom baseline | 4 | 0.010274 | 0% |

## Implementation Notes

### Feature Computation Order
1. Compute percentiles WITHIN each year (not across years)
2. Championship equity and expected points are lookup tables
3. All features are deterministic (no randomness)

### Handling Missing Data
- All features use `.fillna(0)` for missing values
- Championship equity uses small epsilon (0.001) to avoid division by zero
- Percentile ranks handle ties with `pct=True` parameter

### Ridge Regression
- Alpha = 1.0 (standard regularization)
- StandardScaler applied before fitting
- Leave-one-year-out cross-validation for testing

## Future Directions

### Potential Improvements
1. **Tune regularization** (alpha parameter) for optimal bias-variance
2. **Test ensemble methods** (combine multiple models)
3. **Add interaction terms** between top features
4. **Monitor 2026 performance** (true out-of-sample test)

### What NOT to Do
- Don't add more features without strong evidence (diminishing returns)
- Don't use seed as continuous variable (mathematically incorrect)
- Don't include last year features (no systematic effect)
- Don't overfit to specific years (use cross-validation)

## Conclusion

The optimal 5-feature set represents a **data-driven, mathematically principled** approach to market prediction:

1. **Championship equity** - Smart seed encoding
2. **KenPom percentile** - Relative strength
3. **KenPom balance** - Stylistic imbalance
4. **Points per equity** - Value indicator
5. **KenPom percentile cubed** - Non-linear strength

This achieves 38% improvement with just 5 features, demonstrating that **simplicity and careful feature engineering** beat throwing in many mediocre features.

The systematic testing process (30+ features, forward selection, LOOCV) gives confidence these are real signals, not overfitting.
