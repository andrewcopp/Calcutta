-- Jubilee Seed Data Down Migration
-- Truncate in reverse dependency order

TRUNCATE core.calcutta_snapshot_scoring_rules, core.calcutta_snapshot_payouts,
  core.calcutta_snapshot_entry_teams, core.calcutta_snapshot_entries,
  core.calcutta_snapshots CASCADE;
TRUNCATE core.entry_teams, core.entries, core.payouts,
  core.calcutta_scoring_rules, core.calcuttas CASCADE;
TRUNCATE core.team_kenpom_stats, core.teams, core.tournaments,
  core.schools, core.competitions, core.seasons CASCADE;
TRUNCATE core.grants, core.label_permissions, core.labels,
  core.permissions, core.users CASCADE;
