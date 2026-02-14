-- Rollback: remove all backfilled core data.
-- Note: this is intended only for immediate rollback during development.

DELETE FROM core.entry_teams;
DELETE FROM core.entries;
DELETE FROM core.payouts;
DELETE FROM core.calcutta_scoring_rules;
DELETE FROM core.calcuttas;
DELETE FROM core.team_kenpom_stats;
DELETE FROM core.teams;
DELETE FROM core.schools;
DELETE FROM core.tournaments;
DELETE FROM core.competitions;
DELETE FROM core.seasons;
