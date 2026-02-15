DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_class c
        JOIN pg_namespace n ON n.oid = c.relnamespace
        WHERE c.relkind = 'i'
          AND c.relname = 'uq_analytics_tournament_simulation_batches_natural_key'
          AND n.nspname = 'analytics'
    ) THEN
        CREATE UNIQUE INDEX uq_analytics_tournament_simulation_batches_natural_key
        ON analytics.tournament_simulation_batches(
            tournament_id,
            tournament_state_snapshot_id,
            n_sims,
            seed,
            probability_source_key
        )
        WHERE deleted_at IS NULL;
    END IF;
END $$;

-- Create one legacy tournament snapshot per tournament referenced by existing simulations.
WITH tournaments AS (
    SELECT DISTINCT
        bt.core_tournament_id AS core_tournament_id
    FROM analytics.simulated_tournaments st
    JOIN lab_bronze.tournaments bt ON bt.id = st.tournament_id
    WHERE st.deleted_at IS NULL
      AND bt.core_tournament_id IS NOT NULL
)
INSERT INTO analytics.tournament_state_snapshots (
    tournament_id,
    source,
    description
)
SELECT
    t.core_tournament_id,
    'legacy_backfill' AS source,
    'Legacy backfill snapshot' AS description
FROM tournaments t
WHERE NOT EXISTS (
    SELECT 1
    FROM analytics.tournament_state_snapshots s
    WHERE s.tournament_id = t.core_tournament_id
      AND s.source = 'legacy_backfill'
      AND s.deleted_at IS NULL
);

-- Populate teams for legacy snapshots using current core tournament state.
WITH legacy_snapshots AS (
    SELECT
        s.id,
        s.tournament_id
    FROM analytics.tournament_state_snapshots s
    WHERE s.source = 'legacy_backfill'
      AND s.deleted_at IS NULL
)
INSERT INTO analytics.tournament_state_snapshot_teams (
    tournament_state_snapshot_id,
    team_id,
    wins,
    byes,
    eliminated
)
SELECT
    ls.id,
    ct.id,
    ct.wins,
    ct.byes,
    ct.eliminated
FROM legacy_snapshots ls
JOIN core.teams ct
  ON ct.tournament_id = ls.tournament_id
 AND ct.deleted_at IS NULL
ON CONFLICT (tournament_state_snapshot_id, team_id) DO NOTHING;

-- Create one legacy simulation batch per tournament.
WITH sim_meta AS (
    SELECT
        st.tournament_id AS lab_tournament_id,
        COUNT(DISTINCT st.sim_id)::int AS n_sims
    FROM analytics.simulated_tournaments st
    WHERE st.deleted_at IS NULL
    GROUP BY st.tournament_id
),
lab_to_core AS (
    SELECT
        bt.id AS lab_tournament_id,
        bt.core_tournament_id AS core_tournament_id
    FROM lab_bronze.tournaments bt
    WHERE bt.core_tournament_id IS NOT NULL
),
legacy_snapshots AS (
    SELECT
        s.id AS snapshot_id,
        s.tournament_id
    FROM analytics.tournament_state_snapshots s
    WHERE s.source = 'legacy_backfill'
      AND s.deleted_at IS NULL
),
target AS (
    SELECT
        ltc.lab_tournament_id,
        ltc.core_tournament_id,
        ls.snapshot_id,
        sm.n_sims
    FROM lab_to_core ltc
    JOIN sim_meta sm ON sm.lab_tournament_id = ltc.lab_tournament_id
    JOIN legacy_snapshots ls ON ls.tournament_id = ltc.core_tournament_id
)
INSERT INTO analytics.tournament_simulation_batches (
    tournament_id,
    tournament_state_snapshot_id,
    n_sims,
    seed,
    probability_source_key
)
SELECT
    t.core_tournament_id,
    t.snapshot_id,
    t.n_sims,
    0 AS seed,
    'legacy_unknown' AS probability_source_key
FROM target t
WHERE NOT EXISTS (
    SELECT 1
    FROM analytics.tournament_simulation_batches b
    WHERE b.tournament_id = t.core_tournament_id
      AND b.tournament_state_snapshot_id = t.snapshot_id
      AND b.n_sims = t.n_sims
      AND b.seed = 0
      AND b.probability_source_key = 'legacy_unknown'
      AND b.deleted_at IS NULL
);

-- Create a legacy "as-is" calcutta snapshot per calcutta.
WITH calcuttas AS (
    SELECT
        c.id AS base_calcutta_id
    FROM core.calcuttas c
    WHERE c.deleted_at IS NULL
)
INSERT INTO core.calcutta_snapshots (
    base_calcutta_id,
    snapshot_type,
    description
)
SELECT
    c.base_calcutta_id,
    'as_is' AS snapshot_type,
    'Legacy backfill snapshot' AS description
FROM calcuttas c
WHERE NOT EXISTS (
    SELECT 1
    FROM core.calcutta_snapshots cs
    WHERE cs.base_calcutta_id = c.base_calcutta_id
      AND cs.description = 'Legacy backfill snapshot'
      AND cs.deleted_at IS NULL
);

-- Snapshot entries.
WITH legacy_calcutta_snapshots AS (
    SELECT
        cs.id AS calcutta_snapshot_id,
        cs.base_calcutta_id
    FROM core.calcutta_snapshots cs
    WHERE cs.description = 'Legacy backfill snapshot'
      AND cs.deleted_at IS NULL
)
INSERT INTO core.calcutta_snapshot_entries (
    calcutta_snapshot_id,
    entry_id,
    display_name,
    is_synthetic
)
SELECT
    lcs.calcutta_snapshot_id,
    e.id,
    e.name,
    FALSE
FROM legacy_calcutta_snapshots lcs
JOIN core.entries e
  ON e.calcutta_id = lcs.base_calcutta_id
 AND e.deleted_at IS NULL
ON CONFLICT (calcutta_snapshot_id, display_name) DO NOTHING;

-- Snapshot entry teams.
WITH legacy_entries AS (
    SELECT
        cse.id AS calcutta_snapshot_entry_id,
        cse.entry_id
    FROM core.calcutta_snapshot_entries cse
    JOIN core.calcutta_snapshots cs ON cs.id = cse.calcutta_snapshot_id
    WHERE cs.description = 'Legacy backfill snapshot'
      AND cs.deleted_at IS NULL
      AND cse.deleted_at IS NULL
      AND cse.entry_id IS NOT NULL
)
INSERT INTO core.calcutta_snapshot_entry_teams (
    calcutta_snapshot_entry_id,
    team_id,
    bid_points
)
SELECT
    le.calcutta_snapshot_entry_id,
    cet.team_id,
    cet.bid_points
FROM legacy_entries le
JOIN core.entry_teams cet
  ON cet.entry_id = le.entry_id
 AND cet.deleted_at IS NULL
ON CONFLICT (calcutta_snapshot_entry_id, team_id) DO NOTHING;

-- Snapshot payouts.
WITH legacy_calcutta_snapshots AS (
    SELECT
        cs.id AS calcutta_snapshot_id,
        cs.base_calcutta_id
    FROM core.calcutta_snapshots cs
    WHERE cs.description = 'Legacy backfill snapshot'
      AND cs.deleted_at IS NULL
)
INSERT INTO core.calcutta_snapshot_payouts (
    calcutta_snapshot_id,
    position,
    amount_cents
)
SELECT
    lcs.calcutta_snapshot_id,
    p.position,
    p.amount_cents
FROM legacy_calcutta_snapshots lcs
JOIN core.payouts p
  ON p.calcutta_id = lcs.base_calcutta_id
 AND p.deleted_at IS NULL
ON CONFLICT (calcutta_snapshot_id, position) DO NOTHING;

-- Snapshot scoring rules.
WITH legacy_calcutta_snapshots AS (
    SELECT
        cs.id AS calcutta_snapshot_id,
        cs.base_calcutta_id
    FROM core.calcutta_snapshots cs
    WHERE cs.description = 'Legacy backfill snapshot'
      AND cs.deleted_at IS NULL
)
INSERT INTO core.calcutta_snapshot_scoring_rules (
    calcutta_snapshot_id,
    win_index,
    points_awarded
)
SELECT
    lcs.calcutta_snapshot_id,
    csr.win_index,
    csr.points_awarded
FROM legacy_calcutta_snapshots lcs
JOIN core.calcutta_scoring_rules csr
  ON csr.calcutta_id = lcs.base_calcutta_id
 AND csr.deleted_at IS NULL
ON CONFLICT (calcutta_snapshot_id, win_index) DO NOTHING;

-- Create one legacy evaluation run per calcutta snapshot, using the legacy simulation batch for that tournament.
WITH legacy_calcutta_snapshots AS (
    SELECT
        cs.id AS calcutta_snapshot_id,
        cs.base_calcutta_id
    FROM core.calcutta_snapshots cs
    WHERE cs.description = 'Legacy backfill snapshot'
      AND cs.deleted_at IS NULL
),
legacy_batches AS (
    SELECT
        b.id AS batch_id,
        b.tournament_id
    FROM analytics.tournament_simulation_batches b
    JOIN analytics.tournament_state_snapshots s
      ON s.id = b.tournament_state_snapshot_id
    WHERE b.deleted_at IS NULL
      AND b.seed = 0
      AND b.probability_source_key = 'legacy_unknown'
      AND s.source = 'legacy_backfill'
      AND s.deleted_at IS NULL
)
INSERT INTO analytics.calcutta_evaluation_runs (
    tournament_simulation_batch_id,
    calcutta_snapshot_id,
    purpose
)
SELECT
    lb.batch_id,
    lcs.calcutta_snapshot_id,
    'legacy_backfill' AS purpose
FROM legacy_calcutta_snapshots lcs
JOIN core.calcuttas c
  ON c.id = lcs.base_calcutta_id
 AND c.deleted_at IS NULL
JOIN legacy_batches lb
  ON lb.tournament_id = c.tournament_id
WHERE NOT EXISTS (
    SELECT 1
    FROM analytics.calcutta_evaluation_runs cer
    WHERE cer.calcutta_snapshot_id = lcs.calcutta_snapshot_id
      AND cer.tournament_simulation_batch_id = lb.batch_id
      AND cer.purpose = 'legacy_backfill'
      AND cer.deleted_at IS NULL
);
