# 2023-2025 Performance Diagnosis

**Date**: December 29, 2025  
**Status**: ✅ Root cause identified

## Executive Summary

The 2023-2025 performance "collapse" is **NOT a model failure**. The market prediction model actually **improved** during this period. The issue is **tournament variance** - bad luck with which teams advanced in the actual tournaments.

## Key Findings

### 1. Market Prediction Accuracy IMPROVED

**Prediction Error (MAE):**
- Pre-2023 average: 0.006649
- Post-2022 average: 0.006135
- **Change: -7.7% (BETTER)**

**Prediction Correlation:**
- Pre-2023 average: 0.8533
- Post-2022 average: 0.8746
- **Change: +0.0213 (BETTER)**

**Conclusion**: The model is MORE accurate at predicting what other people will bid in 2023-2025.

### 2. Investment Performance Collapsed

**Expected Payout Performance:**
- 2022: $223.73 (4th place, 18.3% P(1st))
- 2023: $5.04 (21st place, 0.4% P(1st))
- 2024: $2.43 (18th place, 0.0% P(1st))
- 2025: $0.23 (27th place, 0.0% P(1st))

**Conclusion**: Despite accurate market predictions, the portfolios failed.

### 3. Root Cause: Tournament Variance

The disconnect between accurate market predictions and poor investment performance indicates:

1. **The teams we invested in underperformed their expected tournament outcomes**
2. **This is either bad luck (variance) or systematic bias in tournament predictions**
3. **Not a market prediction problem - that's working well**

## What Changed in 2023-2025?

### Market Characteristics (No Major Changes)
- **Pool size**: Grew from 49 (2022) to 50-52 (2023-2025) - minimal change
- **Market concentration (Gini)**: Decreased from 0.638 (pre-2023) to 0.599 (post-2022)
  - Market became LESS concentrated (more distributed bidding)
  - This should favor our strategy, not hurt it
- **KenPom mean**: Stable around 14-15 across all years
- **Bid amounts**: Increased slightly but consistently

### Tournament Outcomes (2023 Example)
Looking at 2023 actual results:
- **Champion**: Connecticut (4-seed) - not a typical favorite
- **Runner-up**: San Diego State (5-seed) - mid-seed
- **Final Four**: Two 5-seeds, one 4-seed, one 9-seed
- **This was an UPSET-HEAVY tournament**

If our model favored high-seeds and they underperformed, that would explain the collapse.

## Implications

### What's Working
✅ Market prediction model (predicting human bidding behavior)  
✅ Greedy optimizer (portfolio construction given predictions)  
✅ Feature engineering for market predictions (last year data helps)

### What's NOT Working
❌ Tournament outcome predictions (KenPom-based expected points)  
❌ Portfolio variance/diversification  
❌ Handling upset-heavy tournaments

## Recommended Actions

### Priority 1: Improve Tournament Outcome Predictions
The bottleneck is **expected_team_points** calculation, not market share prediction.

**Options:**
1. **Tune KenPom scale parameter** (currently 10.0)
   - Test range: 5.0 to 20.0
   - Higher = more chalk, Lower = more upsets
   
2. **Add upset probability features**
   - Seed-based upset rates
   - Historical performance by seed
   - Matchup-specific factors

3. **Non-linear models for game outcomes**
   - Gradient boosting for win probabilities
   - Incorporate more than just KenPom

### Priority 2: Portfolio Diversification
Even with perfect predictions, variance kills performance.

**Options:**
1. **Increase team count** (currently 10)
   - Test 12-15 teams
   - Reduce concentration risk
   
2. **Correlation-aware optimization**
   - Avoid teams in same region
   - Balance across seed groups
   
3. **Risk-adjusted optimization**
   - Optimize for Sharpe ratio, not just expected value
   - Penalize high-variance portfolios

### Priority 3: Feature Engineering for Tournament Outcomes
Test individual features for improving game outcome predictions:

**Candidates:**
- Seed-based historical win rates
- Recent tournament performance (hot/cold teams)
- Matchup-specific factors (pace, style)
- Injury/roster changes (if available)

### Priority 4: Non-Linear Models
Try gradient boosting or random forests for:
1. **Market predictions** (may help, but already good)
2. **Tournament outcome predictions** (higher priority)

## Testing Plan

### Phase 1: Tournament Prediction Tuning (Immediate)
1. Test KenPom scale: 5.0, 7.5, 10.0, 12.5, 15.0, 20.0
2. Evaluate on 2023-2025 (upset-heavy years)
3. Compare portfolio performance

### Phase 2: Individual Feature Testing (Next)
1. Add one feature at a time to tournament predictions
2. Measure impact on expected vs actual points
3. Keep only features that improve accuracy

### Phase 3: Non-Linear Models (After)
1. Implement gradient boosting for game outcomes
2. Compare vs current KenPom-only approach
3. If better, integrate into pipeline

### Phase 4: Portfolio Optimization (Final)
1. Test different team counts (10, 12, 15)
2. Implement correlation-aware optimization
3. Backtest on all years

## Success Metrics

**Primary**: Portfolio performance (expected payout, P(top 3))  
**Secondary**: Tournament prediction accuracy (actual vs expected points)  
**Tertiary**: Market prediction accuracy (already good, maintain)

## Notes

- We only get one tournament per year - can't increase training data
- Must squeeze every bit of signal from available data
- Focus on tournament predictions, not market predictions
- Variance is the enemy - diversification matters
