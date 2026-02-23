# UX Reviewer Memory - Calcutta Project

## Project Context
March Madness investment pool application with ML-driven entry optimization pipeline.

## Lab Pipeline Architecture (Critical for UX Understanding)

### Data Flow
1. **Investment Model** predicts market behavior (what OTHER participants will bid)
2. **Optimizer** (MINLP) generates OUR optimal entry based on those predictions
3. **Evaluator** runs simulations to measure performance

### Key Data Distinctions (MUST preserve in UI)
- **BidPoints**: What WE bid on each team (our optimized entry)
- **NaivePoints**: Rational allocation proportional to expected value (KenPom-adjusted)
- **ExpectedROI**: Expected return BEFORE our bid affects market
- **OurROI**: Expected return AFTER our bid is factored in
- **EdgePercent**: (NaivePoints - BidPoints) / NaivePoints * 100
  - Positive = undervalued (we bid less than rational would)
  - Negative = overvalued (we bid more than rational would)

### Current UI Confusion
EntryDetailPage.tsx conflates:
- Market predictions (what model thinks others will bid)
- Our optimized bids (what we should bid to maximize ROI)
- Rational baseline (value-proportional allocation)

The current "Investment" column shows BidPoints, but the UX doesn't clarify this is OUR bid, not market prediction.

## Design System
- Tailwind CSS with custom color palette
- Green = positive/opportunity (100/50 shades for different thresholds)
- Red = negative/avoid (100/50 shades)
- Blue = interactive/selected (50 bg, 600 text)
- Gray scale for neutral states
- Tabular numbers for alignment
- Sort headers with triangle indicators
- Progress bars for pipeline stages (registered → entries → evaluated)

## Component Patterns
- Card component for content sections
- StatusBadge for pipeline states
- PipelineProgressBar shows 3-stage flow
- Sortable table headers with visual indicators
- Filter toggles (checkbox with count)
- Breadcrumb navigation (Lab → Model → Entry)
