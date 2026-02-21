-- Rollback: restore original constraint/index names.
-- Note: backfilled calcutta names are NOT rolled back (data is better with names).

ALTER TABLE core.labels
    RENAME CONSTRAINT uq_core_labels_key
    TO labels_key_key;

ALTER TABLE core.permissions
    RENAME CONSTRAINT uq_core_permissions_key
    TO permissions_key_key;

ALTER TABLE lab.evaluation_entry_results
    RENAME CONSTRAINT uq_lab_evaluation_entry_results_evaluation_entry
    TO evaluation_entry_results_evaluation_id_entry_name_key;

ALTER INDEX IF EXISTS uq_core_tournaments_competition_season_active
    RENAME TO idx_tournaments_competition_season_active;
