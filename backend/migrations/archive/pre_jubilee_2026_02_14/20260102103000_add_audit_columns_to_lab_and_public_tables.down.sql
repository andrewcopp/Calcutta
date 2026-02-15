-- Rollback audit columns added to lab-tier tables and public tables.

-- Remove triggers created by this migration (do not touch core).
DO $$
DECLARE
    r record;
BEGIN
    FOR r IN (
        SELECT
            n.nspname AS schema_name,
            c.relname AS table_name
        FROM pg_class c
        JOIN pg_namespace n ON n.oid = c.relnamespace
        JOIN pg_attribute a ON a.attrelid = c.oid
        WHERE n.nspname IN ('bronze', 'silver', 'gold', 'public')
          AND c.relkind IN ('r', 'p')
          AND a.attname = 'updated_at'
          AND a.attnum > 0
          AND NOT a.attisdropped
    ) LOOP
        EXECUTE format(
            'DROP TRIGGER IF EXISTS %I ON %I.%I',
            'set_updated_at',
            r.schema_name,
            r.table_name
        );
    END LOOP;
END;
$$;

-- BRONZE
ALTER TABLE bronze.tournaments DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE bronze.tournaments DROP COLUMN IF EXISTS updated_at;

ALTER TABLE bronze.teams DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE bronze.teams DROP COLUMN IF EXISTS updated_at;

ALTER TABLE bronze.calcuttas DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE bronze.calcuttas DROP COLUMN IF EXISTS updated_at;

ALTER TABLE bronze.entry_bids DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE bronze.entry_bids DROP COLUMN IF EXISTS updated_at;

ALTER TABLE bronze.payouts DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE bronze.payouts DROP COLUMN IF EXISTS updated_at;

-- SILVER
ALTER TABLE silver.predicted_game_outcomes DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE silver.predicted_game_outcomes DROP COLUMN IF EXISTS updated_at;

ALTER TABLE silver.simulated_tournaments DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE silver.simulated_tournaments DROP COLUMN IF EXISTS updated_at;

ALTER TABLE silver.predicted_market_share DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE silver.predicted_market_share DROP COLUMN IF EXISTS updated_at;

-- GOLD
ALTER TABLE gold.optimization_runs DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE gold.optimization_runs DROP COLUMN IF EXISTS updated_at;

ALTER TABLE gold.recommended_entry_bids DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE gold.recommended_entry_bids DROP COLUMN IF EXISTS updated_at;

ALTER TABLE gold.entry_simulation_outcomes DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE gold.entry_simulation_outcomes DROP COLUMN IF EXISTS updated_at;

ALTER TABLE gold.entry_performance DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE gold.entry_performance DROP COLUMN IF EXISTS updated_at;

ALTER TABLE gold.detailed_investment_report DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE gold.detailed_investment_report DROP COLUMN IF EXISTS updated_at;

-- PUBLIC
ALTER TABLE public.api_keys DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE public.api_keys DROP COLUMN IF EXISTS updated_at;

ALTER TABLE public.label_permissions DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE public.label_permissions DROP COLUMN IF EXISTS updated_at;

ALTER TABLE public.auth_sessions DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE public.grants DROP COLUMN IF EXISTS deleted_at;
