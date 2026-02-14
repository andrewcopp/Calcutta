-- Best-effort rollback.
--
-- A1 rollback: clear snapshot ids that match legacy suite_scenarios snapshot ids.
UPDATE derived.synthetic_calcuttas dst
SET calcutta_snapshot_id = NULL,
    updated_at = NOW()
FROM derived.suite_scenarios src
WHERE dst.id = src.id
  AND dst.deleted_at IS NULL
  AND src.deleted_at IS NULL
  AND dst.calcutta_snapshot_id = src.calcutta_snapshot_id;

-- A2 rollback: detach and delete snapshots created by the up migration.
UPDATE derived.synthetic_calcuttas sc
SET calcutta_snapshot_id = NULL,
    updated_at = NOW()
WHERE sc.calcutta_snapshot_id IN (
    SELECT id
    FROM core.calcutta_snapshots
    WHERE snapshot_type = 'synthetic_calcutta'
      AND description = 'Synthetic calcutta snapshot (migration backfill)'
);

DELETE FROM core.calcutta_snapshots s
WHERE s.snapshot_type = 'synthetic_calcutta'
  AND s.description = 'Synthetic calcutta snapshot (migration backfill)'
  AND NOT EXISTS (
      SELECT 1
      FROM analytics.calcutta_evaluation_runs cer
      WHERE cer.calcutta_snapshot_id = s.id
  );
