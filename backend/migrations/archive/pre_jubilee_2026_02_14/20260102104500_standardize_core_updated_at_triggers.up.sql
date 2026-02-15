-- Ensure core tables use core.set_updated_at() for updated_at triggers.
-- Removes the legacy trigger named 'set_updated_at' that still points at public.set_updated_at().

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
        WHERE n.nspname = 'core'
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
