SET search_path = '';

ALTER TABLE compute.predicted_team_values
    ADD COLUMN actual_points double precision;

ALTER TABLE compute.predicted_team_values
    ADD CONSTRAINT chk_actual_points_non_negative CHECK (actual_points >= 0 OR actual_points IS NULL);

ALTER TABLE compute.predicted_team_values
    ADD CONSTRAINT chk_actual_points_lte_expected CHECK (actual_points <= expected_points OR actual_points IS NULL);
