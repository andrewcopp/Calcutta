-- A1: Backfill synthetic calcuttas snapshot IDs from legacy suite_scenarios where possible.
UPDATE derived.synthetic_calcuttas dst
SET calcutta_snapshot_id = src.calcutta_snapshot_id,
    updated_at = NOW()
FROM derived.suite_scenarios src
WHERE dst.id = src.id
  AND dst.deleted_at IS NULL
  AND src.deleted_at IS NULL
  AND dst.calcutta_snapshot_id IS NULL
  AND src.calcutta_snapshot_id IS NOT NULL;

-- A2: For any remaining synthetic calcuttas without a snapshot, create one by copying the base calcutta.
DO $$
DECLARE
    r RECORD;
    v_snapshot_id UUID;
    e RECORD;
    v_snapshot_entry_id UUID;
    v_focus_name TEXT;
BEGIN
    FOR r IN (
        SELECT
            sc.id,
            sc.calcutta_id,
            sc.excluded_entry_name,
            sc.focus_strategy_generation_run_id,
            sc.focus_entry_name
        FROM derived.synthetic_calcuttas sc
        WHERE sc.deleted_at IS NULL
          AND sc.calcutta_snapshot_id IS NULL
    ) LOOP
        INSERT INTO core.calcutta_snapshots (base_calcutta_id, snapshot_type, description)
        VALUES (r.calcutta_id, 'synthetic_calcutta', 'Synthetic calcutta snapshot (migration backfill)')
        RETURNING id INTO v_snapshot_id;

        INSERT INTO core.calcutta_snapshot_payouts (calcutta_snapshot_id, position, amount_cents)
        SELECT v_snapshot_id, p.position, p.amount_cents
        FROM core.payouts p
        WHERE p.calcutta_id = r.calcutta_id
          AND p.deleted_at IS NULL;

        INSERT INTO core.calcutta_snapshot_scoring_rules (calcutta_snapshot_id, win_index, points_awarded)
        SELECT v_snapshot_id, csr.win_index, csr.points_awarded
        FROM core.calcutta_scoring_rules csr
        WHERE csr.calcutta_id = r.calcutta_id
          AND csr.deleted_at IS NULL;

        FOR e IN (
            SELECT ce.id, ce.name
            FROM core.entries ce
            WHERE ce.calcutta_id = r.calcutta_id
              AND ce.deleted_at IS NULL
              AND (r.excluded_entry_name IS NULL OR ce.name <> r.excluded_entry_name)
            ORDER BY ce.created_at ASC
        ) LOOP
            INSERT INTO core.calcutta_snapshot_entries (calcutta_snapshot_id, entry_id, display_name, is_synthetic)
            VALUES (v_snapshot_id, e.id, e.name, false)
            RETURNING id INTO v_snapshot_entry_id;

            INSERT INTO core.calcutta_snapshot_entry_teams (calcutta_snapshot_entry_id, team_id, bid_points)
            SELECT v_snapshot_entry_id, et.team_id, et.bid_points
            FROM core.entry_teams et
            WHERE et.entry_id = e.id
              AND et.deleted_at IS NULL;
        END LOOP;

        IF r.focus_strategy_generation_run_id IS NOT NULL AND EXISTS (
            SELECT 1
            FROM derived.recommended_entry_bids reb
            WHERE reb.strategy_generation_run_id = r.focus_strategy_generation_run_id
              AND reb.deleted_at IS NULL
        ) THEN
            v_focus_name := COALESCE(NULLIF(BTRIM(r.focus_entry_name), ''), 'Our Strategy');

            INSERT INTO core.calcutta_snapshot_entries (calcutta_snapshot_id, entry_id, display_name, is_synthetic)
            VALUES (v_snapshot_id, NULL, v_focus_name, true)
            RETURNING id INTO v_snapshot_entry_id;

            INSERT INTO core.calcutta_snapshot_entry_teams (calcutta_snapshot_entry_id, team_id, bid_points)
            SELECT v_snapshot_entry_id, reb.team_id, reb.bid_points
            FROM derived.recommended_entry_bids reb
            WHERE reb.strategy_generation_run_id = r.focus_strategy_generation_run_id
              AND reb.deleted_at IS NULL;
        END IF;

        UPDATE derived.synthetic_calcuttas sc
        SET calcutta_snapshot_id = v_snapshot_id,
            updated_at = NOW()
        WHERE sc.id = r.id;
    END LOOP;
END $$;
