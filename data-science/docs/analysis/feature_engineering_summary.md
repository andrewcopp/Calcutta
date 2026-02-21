# Feature Engineering Summary - Market Prediction Model

**Date**: December 30, 2024  
**Objective**: Improve market prediction model to better identify over/under invested teams

## Final Model: Z-score Squared + Cubed

### Features (6 total)

1. **champ_equity** - Championship probability (smart seed encoding)
   - Maps seed to historical championship probability
   - Non-linear seed weighting that matches market behavior

2. **kenpom_net_zscore** - KenPom z-score normalization
   - Captures magnitude of differences (not just rank)
   - Handles top-heavy years better (e.g., 2025)

3. **kenpom_net_zscore_sq** - KenPom z-score squared
   - Emphasizes elite teams
   - Non-linear effect for historically strong teams

4. **kenpom_net_zscore_cubed** - KenPom z-score cubed
   - Additional non-linearity at extremes
   - Captures that dominant teams behave differently

5. **kenpom_balance** - Offensive/defensive imbalance
   - Absolute difference between O and D percentiles
   - Market reacts to stylistic imbalance

6. **points_per_equity** - Value play indicator
   - Expected points / championship equity
   - Identifies teams with good ROI potential

### Performance vs Previous Model

| Metric | Old (Percentile) | New (Z-score²+³) | Improvement |
|--------|------------------|------------------|-------------|
| **MAE** | 0.006246 | 0.006145 | **+1.61%** |
| **Top Seeds Opp Cost** | 0.8759 | 0.7697 | **+12.12%** |
| **Asymmetric Loss** | 5.1627 | 5.0373 | **+2.43%** |
| **RMSE** | 0.009919 | 0.009671 | **+2.50%** |
| **Weighted MAE** | 0.012804 | 0.012490 | **+2.45%** |
| **R²** | 0.7269 | 0.7405 | **+1.86%** |

### Key Improvements

1. **12% better at finding underpriced top seeds (1-4)**
   - This is the most important metric for value detection
   - Example: Auburn 2025 (1-seed at 3.55% vs predicted 4.87%)

2. **Better overall accuracy**
   - 1.61% improvement in MAE
   - 2.50% improvement in RMSE (penalizes large errors)

3. **Handles top-heavy years**
   - 2025 had historically strong 1-seeds (36.20 KenPom vs 29.18 avg)
   - Z-score captures absolute strength, not just relative rank

## Testing Methodology

### Evaluation Framework
- **Cross-validation**: Leave-one-year-out (LOOCV)
- **Years tested**: 2017, 2018, 2019, 2021, 2022, 2023, 2024, 2025
- **Model**: Ridge regression (alpha=1.0)
- **Preprocessing**: StandardScaler on features

### Metrics Used

**Standard Metrics:**
- MAE (Mean Absolute Error) - average prediction error
- RMSE (Root Mean Squared Error) - penalizes large errors
- R² (coefficient of determination) - variance explained

**Value-Focused Metrics:**
- **Opportunity Cost** - sum of errors where we overpredicted (missed cheap teams)
- **Overpay Risk** - sum of errors where we underpredicted (bought expensive teams)
- **Asymmetric Loss** - 2x penalty for overpredictions (missing value is worse)
- **Top Seeds Opportunity Cost** - opportunity cost on seeds 1-4 only
- **Large Miss Penalty** - squared penalty for errors >1%

## Features Tested and Rejected

### Market Memory Features (0% improvement)
- Prior year investment
- Championship wins last year
- Disappointment (underperformance)
- Missed tournament last year
- **Result**: No systematic market memory effects

### Regional Investment Patterns (correlation ~0)
- Total investment by region
- Average KenPom of top teams in region
- **Result**: No systematic regional bias

### Rank Within Seed (seeds 1-5) (-0.04%)
- Ranking teams within each seed by KenPom
- **Result**: No improvement, market already prices this in

### Raw Seed Number (-7.01%)
- Using seed as numeric feature instead of champ_equity
- **Result**: Much worse, non-linear seed weighting is critical

### Percentile-only Features (baseline)
- kenpom_net_pct (percentile rank)
- **Result**: Loses magnitude information, worse for value detection

## Model Comparison: Gradient Boosting vs Ridge

Tested gradient boosting (sklearn), random forest, and Ridge regression:

| Model | MAE | Result |
|-------|-----|--------|
| Ridge (optimal 5) | 0.006246 | **Best** |
| RandomForest (engineered) | 0.006337 | +1.5% worse |
| GradientBoosting (engineered) | 0.006608 | +6% worse |

**Conclusion**: Ridge regression wins due to:
- Small dataset (~544 observations)
- Regularization prevents overfitting
- Simple linear model with good features > complex model

## Why Z-score Works Better Than Percentile

### The Problem with Percentile
- Auburn 2025: KenPom 38.15 → 100th percentile
- Auburn 2019: KenPom 35.66 → 100th percentile
- **Model sees them as identical**, but 2025 was historically stronger

### The Solution: Z-score Normalization
- Auburn 2025: (38.15 - 16.97) / std = +2.5 std above mean
- Auburn 2019: (35.66 - 15.11) / std = +2.3 std above mean
- **Model captures that 2025's top teams were historically elite**

### Non-linear Terms (squared, cubed)
- Market overweights big differences at the top
- Elite teams (2+ std above mean) get exponentially more attention
- Squared and cubed terms capture this non-linearity

## 2025 Case Study: Top-Heavy Year

**2025 KenPom Statistics:**
- 1-seeds mean: 36.20 (vs 29.18 historical avg) - **+24% higher**
- Top 4 seeds mean: 29.07 (vs 24.90 historical avg) - **+16.7% higher**
- Overall mean: 16.97 (vs 14.84 historical avg) - **+14.4% higher**

**Market Behavior:**
- Houston (1-seed): 9.81% of pool (overpriced)
- Duke (1-seed): 8.94% of pool (overpriced)
- Florida (1-seed): 8.00% of pool (overpriced)
- **Auburn (1-seed): 3.55% of pool (underpriced!)**

**Model Performance:**
- Old model (percentile): Predicted Auburn at 5.61% (missed the opportunity)
- New model (z-score): Predicts Auburn at 4.87% (closer, but still overpredicts)
- **Key insight**: Model can't predict narrative-driven mispricings (brand effects)

## Implementation Details

**File**: `moneyball/models/predicted_market_share.py`

**Changes**:
1. Replaced `kenpom_net_pct` with `kenpom_net_zscore`
2. Added `kenpom_net_zscore_sq` (squared term)
3. Added `kenpom_net_zscore_cubed` (cubed term)
4. Removed `kenpom_pct_cubed` (percentile cubed)

**Feature computation**:
```python
kenpom_mean = base["kenpom_net"].mean()
kenpom_std = base["kenpom_net"].std()
base["kenpom_net_zscore"] = (base["kenpom_net"] - kenpom_mean) / kenpom_std
base["kenpom_net_zscore_sq"] = base["kenpom_net_zscore"] ** 2
base["kenpom_net_zscore_cubed"] = base["kenpom_net_zscore"] ** 3
```

## Limitations and Future Work

### What the Model Can't Predict
1. **Narrative-driven mispricings**
   - Brand effects (Duke overpriced, Auburn underpriced despite similar KenPom)
   - Recency bias (hot teams, cold teams)
   - Conference bias

2. **Information asymmetry**
   - Injury news
   - Roster changes
   - Inside information

3. **Individual player behavior**
   - Research effort variance
   - Risk tolerance
   - Portfolio construction strategies

### The Model's Role
- Predicts **average market behavior** across all players
- Identifies **systematic patterns** (seed weighting, KenPom effects)
- Provides **baseline expectation** to spot deviations
- **Alpha comes from recognizing when market is irrational**

### Future Improvements
1. Test absolute KenPom rating as additional feature
2. Explore interaction terms (seed × z-score)
3. Consider ensemble methods with more data
4. Add conference strength features
5. Test time-series features (momentum, trends)

## Conclusion

The z-score squared + cubed model is **12% better at identifying underpriced elite teams**, which is the primary objective for a blind auction strategy. The model now captures magnitude of differences and handles top-heavy years better, while maintaining simplicity and avoiding overfitting.

**Production model updated**: December 30, 2024
