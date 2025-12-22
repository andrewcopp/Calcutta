# Sandbox Ideas Backlog

This file is a lightweight backlog for sandbox research ideas that are not yet “stable” enough to promote into long-lived docs.

Related docs:
- `docs/data_science_sandbox.md`
- `docs/investability_theories.md`

## How to use this file

- Add ideas here first.
- When an idea becomes an implemented metric/model feature, move it into:
  - `docs/investability_theories.md` (behavioral / bidding theories), or
  - `docs/tournament_modeling.md` (returns modeling), or
  - code + an experiment write-up in `backend/cmd/sandbox/README.md` (how to run it)

## Idea template

### [IDEA] <short name>

- **Problem**:
- **Hypothesis**:
- **Why it might be true (behavioral story)**:
- **What we expect to observe in historical data**:
- **Proposed features**:
- **Proposed evaluation**:
- **Confounders / failure modes**:
- **Decision rule**:
  - If ..., we keep it.
  - If ..., we drop it.

---

## Backlog

### [IDEA] The Overlooked #3 (within-seed pod mispricing)

- **Problem**:
  - A seed-only investment model assumes all teams of the same seed are priced similarly, but the auction is driven by *relative narratives* within “seed pods” (e.g., the four 1-seeds).
- **Hypothesis**:
  - In a 4-team seed pod (especially 1-seeds), the team perceived as “2nd best” and the team perceived as “sneaky value” can both attract extra investment, making the “3rd best” team *systematically under-invested*.
- **Why it might be true (behavioral story)**:
  - People justify overbidding the top team (“they’re that good”) and also chase the “contrarian 1-seed value” narrative (“everyone will overlook them”). This can crowd out the middle option.
- **What we expect to observe in historical data**:
  - Within a seed pod, after controlling for team strength, investment residuals show a dip around the 3rd-ranked team.
  - This effect may be stronger when the top team is a strong brand.
- **Proposed features**:
  - `pod_rank_by_strength` within seed pod (1-4 for 1-seeds, 1-4 for 2-seeds, etc.)
  - Strength proxy candidates (need at least one):
    - KenPom rank / AdjEM rank
    - Vegas title odds / win probabilities (if added later)
    - Simple proxy: prior-year tournament points / appearances / “brand” score (if you build one)
  - Interaction candidates:
    - `is_blue_blood` / `brand_score`
    - `conference`
- **Proposed evaluation**:
  - Train investment model with seed + pod rank features and measure:
    - MAE/RMSE on investment share
    - Within-pod rank correlation
    - Residual plot by `pod_rank_by_strength`
  - Compare to baseline (seed-only).
- **Confounders / failure modes**:
  - Small sample sizes (only 4 teams per pod per year).
  - “Strength” proxy might already be what the group uses (then there’s no behavioral gap).
  - Bid shares may be influenced by a small number of big bidders.
- **Decision rule**:
  - If pod-rank features improve within-pod ordering and reduce residual variance meaningfully, keep.
  - If effect is not stable year-over-year, treat as anecdotal / non-generalizable.

---

### [IDEA] Region of Death (region strength affects investment behavior)

- **Problem**:
  - Predicted investment is partly driven by perceived path difficulty (“that region is brutal”), which is not captured by seed-only models.
- **Hypothesis**:
  - When a region is unusually strong (high average strength among top seeds), teams in that region—especially the 1-seed—receive *less* investment than a similar-strength team in an easier region.
- **Why it might be true (behavioral story)**:
  - People overweight path narratives (“too many landmines”) and underweight that elite teams can still be great value depending on how others react.
- **What we expect to observe in historical data**:
  - Conditional on team strength, higher region-strength predicts lower investment share.
  - Effect is stronger for the 1- and 2-seed than for lower seeds.
- **Proposed features**:
  - `region_strength_top4_avg` (avg strength of top 4 seeds in region)
  - `region_strength_top8_avg` (avg strength of top 8 seeds in region)
  - `region_strength_zscore` within year (to compare “region of death” vs other regions)
  - Strength proxy candidates:
    - KenPom rank / AdjEM
    - Seed-adjusted strength score (placeholder until external metrics exist)
- **Proposed evaluation**:
  - Backtest investment model:
    - baseline: seed-only
    - + team strength
    - + region strength
  - Check:
    - incremental MAE improvement
    - sign + stability of `region_strength_*` coefficients
    - within-year region comparisons
- **Confounders / failure modes**:
  - Region assignment may correlate with team brand/conference, which can dominate.
  - Without an external strength metric, region-strength is circular if derived only from seeds.
- **Decision rule**:
  - If region strength meaningfully improves prediction and is stable across years, keep.
  - If it only “works” with hindsight or depends on post-tournament outcomes, drop (leakage).

---

### [IDEA] “Same seed, different team” requires external features

- **Problem**:
  - Multiple investment ideas require differentiating teams within the same seed.
- **Hypothesis**:
  - Adding a small set of external features (conference + record + KenPom) will unlock both better returns prediction and better investment prediction.
- **Proposed data additions**:
  - `conference`
  - `wins`, `losses` (or win%)
  - KenPom: rank + AdjEM + offense/defense efficiency
- **Decision rule**:
  - If adding *any* of these materially improves within-pod ordering (investment) and/or calibration (returns), prioritize the ingestion pipeline.

---

## Backlog (additional game theory + Calcutta-specific hypotheses)

### [IDEA] Top-heavy payout bias (bidding favors championship equity)

- **Problem**:
  - If payouts are top-heavy (winner-take-most or steep top-N splits), then “expected points” is not the true objective; championship probability matters disproportionately.
- **Hypothesis**:
  - The community systematically overbids teams with high title equity (often 1-seeds / blue bloods) relative to their expected points, because bidders optimize for “win the whole pool” outcomes.
- **Why it might be true (behavioral story)**:
  - People buy lottery tickets: they’d rather maximize the chance of finishing 1st than maximize EV.
- **What we expect to observe in historical data**:
  - Investment share correlates more strongly with deep-run outcomes (Final Four / champion) than with total points, especially at the top of the seed curve.
  - Conditional on expected points, teams with higher “tail outcomes” attract more investment.
- **Proposed features**:
  - Returns-side proxies for championship equity:
    - `p_final_four`, `p_title` (once you have a probabilistic returns model)
    - If not available yet: `expected_points_in_late_rounds` vs `expected_points_total`
  - `seed`
  - `brand_score` (optional)
- **Proposed evaluation**:
  - Compare investment model variants:
    - baseline: seed-only
    - + expected points (single value)
    - + “late-round equity” features (title / Final Four probabilities or proxies)
  - Evaluate within-seed-pod ordering + MAE.
- **Confounders / failure modes**:
  - Title equity is hard to estimate without external strength features.
  - Small sample size for champions; use probabilities not realized outcomes.
- **Decision rule**:
  - If late-round equity materially improves investment prediction and is stable, keep.
  - If it only fits when using realized outcomes, drop (leakage).

---

### [IDEA] Portfolio correlation premium (people pay to diversify paths)

- **Problem**:
  - In a blind auction, individuals often build multi-team portfolios. Correlated outcomes reduce “win-the-pool” probability even if EV is similar.
- **Hypothesis**:
  - The community overbids for teams that *diversify well* with common portfolios (e.g., strong teams in different regions), and underbids for teams that feel redundant (same region / same path).
- **Why it might be true (behavioral story)**:
  - People intuitively diversify, even when they don’t explicitly model payout distributions.
- **What we expect to observe in historical data**:
  - Conditional on seed/strength, teams in the most “popular” regions are underbid.
  - Teams that are natural complements to chalky favorites get bid up.
- **Proposed features**:
  - `region`
  - `region_strength_*` (from Region of Death idea)
  - `conference` (optional; for perceived correlation)
  - “path-correlation” proxies:
    - same region as top overall 1-seed
    - expected probability of meeting a top team by round (once you have bracket simulation)
- **Proposed evaluation**:
  - Start simple: within-year, compare investment residuals by region after controlling for seed + strength.
  - Later: estimate portfolio-level outcomes by simulating brackets and measuring if actual bidding patterns improve “diversification” compared to random.
- **Confounders / failure modes**:
  - Requires knowing individual portfolios (if available) to test directly.
  - Region effects can be dominated by brand.
- **Decision rule**:
  - If region/path features explain residual investment variance beyond seed+strength, keep.
  - If effects are inconsistent or tiny, drop.

---

### [IDEA] “Conference cannibalization” discount

- **Problem**:
  - People believe same-conference teams are more likely to eliminate each other (or are “known quantities”), which can change perceived upside.
- **Hypothesis**:
  - When a conference has many strong bids in the field, teams from that conference receive a pricing discount (or premium) relative to comparable teams.
- **Why it might be true (behavioral story)**:
  - Narratives like “the conference is overrated” or “they’ll beat each other up” are easy to overweight.
- **What we expect to observe in historical data**:
  - After controlling for seed/strength, investment residuals vary systematically by conference.
  - The sign may differ by era (e.g., SEC inflation in some years).
- **Proposed features**:
  - `conference`
  - `conference_strength_in_field`:
    - number of bids from conference
    - average strength of conference teams (KenPom avg, if available)
- **Proposed evaluation**:
  - Regression / hierarchical model with conference random effects.
  - Backtest by year split.
- **Confounders / failure modes**:
  - Conference proxies brand; need to control for team brand.
  - External data needed.
- **Decision rule**:
  - If conference effects remain after controlling for strength/brand, keep.
  - If conference only acts as a strength proxy, fold into strength model instead.

---

### [IDEA] “Bubble fraud” discount (record/seed disagreement)

- **Problem**:
  - Some teams have a strong seed but an unconvincing record / weak schedule, creating a “fraud” narrative.
- **Hypothesis**:
  - Conditional on seed, teams with worse record or weak schedule are underbid (investment discount), even when their underlying strength suggests the seed is deserved.
- **Why it might be true (behavioral story)**:
  - People anchor to a simple narrative like “they didn’t play anyone.”
- **What we expect to observe in historical data**:
  - Within-seed, record/SOS features predict investment residuals.
- **Proposed features**:
  - `wins`, `losses`, `win_pct`
  - Strength of schedule (if available later)
  - `seed_minus_strength_rank` (KenPom/other)
- **Proposed evaluation**:
  - Compare within-seed investment residual variance explained by record/SOS.
- **Confounders / failure modes**:
  - Record is confounded by conference strength.
  - Data availability.
- **Decision rule**:
  - Keep if record/SOS improves within-seed ordering without leaking post-tournament info.

---

### [IDEA] Brand tax vs “new money” discount (familiarity premium)

- **Problem**:
  - Seed-only ignores that some programs are household names and some are unfamiliar.
- **Hypothesis**:
  - Brand-name programs are consistently overbid; unfamiliar programs are underbid, even at the same seed.
- **Why it might be true (behavioral story)**:
  - People pay for the comfort of recognition.
- **What we expect to observe in historical data**:
  - Positive investment residuals for high-appearance teams.
- **Proposed features**:
  - `appearances` (already noted elsewhere in the project)
  - `recent_appearances_last_N_years` (if you compute it)
  - `brand_score` built from appearances + historical bidding residuals (careful about leakage)
- **Proposed evaluation**:
  - Backtest investment prediction with appearances/brand features added.
- **Confounders / failure modes**:
  - Brand correlates with team strength.
- **Decision rule**:
  - Keep if brand features still matter after controlling for strength.

---

### [IDEA] Winner’s curse / overpayment factor (sealed-bid inefficiency)

- **Problem**:
  - In blind auctions, the highest bidder tends to overpay relative to the “true” common value.
- **Hypothesis**:
  - The distribution of bids on a team implies a predictable overpayment gap between the winning bid and the median/mean bid, and this gap varies by seed/brand.
- **Why it might be true (behavioral story)**:
  - Optimistic bidders systematically win.
- **What we expect to observe in historical data**:
  - If you have per-bid data: `winning_bid` > `second_price` (or > median) by a systematic margin.
  - The margin is larger for teams with more narrative uncertainty (mid seeds, trendy upset picks).
- **Proposed features**:
  - If bid distribution is available:
    - `num_bids`, `bid_stddev`, `bid_cv`, `winning_bid_minus_median`
  - `seed`, `brand_score`
- **Proposed evaluation**:
  - Estimate the “overpayment factor” per seed tier and test stability year-over-year.
  - Use it to adjust predicted investment or expected ROI for *your* bids.
- **Confounders / failure modes**:
  - If only total investment is stored (not per-bid distribution), you can’t test this directly.
- **Decision rule**:
  - If bid dispersion exists and predicts overpayment reliably, keep and operationalize.

---

### [IDEA] “Upset chic” inflation (popular upset seeds are overbid)

- **Problem**:
  - Certain seeds (e.g., 10-12) have an “upset pick” narrative that may inflate bids beyond EV.
- **Hypothesis**:
  - Conditional on expected points, teams in upset-favorite seed ranges are overbid, especially when external metrics suggest upset potential.
- **Why it might be true (behavioral story)**:
  - People chase the glory of calling an upset.
- **What we expect to observe in historical data**:
  - Positive investment residuals concentrated in a specific seed band.
- **Proposed features**:
  - `seed`
  - “upset potential” proxy:
    - KenPom / efficiency gap between matchup opponents (requires matchup data)
    - spread (if added later)
- **Proposed evaluation**:
  - Residual analysis by seed band + (later) matchup-driven upset probability.
- **Confounders / failure modes**:
  - Needs matchup data for the cleanest test.
- **Decision rule**:
  - Keep if the pattern persists across years and isn’t purely one-off media hype.
