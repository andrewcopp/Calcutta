-- Pre-launch naming cleanup: backfill empty calcutta names and standardize
-- constraint/index names that become permanent API surface after launch.

-- 1. Backfill any calcuttas with empty names (from the DEFAULT '' in 20260221000000).
-- Uses the tournament + calcutta ID to generate a placeholder name.
UPDATE core.calcuttas c
SET name = 'Pool ' || LEFT(c.id::text, 8)
WHERE c.name = '' OR c.name IS NULL;

-- 2. Rename auto-generated or mis-prefixed unique constraint/index names.

-- idx_ prefix → uq_ for unique index on tournaments
ALTER INDEX IF EXISTS idx_tournaments_competition_season_active
    RENAME TO uq_core_tournaments_competition_season_active;

-- Auto-generated name → standardized
ALTER TABLE lab.evaluation_entry_results
    RENAME CONSTRAINT evaluation_entry_results_evaluation_id_entry_name_key
    TO uq_lab_evaluation_entry_results_evaluation_entry;

-- Auto-generated _key_key suffix → standardized
ALTER TABLE core.permissions
    RENAME CONSTRAINT permissions_key_key
    TO uq_core_permissions_key;

ALTER TABLE core.labels
    RENAME CONSTRAINT labels_key_key
    TO uq_core_labels_key;
