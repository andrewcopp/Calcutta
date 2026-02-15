-- Restore foreign keys from lab_gold.* tables back to lab_gold.optimization_runs.

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint c
        WHERE c.conrelid = 'lab_gold.recommended_entry_bids'::regclass
          AND c.contype = 'f'
          AND c.conname = 'gold_recommended_entry_bids_run_id_fkey'
    ) THEN
        ALTER TABLE lab_gold.recommended_entry_bids
            ADD CONSTRAINT gold_recommended_entry_bids_run_id_fkey
            FOREIGN KEY (run_id)
            REFERENCES lab_gold.optimization_runs(run_id)
            ON DELETE CASCADE;
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint c
        WHERE c.conrelid = 'lab_gold.detailed_investment_report'::regclass
          AND c.contype = 'f'
          AND c.conname = 'gold_detailed_investment_report_run_id_fkey'
    ) THEN
        ALTER TABLE lab_gold.detailed_investment_report
            ADD CONSTRAINT gold_detailed_investment_report_run_id_fkey
            FOREIGN KEY (run_id)
            REFERENCES lab_gold.optimization_runs(run_id)
            ON DELETE CASCADE;
    END IF;
END $$;
