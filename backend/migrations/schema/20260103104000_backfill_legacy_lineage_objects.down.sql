-- Best-effort rollback for legacy lineage backfill.
-- Only deletes rows created by this backfill (identified by source/purpose/description keys)
-- and clears FK columns that were set.

-- Delete eval runs
DELETE FROM analytics.calcutta_evaluation_runs
WHERE purpose = 'legacy_backfill';

-- Delete calcutta snapshots created by backfill (cascades to entries/teams/payouts/scoring rules)
DELETE FROM core.calcutta_snapshots
WHERE description = 'Legacy backfill snapshot';

-- Delete simulation batches + snapshots
DELETE FROM analytics.tournament_simulation_batches
WHERE seed = 0
  AND probability_source_key = 'legacy_unknown';

DELETE FROM analytics.tournament_state_snapshots
WHERE source = 'legacy_backfill'
  AND description = 'Legacy backfill snapshot';

DROP INDEX IF EXISTS analytics.uq_analytics_tournament_simulation_batches_natural_key;
