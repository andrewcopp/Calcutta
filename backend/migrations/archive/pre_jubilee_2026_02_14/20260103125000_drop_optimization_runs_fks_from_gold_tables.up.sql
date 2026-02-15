-- Drop remaining foreign keys that require lab_gold.* tables to reference
-- lab_gold.optimization_runs. We are migrating to lab_gold.strategy_generation_runs.

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_constraint c
        WHERE c.conrelid = 'lab_gold.recommended_entry_bids'::regclass
          AND c.contype = 'f'
          AND c.conname = 'gold_recommended_entry_bids_run_id_fkey'
    ) THEN
        ALTER TABLE lab_gold.recommended_entry_bids
            DROP CONSTRAINT gold_recommended_entry_bids_run_id_fkey;
    END IF;

    IF EXISTS (
        SELECT 1
        FROM pg_constraint c
        WHERE c.conrelid = 'lab_gold.detailed_investment_report'::regclass
          AND c.contype = 'f'
          AND c.conname = 'gold_detailed_investment_report_run_id_fkey'
    ) THEN
        ALTER TABLE lab_gold.detailed_investment_report
            DROP CONSTRAINT gold_detailed_investment_report_run_id_fkey;
    END IF;
END $$;
