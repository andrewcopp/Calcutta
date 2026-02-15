ALTER TABLE IF EXISTS derived.recommended_entry_bids
    RENAME TO strategy_generation_run_bids;

-- Rename any derived constraints that still include the old table name.
DO $$
DECLARE
    r record;
BEGIN
    FOR r IN (
        SELECT c.conname
        FROM pg_constraint c
        JOIN pg_class t ON t.oid = c.conrelid
        JOIN pg_namespace n ON n.oid = t.relnamespace
        WHERE n.nspname = 'derived'
            AND t.relname = 'strategy_generation_run_bids'
            AND c.conname LIKE '%recommended_entry_bids%'
    ) LOOP
        EXECUTE format(
            'ALTER TABLE derived.strategy_generation_run_bids RENAME CONSTRAINT %I TO %I',
            r.conname,
            replace(r.conname, 'recommended_entry_bids', 'strategy_generation_run_bids')
        );
    END LOOP;
END;
$$;

-- Rename any derived indexes that still include the old table name.
DO $$
DECLARE
    r record;
BEGIN
    FOR r IN (
        SELECT c.relname AS index_name
        FROM pg_class c
        JOIN pg_index i ON i.indexrelid = c.oid
        JOIN pg_class t ON t.oid = i.indrelid
        JOIN pg_namespace n ON n.oid = t.relnamespace
        WHERE n.nspname = 'derived'
            AND t.relname = 'strategy_generation_run_bids'
            AND c.relname LIKE '%recommended_entry_bids%'
    ) LOOP
        EXECUTE format(
            'ALTER INDEX IF EXISTS %I.%I RENAME TO %I',
            'derived',
            r.index_name,
            replace(r.index_name, 'recommended_entry_bids', 'strategy_generation_run_bids')
        );
    END LOOP;
END;
$$;
