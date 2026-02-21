-- Rename max_bid to max_bid_points to match domain naming conventions (points suffix for in-game currency)
ALTER TABLE core.calcuttas RENAME COLUMN max_bid TO max_bid_points;

-- Rename constraints to reflect new column name and standardized naming
ALTER TABLE core.calcuttas RENAME CONSTRAINT chk_calcuttas_max_bid_positive TO ck_core_calcuttas_max_bid_points_positive;
ALTER TABLE core.calcuttas RENAME CONSTRAINT ck_core_calcuttas_max_bid_le_budget TO ck_core_calcuttas_max_bid_points_le_budget;

-- Also apply constraint renames from 20260222000001 that failed due to the column mismatch
ALTER TABLE core.calcuttas RENAME CONSTRAINT chk_calcuttas_budget_positive TO ck_core_calcuttas_budget_positive;
ALTER TABLE core.calcuttas RENAME CONSTRAINT chk_calcuttas_max_teams_gte_min TO ck_core_calcuttas_max_teams_gte_min;
ALTER TABLE core.calcuttas RENAME CONSTRAINT chk_calcuttas_min_teams TO ck_core_calcuttas_min_teams;
ALTER TABLE core.calcuttas RENAME CONSTRAINT ck_calcuttas_visibility TO ck_core_calcuttas_visibility;
