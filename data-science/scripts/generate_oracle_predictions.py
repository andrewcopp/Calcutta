#!/usr/bin/env python3
"""
Generate oracle predictions using actual historical market shares.

This creates the "perfect information" benchmark by copying actual
market behavior to predicted_market_share. Measures upper bound of
what's achievable if we predicted the market perfectly.

Pipeline stage 2 (oracle variant):
1. Register model (lab.investment_models) - already done: oracle-actual-market
2. Generate predictions (this script) -> predictions_json
3. Optimize entry (optimize_lab_entries.py) -> bids_json
4. Evaluate (evaluate_lab_entries.py) -> lab.evaluations

Usage:
    python generate_oracle_predictions.py
    python generate_oracle_predictions.py --years 2024,2025
    python generate_oracle_predictions.py --calcutta-id <uuid> --json-output
"""
import argparse
import json
import sys
from pathlib import Path
from typing import Any, Dict, List, Optional

# Add project root to path
project_root = Path(__file__).resolve().parents[1]
if str(project_root) not in sys.path:
    sys.path.insert(0, str(project_root))


def get_historical_calcuttas():
    """Get all historical calcuttas from the database."""
    from moneyball.db.connection import get_db_connection

    with get_db_connection() as conn:
        with conn.cursor() as cur:
            cur.execute("""
                SELECT c.id, c.name, s.year, t.id as tournament_id
                FROM core.calcuttas c
                JOIN core.tournaments t ON t.id = c.tournament_id AND t.deleted_at IS NULL
                JOIN core.seasons s ON s.id = t.season_id AND s.deleted_at IS NULL
                WHERE c.deleted_at IS NULL
                ORDER BY s.year DESC
            """)
            return [
                {"id": str(row[0]), "name": row[1], "year": row[2], "tournament_id": str(row[3])}
                for row in cur.fetchall()
            ]


def get_team_id_map(tournament_id: str) -> Dict[str, str]:
    """Get mapping from school_slug to team_id for a tournament."""
    from moneyball.db.connection import get_db_connection

    with get_db_connection() as conn:
        with conn.cursor() as cur:
            cur.execute("""
                SELECT s.slug, t.id
                FROM core.teams t
                JOIN core.schools s ON s.id = t.school_id AND s.deleted_at IS NULL
                WHERE t.tournament_id = %s AND t.deleted_at IS NULL
            """, (tournament_id,))
            return {row[0]: str(row[1]) for row in cur.fetchall()}


def get_actual_market_shares(
    calcutta_id: str,
    tournament_id: str,
    excluded_entry_names: Optional[List[str]] = None,
) -> Dict[str, float]:
    """
    Query actual historical market shares from core.entry_teams.

    Returns:
        Dict mapping team_slug -> market_share (0.0-1.0, sums to 1.0)
    """
    from moneyball.db.connection import get_db_connection

    exclude = [
        str(n).strip()
        for n in (excluded_entry_names or [])
        if str(n).strip()
    ]
    exclude_clause = ""
    params: List[Any] = [calcutta_id]
    if exclude:
        exclude_clause = " AND ce.name <> ALL(%s::text[]) "
        params.append(exclude)

    query = f"""
    WITH team_totals AS (
        SELECT
            s.slug AS team_slug,
            SUM(cet.bid_points)::float8 AS team_total
        FROM core.entry_teams cet
        JOIN core.entries ce
          ON ce.id = cet.entry_id
         AND ce.deleted_at IS NULL
        JOIN core.teams tt
          ON tt.id = cet.team_id
         AND tt.deleted_at IS NULL
        JOIN core.schools s
          ON s.id = tt.school_id
         AND s.deleted_at IS NULL
        WHERE ce.calcutta_id = %s::uuid
          AND cet.deleted_at IS NULL
          {exclude_clause}
        GROUP BY s.slug
    ),
    denom AS (
        SELECT COALESCE(SUM(team_total), 0)::float8 AS total
        FROM team_totals
    )
    SELECT
        s.slug AS team_slug,
        CASE
            WHEN (SELECT total FROM denom) > 0 THEN
                COALESCE(ttt.team_total, 0)::float8 / (SELECT total FROM denom)
            ELSE 0::float8
        END AS market_share
    FROM core.teams tt
    JOIN core.schools s
      ON s.id = tt.school_id
     AND s.deleted_at IS NULL
    LEFT JOIN team_totals ttt ON ttt.team_slug = s.slug
    WHERE tt.tournament_id = %s::uuid
      AND tt.deleted_at IS NULL
    ORDER BY tt.seed ASC, s.name ASC
    """

    params = [*params, tournament_id]
    with get_db_connection() as conn:
        with conn.cursor() as cur:
            cur.execute(query, tuple(params))
            rows = cur.fetchall()

    result = {row[0]: float(row[1]) for row in rows}

    # Normalize to ensure sum is exactly 1.0
    total = sum(result.values())
    if total > 0:
        result = {k: v / total for k, v in result.items()}
    elif result:
        # If no bids, distribute uniformly
        uniform = 1.0 / len(result)
        result = {k: uniform for k in result}

    return result


def get_expected_points_map(calcutta_id: str) -> Dict[str, float]:
    """
    Get expected tournament points for each team from simulation data.

    Uses same logic as generate_lab_predictions.py for consistency.
    """
    from moneyball.db.connection import get_db_connection

    with get_db_connection() as conn:
        with conn.cursor() as cur:
            cur.execute("""
                WITH calcutta_ctx AS (
                    SELECT c.id AS calcutta_id, t.id AS tournament_id
                    FROM core.calcuttas c
                    JOIN core.tournaments t ON t.id = c.tournament_id AND t.deleted_at IS NULL
                    WHERE c.id = %s AND c.deleted_at IS NULL
                ),
                win_distribution AS (
                    SELECT
                        st.team_id,
                        st.wins,
                        st.byes,
                        COUNT(*)::float AS sim_count
                    FROM derived.simulated_teams st
                    WHERE st.tournament_id = (SELECT tournament_id FROM calcutta_ctx)
                    GROUP BY st.team_id, st.wins, st.byes
                ),
                team_totals AS (
                    SELECT team_id, SUM(sim_count) AS total_sims
                    FROM win_distribution
                    GROUP BY team_id
                ),
                team_expected AS (
                    SELECT
                        s.slug AS team_slug,
                        SUM(
                            wd.sim_count * core.calcutta_points_for_progress(
                                (SELECT calcutta_id FROM calcutta_ctx),
                                wd.wins,
                                wd.byes
                            )
                        ) / tt.total_sims AS expected_points
                    FROM win_distribution wd
                    JOIN team_totals tt ON tt.team_id = wd.team_id
                    JOIN core.teams t ON t.id = wd.team_id AND t.deleted_at IS NULL
                    JOIN core.schools s ON s.id = t.school_id AND s.deleted_at IS NULL
                    GROUP BY s.slug, tt.total_sims
                )
                SELECT team_slug, expected_points::float FROM team_expected
            """, (calcutta_id,))
            result = {row[0]: row[1] for row in cur.fetchall()}

            # Fall back to seed-based estimates if no simulation data
            if not result:
                seed_expected_points = {
                    1: 80.0, 2: 55.0, 3: 42.0, 4: 35.0,
                    5: 28.0, 6: 23.0, 7: 19.0, 8: 16.0,
                    9: 14.0, 10: 12.0, 11: 10.0, 12: 9.0,
                    13: 5.0, 14: 4.0, 15: 2.5, 16: 1.0,
                }
                cur.execute("""
                    SELECT s.slug, t.seed
                    FROM core.teams t
                    JOIN core.schools s ON s.id = t.school_id AND s.deleted_at IS NULL
                    JOIN core.calcuttas c ON c.tournament_id = t.tournament_id AND c.deleted_at IS NULL
                    WHERE c.id = %s AND t.deleted_at IS NULL
                """, (calcutta_id,))
                result = {row[0]: seed_expected_points.get(row[1], 10.0) for row in cur.fetchall()}

            return result


def create_oracle_predictions_for_calcutta(
    model_id: str,
    calcutta_id: str,
    tournament_id: str,
    excluded_entry_name: Optional[str] = None,
):
    """Create a lab entry with oracle (actual market) predictions."""
    from moneyball.lab.models import Prediction, create_entry_with_predictions

    excluded_names = [excluded_entry_name] if excluded_entry_name else None
    actual_shares = get_actual_market_shares(calcutta_id, tournament_id, excluded_names)
    if not actual_shares:
        return None

    team_id_map = get_team_id_map(tournament_id)
    expected_points_map = get_expected_points_map(calcutta_id)

    predictions = []
    for team_slug, market_share in actual_shares.items():
        team_id = team_id_map.get(team_slug)
        if not team_id:
            continue

        expected_points = expected_points_map.get(team_slug, 10.0)

        predictions.append(Prediction(
            team_id=team_id,
            predicted_market_share=market_share,
            expected_points=expected_points,
        ))

    if not predictions:
        return None

    # Validate market shares sum to 1.0
    total_share = sum(p.predicted_market_share for p in predictions)
    if abs(total_share - 1.0) > 0.001:
        # Renormalize
        for p in predictions:
            p.predicted_market_share = p.predicted_market_share / total_share

    entry = create_entry_with_predictions(
        investment_model_id=model_id,
        calcutta_id=calcutta_id,
        predictions=predictions,
        game_outcome_kind="kenpom",
        game_outcome_params={},
        starting_state_key="post_first_four",
    )

    return entry


def main():
    parser = argparse.ArgumentParser(
        description="Generate oracle predictions using actual historical market shares"
    )
    parser.add_argument(
        "--model-name",
        default="oracle-actual-market",
        help="Lab model name (default: oracle-actual-market)",
    )
    parser.add_argument(
        "--calcutta-id",
        help="Process only this specific calcutta (for pipeline worker)",
    )
    parser.add_argument(
        "--excluded-entry",
        default="Andrew Copp",
        help="Entry name to exclude from market (default: Andrew Copp)",
    )
    parser.add_argument(
        "--years",
        type=str,
        help="Comma-separated years to process (default: all)",
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Show what would be done without writing",
    )
    parser.add_argument(
        "--json-output",
        action="store_true",
        help="Output machine-readable JSON result",
    )

    args = parser.parse_args()

    def log(msg):
        if not args.json_output:
            print(msg)

    from moneyball.lab.models import get_investment_model

    model = get_investment_model(args.model_name)
    if not model:
        if args.json_output:
            print(json.dumps({"ok": False, "entries_created": 0, "errors": ["Model not found"]}))
            sys.exit(1)
        else:
            print(f"Error: Model '{args.model_name}' not found")
            print("Run: python scripts/register_investment_models.py")
            sys.exit(1)

    log(f"Model: {model.name} ({model.kind})")
    log(f"Excluded entry: {args.excluded_entry}")
    log("")

    calcuttas = get_historical_calcuttas()
    log(f"Found {len(calcuttas)} historical calcuttas")

    # Filter by calcutta_id if specified
    if args.calcutta_id:
        calcuttas = [c for c in calcuttas if c["id"] == args.calcutta_id]
        if not calcuttas:
            if args.json_output:
                print(json.dumps({"ok": False, "entry_id": None, "error": f"Calcutta {args.calcutta_id} not found"}))
                sys.exit(1)
            else:
                print(f"Error: Calcutta {args.calcutta_id} not found")
                sys.exit(1)
        log(f"Processing single calcutta: {calcuttas[0]['name']}")

    # Filter by years if specified
    if args.years:
        target_years = [int(y.strip()) for y in args.years.split(",")]
        calcuttas = [c for c in calcuttas if c["year"] in target_years]
        log(f"Filtered to {len(calcuttas)} calcuttas for years: {target_years}")

    log("")

    entries_created = 0
    errors = []
    last_entry_id = None

    for calcutta in calcuttas:
        log(f"Processing {calcutta['name']} ({calcutta['year']})...")

        if args.dry_run:
            actual_shares = get_actual_market_shares(
                calcutta["id"],
                calcutta["tournament_id"],
                [args.excluded_entry] if args.excluded_entry else None,
            )
            log(f"  Would create entry with {len(actual_shares)} predictions")
            log(f"  Market share sum: {sum(actual_shares.values()):.4f}")
        else:
            entry = create_oracle_predictions_for_calcutta(
                model.id,
                calcutta["id"],
                calcutta["tournament_id"],
                args.excluded_entry,
            )
            if entry:
                log(f"  Created entry {entry.id} with {len(entry.predictions)} predictions")
                entries_created += 1
                last_entry_id = entry.id
            else:
                log(f"  No entry created (no market data)")
                errors.append(f"{calcutta['name']}: No market data")

    log("")
    log("Done!")
    log(f"Created {entries_created} entries with oracle predictions")
    log("Run optimize_lab_entries.py to generate optimized bids")

    if args.json_output:
        if args.calcutta_id and entries_created == 1:
            result = {
                "ok": True,
                "entry_id": last_entry_id,
                "errors": errors if errors else [],
            }
        else:
            result = {
                "ok": True,
                "entries_created": entries_created,
                "errors": errors if errors else [],
            }
        print(json.dumps(result))


if __name__ == "__main__":
    main()
