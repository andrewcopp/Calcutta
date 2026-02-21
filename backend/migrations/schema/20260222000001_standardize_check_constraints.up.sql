-- 2H: Standardize CHECK constraint naming to ck_{schema}_{table}_{description}

-- core.teams: chk_ -> ck_core_teams_
ALTER TABLE core.teams RENAME CONSTRAINT chk_teams_seed_range TO ck_core_teams_seed_range;
ALTER TABLE core.teams RENAME CONSTRAINT chk_teams_byes_range TO ck_core_teams_byes_range;
ALTER TABLE core.teams RENAME CONSTRAINT chk_teams_wins_range TO ck_core_teams_wins_range;

-- core.calcuttas: chk_ -> ck_core_calcuttas_
ALTER TABLE core.calcuttas RENAME CONSTRAINT chk_calcuttas_budget_positive TO ck_core_calcuttas_budget_positive;
ALTER TABLE core.calcuttas RENAME CONSTRAINT chk_calcuttas_max_bid_points_positive TO ck_core_calcuttas_max_bid_points_positive;
ALTER TABLE core.calcuttas RENAME CONSTRAINT chk_calcuttas_max_teams_gte_min TO ck_core_calcuttas_max_teams_gte_min;
ALTER TABLE core.calcuttas RENAME CONSTRAINT chk_calcuttas_min_teams TO ck_core_calcuttas_min_teams;
-- ck_core_calcuttas_max_bid_points_le_budget already has correct prefix
-- ck_calcuttas_visibility: missing schema prefix
ALTER TABLE core.calcuttas RENAME CONSTRAINT ck_calcuttas_visibility TO ck_core_calcuttas_visibility;

-- core.calcutta_scoring_rules: chk_ -> ck_core_
ALTER TABLE core.calcutta_scoring_rules RENAME CONSTRAINT chk_scoring_rules_points_nonneg TO ck_core_calcutta_scoring_rules_points_nonneg;
ALTER TABLE core.calcutta_scoring_rules RENAME CONSTRAINT chk_scoring_rules_win_index_nonneg TO ck_core_calcutta_scoring_rules_win_index_nonneg;

-- core.calcutta_invitations: already ck_ but missing schema
ALTER TABLE core.calcutta_invitations RENAME CONSTRAINT ck_calcutta_invitations_status TO ck_core_calcutta_invitations_status;

-- core.entry_teams: chk_ -> ck_core_
ALTER TABLE core.entry_teams RENAME CONSTRAINT chk_entry_teams_bid_positive TO ck_core_entry_teams_bid_positive;

-- core.payouts: chk_ -> ck_core_
ALTER TABLE core.payouts RENAME CONSTRAINT chk_payouts_amount_nonneg TO ck_core_payouts_amount_nonneg;
ALTER TABLE core.payouts RENAME CONSTRAINT chk_payouts_position_positive TO ck_core_payouts_position_positive;
