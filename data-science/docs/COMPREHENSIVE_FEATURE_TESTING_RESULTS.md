# Comprehensive Feature Testing Results

## Date
December 30, 2025

## Summary

Systematically tested all candidate features using sklearn Ridge regression with leave-one-year-out cross-validation. Discovered that **KenPom interaction features** provide a massive 27.7% improvement over seed-only baseline.

## Testing Framework

**Method**: Leave-one-year-out cross-validation
- Train on 7 years, test on 1 held-out year
- Repeat for all 8 years (2017-2025)
- Use sklearn Ridge (alpha=1.0) with StandardScaler
- Compare against seed-only baseline

**Baseline**: Seed only
- MAE: 0.009199
- Correlation: 0.7437
- This is surprisingly strong - the market primarily prices by seed

## Results

| Feature | MAE Improvement | Corr Improvement | Years Improved | Recommendation |
|---------|----------------|------------------|----------------|----------------|
| **KenPom + Interactions** | **+27.7%** | **+0.118** | **8/8** | ✓ **IMPLEMENT** |
| Within-Seed Ranking | +3.6% | +0.014 | 7/8 | ✓ Implement |
| Brand Tax | +3.2% | +0.010 | 6/8 | ✓ Implement |
| Upset Chic | +1.6% | +0.004 | 7/8 | ⚠ Consider |
| KenPom (raw) | -1.5% | +0.002 | 1/8 | ✗ Skip |
| All KenPom (no interactions) | -1.9% | +0.002 | 2/8 | ✗ Skip |
| Region Strength | -0.1% | -0.000 | 2/8 | ✗ Skip |
| Last Year Performance | 0.0% | +0.000 | 0/8 | ✗ Skip |

## Key Findings

### 1. KenPom Interactions are Critical (27.7% improvement!)

**Features**: `seed_sq`, `kenpom_x_seed`

The interaction between seed and KenPom captures non-linear market behavior:
- High-seed teams with good KenPom get disproportionately more bids
- Low-seed teams with bad KenPom get disproportionately fewer bids
- The market doesn't just add seed + KenPom linearly

**Why this matters**: This is the single most important feature for market prediction. Without it, we're missing the fundamental way the market prices teams.

### 2. Seed Alone is a Strong Baseline

Seed-only achieves 0.7437 correlation - better than adding raw KenPom features!

This reveals that:
- The market primarily prices teams by seed
- KenPom is a secondary signal
- Adding raw KenPom without interactions actually hurts predictions

### 3. Market Behavior Features Still Matter

Even after KenPom interactions, market behavior features add value:
- **Within-Seed Ranking**: +3.6% (7/8 years)
- **Brand Tax**: +3.2% (6/8 years)
- **Upset Chic**: +1.6% (7/8 years)

These capture systematic biases that KenPom doesn't explain.

### 4. Region Strength and Last Year Don't Help

- **Region Strength**: No improvement (-0.1%)
- **Last Year Performance**: No improvement (0.0%)

The market doesn't systematically misprice based on these factors.

## Recommended Feature Set

Based on testing, the optimal feature set is:

**Core Features**:
1. `seed`
2. `kenpom_net`
3. `kenpom_o`
4. `kenpom_d`
5. `kenpom_adj_t`

**Interaction Features**:
6. `seed_sq` (seed squared)
7. `kenpom_x_seed` (KenPom × seed interaction)

**Market Behavior Features**:
8. `is_blue_blood` (brand tax)
9. `is_upset_seed` (upset chic for seeds 10-12)
10. `kenpom_rank_within_seed_norm` (within-seed ranking)

**Expected Combined Impact**: ~35% MAE improvement over seed-only baseline

## Implementation Priority

**Phase 1: Critical (Implement Immediately)**
- KenPom interaction features (`seed_sq`, `kenpom_x_seed`)
- This alone provides 27.7% improvement

**Phase 2: High Value (Implement Next)**
- Within-seed ranking (+3.6%)
- Brand tax (+3.2%)

**Phase 3: Marginal (Consider)**
- Upset chic (+1.6%)

**Skip**:
- Region strength (no improvement)
- Last year performance (no improvement)

## Next Steps

1. **Update production model** to use recommended feature set
2. **Regenerate all investment reports** with new features
3. **Measure portfolio performance improvement**
4. **Consider additional interaction terms** (e.g., `seed × region`, `kenpom × brand`)

## Technical Notes

- Used sklearn Ridge instead of custom implementation (more robust)
- StandardScaler applied to all features before training
- Alpha=1.0 for regularization (could tune this further)
- Leave-one-year-out ensures no data leakage
- All features tested individually to isolate effects

## Validation

All tests run successfully with consistent results across years. The framework is now available in `scripts/feature_tests/` for future feature testing.
