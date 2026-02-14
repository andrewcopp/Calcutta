-- Revert bid_points standardization.

ALTER TABLE lab_bronze.entry_bids
    RENAME COLUMN bid_points TO bid_amount_points;

ALTER TABLE lab_gold.recommended_entry_bids
    RENAME COLUMN bid_points TO recommended_bid_points;
