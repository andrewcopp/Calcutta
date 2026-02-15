-- Recreate legacy core triggers named 'set_updated_at' that execute public.set_updated_at().
-- This intentionally reintroduces the redundant trigger behavior as a rollback.

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

        EXECUTE format(
            'CREATE TRIGGER %I BEFORE UPDATE ON %I.%I FOR EACH ROW EXECUTE FUNCTION public.set_updated_at()',
            'set_updated_at',
            r.schema_name,
            r.table_name
        );
    END LOOP;
END;
$$;
