-- Domain check constraints on core tables to enforce business invariants at the DB level.

-- entry_teams: bid must be positive
ALTER TABLE core.entry_teams ADD CONSTRAINT chk_entry_teams_bid_positive CHECK (bid_points > 0);

-- calcuttas: budget must be positive, team limits must be sensible, max_bid must be positive
ALTER TABLE core.calcuttas ADD CONSTRAINT chk_calcuttas_budget_positive CHECK (budget_points > 0);
ALTER TABLE core.calcuttas ADD CONSTRAINT chk_calcuttas_min_teams CHECK (min_teams >= 1);
ALTER TABLE core.calcuttas ADD CONSTRAINT chk_calcuttas_max_bid_positive CHECK (max_bid > 0);
ALTER TABLE core.calcuttas ADD CONSTRAINT chk_calcuttas_max_teams_gte_min CHECK (max_teams >= min_teams);

-- teams: seed within 1-16, wins non-negative
ALTER TABLE core.teams ADD CONSTRAINT chk_teams_seed_range CHECK (seed BETWEEN 1 AND 16);
ALTER TABLE core.teams ADD CONSTRAINT chk_teams_wins_nonneg CHECK (wins >= 0);

-- payouts: position positive, amount non-negative
ALTER TABLE core.payouts ADD CONSTRAINT chk_payouts_position_positive CHECK ("position" >= 1);
ALTER TABLE core.payouts ADD CONSTRAINT chk_payouts_amount_nonneg CHECK (amount_cents >= 0);

-- calcutta_scoring_rules: points and win_index non-negative
ALTER TABLE core.calcutta_scoring_rules ADD CONSTRAINT chk_scoring_rules_points_nonneg CHECK (points_awarded >= 0);
ALTER TABLE core.calcutta_scoring_rules ADD CONSTRAINT chk_scoring_rules_win_index_nonneg CHECK (win_index >= 0);
