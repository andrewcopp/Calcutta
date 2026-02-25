SET search_path = '';

ALTER TABLE compute.predicted_team_values
    DROP CONSTRAINT IF EXISTS chk_actual_points_lte_expected;

ALTER TABLE compute.predicted_team_values
    DROP CONSTRAINT IF EXISTS chk_actual_points_non_negative;

ALTER TABLE compute.predicted_team_values
    DROP COLUMN IF EXISTS actual_points;
