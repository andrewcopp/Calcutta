# Team Investability Theories & Future Analytics

## Core Concept: Investability Scores

Every team needs two key metrics to understand bidding behavior:

1. **Projected Points** - Expected points based on seed, historical performance, and other factors
2. **Projected Investment** - Expected investment based on seed and historical bidding patterns

The delta between actual and projected values reveals biases and opportunities.

---

## Theory 1: The Ugly Duckling Theory

### Hypothesis
Within each seed tier (especially 2-4 seeds), there are significant variations in "investability" - some teams attract disproportionate investment while others are overlooked, despite similar expected performance.

### Observation
- ROI data shows 2, 3, and 4 seeds appear properly valued on average (ROI ≈ 1.0)
- However, anecdotal experience suggests high variance within these seed tiers
- Every year there's a "hot" 2-seed everyone wants and a "cold" 2-seed nobody wants

### Statistical Validation Needed
- Calculate **standard deviation of investment** within each seed tier
- High standard deviation = high variance in investability
- Compare to standard deviation of actual points scored
- If investment variance >> performance variance, theory is validated

### Implementation Plan

#### Phase 1: Variance Analysis
```sql
-- Calculate investment variance by seed
SELECT 
  tt.seed,
  STDDEV(cet.bid) as investment_stddev,
  STDDEV(cpt.actual_points) as points_stddev,
  AVG(cet.bid) as avg_investment,
  AVG(cpt.actual_points) as avg_points,
  (STDDEV(cet.bid) / NULLIF(AVG(cet.bid), 0)) as investment_cv,
  (STDDEV(cpt.actual_points) / NULLIF(AVG(cpt.actual_points), 0)) as points_cv
FROM tournament_teams tt
JOIN calcutta_entry_teams cet ON cet.team_id = tt.id
LEFT JOIN calcutta_portfolio_teams cpt ON cpt.team_id = tt.id
GROUP BY tt.seed
ORDER BY tt.seed;
```

#### Phase 2: Outlier Detection
For each seed tier, identify:
- **Hot Teams** (investment > mean + 1.5 * stddev)
- **Cold Teams** (investment < mean - 1.5 * stddev)
- **Neutral Teams** (within 1 stddev of mean)

#### Phase 3: Investability Score
Create a normalized investability score:
```
Investability Score = (Actual Investment - Expected Investment) / Investment StdDev
```
- Score > 1.5: "Hot" team (over-invested)
- Score < -1.5: "Cold" team (under-invested)
- Score ≈ 0: Properly valued

### Expected Insights
- Identify which team characteristics drive "hotness" (brand, conference, recent success, etc.)
- Find value in "cold" teams that perform as well as "hot" teams
- Predict which teams will be over/under-valued in future years

---

## Theory 2: The Conference Bias Theory

### Hypothesis
Teams from certain conferences (ACC, Big Ten, SEC) receive disproportionate investment compared to their actual performance, while teams from mid-major conferences are systematically undervalued.

### Validation Approach
- Add conference data to team records
- Calculate average ROI by conference
- Control for seed (compare same-seed teams across conferences)
- Identify "premium conferences" (negative ROI) and "value conferences" (positive ROI)

### Implementation
- Add `conference` field to `tournament_teams` or `schools` table
- Create conference analytics endpoint
- Visualize ROI by conference with seed normalization

### Expected Insights
- Quantify the "blue blood" premium
- Identify mid-major conferences that consistently over-perform
- Find arbitrage opportunities in conference perception gaps

---

## Theory 3: The Recency Bias Theory

### Hypothesis
Teams that won their conference tournament or had a strong finish to the regular season receive inflated investment, regardless of their overall season performance or seed.

### Validation Approach
- Add conference tournament winner flag to team data
- Add "last 10 games" record or momentum indicator
- Compare investment for conference champs vs non-champs at same seed
- Measure if "hot team" premium is justified by performance

### Data Needed
- Conference tournament results
- End-of-season momentum metrics
- Regular season vs tournament performance

### Expected Insights
- Quantify the "hot team" premium
- Determine if conference tournament success predicts NCAA success
- Identify when recency bias creates value opportunities

---

## Theory 4: The Bracket Position Theory

### Hypothesis
Teams in certain bracket positions (e.g., playing in popular cities, favorable time slots, or against perceived "easy" opponents) receive higher investment independent of seed.

### Validation Approach
- Add first-round opponent data
- Add game location/region data
- Compare investment for teams with "favorable" vs "unfavorable" first-round matchups
- Control for seed

### Data Needed
- First-round matchup data
- Game location/venue
- Historical performance vs specific opponent types

### Expected Insights
- Quantify the "easy path" premium
- Identify if bracket position matters for actual performance
- Find value in teams with perceived "tough" draws

---

## Theory 5: The KenPom Divergence Theory

### Hypothesis
When a team's KenPom ranking significantly differs from their seed, the group either over-corrects or under-corrects, creating arbitrage opportunities.

### Validation Approach
- Integrate KenPom efficiency ratings
- Calculate divergence: `KenPom Rank - (Seed * 4)`
- Measure investment correlation with KenPom vs seed
- Identify if group values analytics or traditional seeding

### Data Needed
- KenPom ratings (efficiency, offensive/defensive ratings)
- Historical KenPom data for past tournaments

### Expected Insights
- Determine if group is "analytics-aware" or "seed-focused"
- Find teams where analytics suggest value
- Identify when KenPom divergence predicts upsets

---

## Theory 6: The Familiarity Bias Theory

### Hypothesis
Teams that appear frequently in the tournament receive more consistent (and possibly inflated) investment due to familiarity, while first-time or rare participants are undervalued.

### Validation Approach
- Calculate team "appearance frequency" in dataset
- Compare investment for frequent vs rare participants at same seed
- Measure if familiarity correlates with over-investment

### Implementation
- Already have `appearances` count in team analytics
- Segment teams by appearance frequency
- Calculate ROI by appearance tier

### Expected Insights
- Quantify the "brand recognition" premium
- Find value in unfamiliar teams
- Identify if experience matters for actual performance

---

## Theory 7: The Upset Potential Theory

### Hypothesis
Teams perceived as "upset picks" (typically 5-12 seeds) receive inflated investment from participants seeking high-risk/high-reward portfolios.

### Validation Approach
- Analyze investment patterns in "upset-prone" seed ranges
- Compare investment volatility across seed tiers
- Identify if "upset hunting" is rational or emotional

### Expected Insights
- Quantify the "upset premium"
- Determine optimal upset-hunting strategy
- Identify which seeds offer true upset value

---

## Theory 8: The Hometown Hero Theory

### Hypothesis
If any participants have alumni connections to specific schools, those teams receive inflated investment regardless of seed or performance potential.

### Validation Approach
- Add participant-school affiliation data (if available)
- Compare investment in "connected" vs "unconnected" teams
- Measure the "alma mater premium"

### Data Needed
- Participant profiles with school affiliations
- Historical bidding data by participant

### Expected Insights
- Quantify emotional bidding
- Identify when personal connections override value
- Find opportunities when others overpay for "their" team

---

## Implementation Roadmap

### Phase 1: Standard Deviation Analysis (Immediate)
- Add variance/stddev calculations to existing analytics
- Create "Investability Variance" view showing stddev by seed
- Validate Ugly Duckling Theory with current data

### Phase 2: Projected Scores Admin Tools
Create two new admin tools:

#### A. Projected Points Calculator
- Input: Team, Seed, Conference, KenPom (optional)
- Algorithm: Regression model based on historical seed performance
- Output: Expected points with confidence interval
- Use: Compare actual points to projection to find over/under-performers

#### B. Projected Investment Calculator
- Input: Team, Seed, Conference, Brand Metrics
- Algorithm: Regression model based on historical bidding patterns
- Output: Expected investment with confidence interval
- Use: Identify "hot" and "cold" teams before auction

### Phase 3: Outlier Detection Dashboard
- Visual identification of hot/cold teams at each seed
- Real-time investability scores during data entry
- Historical outlier performance tracking

### Phase 4: External Data Integration
- KenPom API integration
- Conference tournament results
- Advanced metrics (tempo, efficiency, etc.)

### Phase 5: Predictive Modeling
- Machine learning model for investment prediction
- Feature importance analysis (what makes a team "hot"?)
- Pre-auction value identification tool

---

## Data Requirements for Full Implementation

### Immediate (Already Have)
- ✅ Seed
- ✅ Points scored
- ✅ Investment amounts
- ✅ Team appearances

### Short-term (Easy to Add)
- Conference affiliation
- Conference tournament winner flag
- Regular season record
- Tournament region/location

### Medium-term (Requires Integration)
- KenPom ratings
- Strength of schedule
- Recent performance trends
- First-round opponent

### Long-term (Advanced)
- Participant profiles/affiliations
- Real-time betting odds
- Social media sentiment
- Historical upset rates by seed matchup

---

## Success Metrics

### For Ugly Duckling Theory
- Investment stddev > 1.5x points stddev within seed tier
- Ability to predict "hot" vs "cold" teams with >70% accuracy
- Identification of 2-3 undervalued teams per seed tier

### For Other Theories
- Statistical significance (p < 0.05) for each bias
- Quantified premium/discount for each factor
- Actionable insights that improve ROI by >10%

---

## Notes for Future Development

### Admin Tool Ideas
1. **Pre-Auction Value Scanner** - Before auction, identify likely undervalued teams
2. **Live Auction Tracker** - During auction, flag when teams exceed projected investment
3. **Post-Auction Analysis** - After auction, show which participants got best/worst value
4. **Historical Bias Report** - Annual report on group's bidding biases and patterns

### Visualization Ideas
1. **Investability Heatmap** - Visual grid of seed vs investability score
2. **Outlier Scatter Plot** - Investment vs Points with outliers highlighted
3. **Bias Timeline** - How biases change year-over-year
4. **Value Opportunity Board** - Real-time list of undervalued teams

### Research Questions
1. Do biases compound (e.g., ACC + Conference Champ + High Seed)?
2. Can we predict which participant will overpay for which team?
3. Does group learning reduce biases over time?
4. What's the optimal contrarian strategy?
