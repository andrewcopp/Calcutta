-- Rename payout and performance columns to correct units/meaning.
-- payout is real money => cents.
-- mean/median are normalized payout (unitless).

ALTER TABLE IF EXISTS derived.entry_simulation_outcomes
    RENAME COLUMN payout_points TO payout_cents;

ALTER TABLE IF EXISTS derived.entry_performance
    RENAME COLUMN mean_payout TO mean_normalized_payout;

ALTER TABLE IF EXISTS derived.entry_performance
    RENAME COLUMN median_payout TO median_normalized_payout;
