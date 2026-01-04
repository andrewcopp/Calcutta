-- Phase 3 rollback: restore legacy derived dimension tables (if they exist).

DROP VIEW IF EXISTS derived.calcuttas;
DROP VIEW IF EXISTS derived.teams;
DROP VIEW IF EXISTS derived.tournaments;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_class c
        JOIN pg_namespace n ON n.oid = c.relnamespace
        WHERE n.nspname = 'derived'
          AND c.relname = 'tournaments_legacy'
          AND c.relkind = 'r'
    ) THEN
        ALTER TABLE derived.tournaments_legacy RENAME TO tournaments;
    END IF;

    IF EXISTS (
        SELECT 1
        FROM pg_class c
        JOIN pg_namespace n ON n.oid = c.relnamespace
        WHERE n.nspname = 'derived'
          AND c.relname = 'teams_legacy'
          AND c.relkind = 'r'
    ) THEN
        ALTER TABLE derived.teams_legacy RENAME TO teams;
    END IF;

    IF EXISTS (
        SELECT 1
        FROM pg_class c
        JOIN pg_namespace n ON n.oid = c.relnamespace
        WHERE n.nspname = 'derived'
          AND c.relname = 'calcuttas_legacy'
          AND c.relkind = 'r'
    ) THEN
        ALTER TABLE derived.calcuttas_legacy RENAME TO calcuttas;
    END IF;
END $$;
