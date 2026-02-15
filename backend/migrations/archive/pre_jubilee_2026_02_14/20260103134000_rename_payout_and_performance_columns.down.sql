-- Rollback payout/performance column renames.

ALTER TABLE IF EXISTS derived.entry_simulation_outcomes
    RENAME COLUMN payout_cents TO payout_points;

ALTER TABLE IF EXISTS derived.entry_performance
    RENAME COLUMN mean_normalized_payout TO mean_payout;

ALTER TABLE IF EXISTS derived.entry_performance
    RENAME COLUMN median_normalized_payout TO median_payout;
