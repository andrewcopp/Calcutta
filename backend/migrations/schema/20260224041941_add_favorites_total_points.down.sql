-- Rollback: add_favorites_total_points
-- Created: 2026-02-24 04:19:41 UTC

SET search_path = '';
SET lock_timeout = '5s';
SET statement_timeout = '30s';

ALTER TABLE IF EXISTS compute.predicted_team_values
    DROP CONSTRAINT IF EXISTS ck_compute_predicted_team_values_favorites_total_points_gte0;

ALTER TABLE compute.predicted_team_values
    DROP COLUMN IF EXISTS favorites_total_points;
