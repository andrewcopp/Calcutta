ALTER TABLE IF EXISTS derived.strategy_generation_run_bids
    RENAME TO recommended_entry_bids;

-- Rename any derived constraints that include the new table name.
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
            AND t.relname = 'recommended_entry_bids'
            AND c.conname LIKE '%strategy_generation_run_bids%'
    ) LOOP
        EXECUTE format(
            'ALTER TABLE derived.recommended_entry_bids RENAME CONSTRAINT %I TO %I',
            r.conname,
            replace(r.conname, 'strategy_generation_run_bids', 'recommended_entry_bids')
        );
    END LOOP;
END;
$$;

-- Rename any derived indexes that include the new table name.
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
            AND t.relname = 'recommended_entry_bids'
            AND c.relname LIKE '%strategy_generation_run_bids%'
    ) LOOP
        EXECUTE format(
            'ALTER INDEX IF EXISTS %I.%I RENAME TO %I',
            'derived',
            r.index_name,
            replace(r.index_name, 'strategy_generation_run_bids', 'recommended_entry_bids')
        );
    END LOOP;
END;
$$;
