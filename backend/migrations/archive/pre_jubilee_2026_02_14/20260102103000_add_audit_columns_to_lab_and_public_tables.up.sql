-- Add audit columns (updated_at, deleted_at) to lab-tier tables and remaining public tables.

-- BRONZE
ALTER TABLE bronze.tournaments ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
ALTER TABLE bronze.tournaments ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

ALTER TABLE bronze.teams ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
ALTER TABLE bronze.teams ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

ALTER TABLE bronze.calcuttas ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
ALTER TABLE bronze.calcuttas ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

ALTER TABLE bronze.entry_bids ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
ALTER TABLE bronze.entry_bids ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

ALTER TABLE bronze.payouts ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
ALTER TABLE bronze.payouts ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

-- SILVER
ALTER TABLE silver.predicted_game_outcomes ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
ALTER TABLE silver.predicted_game_outcomes ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

ALTER TABLE silver.simulated_tournaments ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
ALTER TABLE silver.simulated_tournaments ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

ALTER TABLE silver.predicted_market_share ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
ALTER TABLE silver.predicted_market_share ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

-- GOLD
ALTER TABLE gold.optimization_runs ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
ALTER TABLE gold.optimization_runs ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

ALTER TABLE gold.recommended_entry_bids ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
ALTER TABLE gold.recommended_entry_bids ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

ALTER TABLE gold.entry_simulation_outcomes ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
ALTER TABLE gold.entry_simulation_outcomes ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

ALTER TABLE gold.entry_performance ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
ALTER TABLE gold.entry_performance ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

ALTER TABLE gold.detailed_investment_report ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
ALTER TABLE gold.detailed_investment_report ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

-- PUBLIC
ALTER TABLE public.api_keys ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
ALTER TABLE public.api_keys ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

ALTER TABLE public.label_permissions ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
ALTER TABLE public.label_permissions ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

ALTER TABLE public.auth_sessions ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
ALTER TABLE public.grants ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

-- Ensure updated_at triggers exist anywhere updated_at is present for the schemas we touched.
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

        EXECUTE format(
            'CREATE TRIGGER %I BEFORE UPDATE ON %I.%I FOR EACH ROW EXECUTE FUNCTION public.set_updated_at()',
            'set_updated_at',
            r.schema_name,
            r.table_name
        );
    END LOOP;
END;
$$;
