# Analytics Future Enhancements

## Team Analytics - Seed Normalization

### Overview
The current team analytics shows raw totals and averages, but doesn't account for the fact that teams with better seeds naturally score more points and receive higher investment. To truly identify biases in bidding behavior, we need to normalize team performance by seed.

### Approach

#### 1. Expected Value Calculation
For each seed position (1-16), calculate:
- **Expected Points**: Average points scored by all teams at that seed
- **Expected Investment**: Average investment for all teams at that seed

#### 2. Performance Metrics
For each team appearance, calculate:
- **Points Differential**: Actual points - Expected points for that seed
- **Investment Differential**: Actual investment - Expected investment for that seed
- **Points Ratio**: Actual points / Expected points (1.0 = average, >1.0 = over-performed)
- **Investment Ratio**: Actual investment / Expected investment (1.0 = average, >1.0 = over-valued)

#### 3. Aggregate Team Metrics
Across all appearances for a team:
- **Average Points Differential**: Mean of all points differentials
- **Average Investment Differential**: Mean of all investment differentials
- **Consistency Score**: Standard deviation of performance ratios
- **Bias Score**: Investment ratio / Points ratio
  - Score > 1.0: Team is over-valued (people pay more than performance warrants)
  - Score < 1.0: Team is under-valued (bargain pick)
  - Score ≈ 1.0: Fairly valued

### Example Use Cases

#### Finding Over-Valued Teams (Bias Detection)
Teams with high bias scores indicate the group has an emotional attachment or brand preference:
- Michigan as a #5 seed might get #2 seed investment
- Duke might consistently be over-valued regardless of seed

#### Finding Under-Valued Teams (Value Picks)
Teams with low bias scores are potential bargains:
- Mid-major teams that consistently over-perform their seed
- Teams from less popular conferences

#### Consistency Analysis
Teams with low consistency scores are reliable performers:
- Good for risk-averse strategies
- Predictable ROI

### Implementation Plan

1. **Backend Changes**
   - Add new analytics endpoint: `/api/analytics/teams/normalized`
   - Create queries to calculate expected values by seed
   - Implement differential and ratio calculations
   - Add filtering options (minimum appearances, specific tournaments, etc.)

2. **Frontend Changes**
   - Add new tab or section for normalized team analytics
   - Create visualizations:
     - Scatter plot: Investment Ratio vs Points Ratio (identify quadrants)
     - Table: Top over-valued and under-valued teams
     - Trend analysis: How team valuation changes over time
   - Add filters for minimum appearances, seed ranges, etc.

3. **Database Considerations**
   - Current schema supports this analysis
   - May want to add materialized views for performance
   - Consider caching expected values per seed

### Additional Data Integration

#### KenPom Ratings
- Integrate KenPom efficiency ratings
- Compare investment to KenPom ranking vs seed
- Identify when group values analytics over traditional seeding

#### Conference Tournament Winners
- Flag teams that won their conference tournament
- Analyze if "hot team" bias affects investment
- Compare performance of auto-bids vs at-large teams

#### Regular Season Record
- Add wins/losses to team data
- Analyze if record affects investment independent of seed
- Identify if group values "eye test" over metrics

### SQL Query Example

```sql
-- Calculate expected values by seed
WITH seed_expectations AS (
  SELECT 
    tt.seed,
    AVG(cpt.points_earned) as expected_points,
    AVG(cet.bid) as expected_investment
  FROM tournament_teams tt
  LEFT JOIN calcutta_entry_teams cet ON cet.team_id = tt.id
  LEFT JOIN calcutta_portfolio_teams cpt ON cpt.team_id = tt.id
  WHERE tt.deleted_at IS NULL
  GROUP BY tt.seed
),
-- Calculate team performance vs expectations
team_performance AS (
  SELECT 
    s.id as school_id,
    s.name as school_name,
    tt.seed,
    cpt.points_earned,
    cet.bid,
    se.expected_points,
    se.expected_investment,
    (cpt.points_earned - se.expected_points) as points_diff,
    (cet.bid - se.expected_investment) as investment_diff,
    (cpt.points_earned / NULLIF(se.expected_points, 0)) as points_ratio,
    (cet.bid / NULLIF(se.expected_investment, 0)) as investment_ratio
  FROM schools s
  JOIN tournament_teams tt ON tt.school_id = s.id
  JOIN seed_expectations se ON se.seed = tt.seed
  LEFT JOIN calcutta_entry_teams cet ON cet.team_id = tt.id
  LEFT JOIN calcutta_portfolio_teams cpt ON cpt.team_id = tt.id
  WHERE tt.deleted_at IS NULL
)
-- Aggregate by team
SELECT 
  school_id,
  school_name,
  COUNT(*) as appearances,
  AVG(points_diff) as avg_points_diff,
  AVG(investment_diff) as avg_investment_diff,
  AVG(points_ratio) as avg_points_ratio,
  AVG(investment_ratio) as avg_investment_ratio,
  (AVG(investment_ratio) / NULLIF(AVG(points_ratio), 0)) as bias_score
FROM team_performance
GROUP BY school_id, school_name
HAVING COUNT(*) >= 3  -- Minimum appearances for statistical significance
ORDER BY bias_score DESC;
```

### Visualization Ideas

1. **Bias Quadrant Chart**
   - X-axis: Points Ratio (performance)
   - Y-axis: Investment Ratio (valuation)
   - Quadrants:
     - Top-Right: Over-valued, Over-performers (justified premium)
     - Top-Left: Over-valued, Under-performers (bias/bad picks)
     - Bottom-Right: Under-valued, Over-performers (value picks!)
     - Bottom-Left: Under-valued, Under-performers (correctly valued)

2. **Team Timeline**
   - Show how a specific team's bias score changes over years
   - Identify if biases increase/decrease with experience

3. **Seed-Specific Analysis**
   - Compare teams only within same seed ranges
   - "Best value #8 seeds" or "Most over-valued #1 seeds"
