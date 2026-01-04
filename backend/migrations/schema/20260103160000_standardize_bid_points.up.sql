-- Standardize bid columns to canonical bid_points naming.

ALTER TABLE lab_bronze.entry_bids
    RENAME COLUMN bid_amount_points TO bid_points;

ALTER TABLE lab_gold.recommended_entry_bids
    RENAME COLUMN recommended_bid_points TO bid_points;
