# Analysis Findings: 5000-Simulation Results (2017-2025)

**Date**: December 29, 2025
**Simulations**: 5000 per year
**Metric**: Expected payout (pool-size-invariant)

## Summary Statistics

- **Mean expected payout**: $88.06
- **Mean expected position**: 16.2 / ~45 entries
- **Mean P(1st)**: 13.5%
- **Mean P(top 3)**: 23.0%

## Performance by Year

| Year | Entries | Expected Payout | Position | P(1st) | P(Top 3) | Notes |
|------|---------|-----------------|----------|--------|----------|-------|
| 2017 | 23 | $26.23 | 5.5 | 0.3% | 13.3% | Moderate |
| 2018 | 42 | $53.72 | 14.3 | 2.4% | 13.0% | Moderate |
| **2019** | 37 | **$393.10** | **1.2** | **86.6%** | **97.4%** | **🏆 Exceptional** |
| **2021** | 41 | **$0.00** | **37.9** | **0.0%** | **0.0%** | **❌ Worst** |
| **2022** | 49 | **$223.73** | **4.0** | **18.3%** | **59.7%** | **🏆 Strong** |
| 2023 | 50 | $5.04 | 21.3 | 0.4% | 0.9% | Poor |
| 2024 | 52 | $2.43 | 18.1 | 0.0% | 0.0% | Poor |
| 2025 | 48 | $0.23 | 27.3 | 0.0% | 0.0% | Poor |

## Key Observations

### 🔴 Regression Alert: 2023-2025 Decline

**Concerning trend**: Performance has degraded significantly in recent years.
- 2022: $223.73 (excellent)
- 2023: $5.04 (poor)
- 2024: $2.43 (poor)
- 2025: $0.23 (poor)

**Potential causes to investigate**:
1. Market efficiency increased (other participants using similar models)
2. KenPom predictive power decreased
3. Algorithm assumptions no longer valid
4. Feature engineering needs updating
5. Auction dynamics changed

### High Variance Strategy

- Best year (2019): $393.10
- Worst year (2021): $0.00
- This is a **high-risk, high-reward** strategy
- Occasional home runs (2019, 2022) drive mean up
- But also complete misses (2021, 2024, 2025)

### No Clear Pool Size Pattern

- Small pools (23): $26
- Medium pools (37-42): $0 to $393 (huge variance)
- Large pools (48-52): Generally poor ($0-$5)
- Hypothesis: Algorithm may struggle as competition increases

## Action Items for Future Investigation

### Priority 1: Debug Recent Failures (2023-2025)
- [ ] Compare predicted vs actual tournament outcomes
- [ ] Analyze prediction accuracy (KenPom vs results)
- [ ] Review portfolio compositions
- [ ] Check if market became more efficient

### Priority 2: Understand 2019 Success
- [ ] What made 2019 different?
- [ ] Portfolio composition analysis
- [ ] Prediction accuracy vs actual results
- [ ] Market inefficiency identification

### Priority 3: Strategy Improvements
- [ ] Backtest alternative allocation strategies
- [ ] Test different concentration levels
- [ ] Evaluate risk management approaches
- [ ] Compare greedy vs waterfill allocation

## Notes

- Analysis based on 5000 simulations per year
- Using pool-size-invariant metric (expected payout in dollars)
- All portfolios: 10 teams, 100 points budget
- Greedy optimizer used for portfolio construction
