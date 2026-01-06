import argparse
import json
import os
import sys
from pathlib import Path
from typing import Any, Dict, List, Optional


def _ensure_project_root_on_path() -> Path:
    project_root = Path(__file__).resolve().parents[1]
    if str(project_root) not in sys.path:
        sys.path.insert(0, str(project_root))
    return project_root


def _read_calcutta_context(*, calcutta_id: str) -> Dict[str, str]:
    from moneyball.db.connection import get_db_connection

    with get_db_connection() as conn:
        with conn.cursor() as cur:
            cur.execute(
                """
                SELECT
                    c.id::text AS calcutta_id,
                    c.tournament_id::text AS tournament_id,
                    s.year::int AS season_year
                FROM core.calcuttas c
                JOIN core.tournaments t
                  ON t.id = c.tournament_id
                 AND t.deleted_at IS NULL
                JOIN core.seasons s
                  ON s.id = t.season_id
                 AND s.deleted_at IS NULL
                WHERE c.id = %s::uuid
                  AND c.deleted_at IS NULL
                LIMIT 1
                """,
                (calcutta_id,),
            )
            row = cur.fetchone()
            if not row:
                raise ValueError(f"calcutta not found: {calcutta_id}")
            return {
                "calcutta_id": str(row[0]),
                "tournament_id": str(row[1]),
                "season_year": str(row[2]),
            }


def _read_team_id_map(*, tournament_id: str) -> Dict[str, str]:
    from moneyball.db.connection import get_db_connection

    with get_db_connection() as conn:
        with conn.cursor() as cur:
            cur.execute(
                """
                SELECT s.slug, tt.id
                FROM core.teams tt
                JOIN core.schools s
                  ON s.id = tt.school_id
                 AND s.deleted_at IS NULL
                WHERE tt.tournament_id = %s::uuid
                  AND tt.deleted_at IS NULL
                """,
                (tournament_id,),
            )
            return {str(r[0]): str(r[1]) for r in cur.fetchall()}


def _default_train_years(*, target_year: int) -> List[int]:
    all_years = [2017, 2018, 2019, 2021, 2022, 2023, 2024, 2025]
    return [y for y in all_years if int(y) != int(target_year)]


def main() -> int:
    _ensure_project_root_on_path()

    parser = argparse.ArgumentParser(
        description=(
            "Run market-share model for a specific calcutta and persist "
            "artifacts."
        )
    )
    parser.add_argument("--calcutta-id", required=True)
    parser.add_argument("--excluded-entry-name", required=True)
    parser.add_argument("--ridge-alpha", type=float, default=1.0)
    parser.add_argument("--feature-set", default="optimal")
    parser.add_argument("--algorithm-name", default="ridge")
    parser.add_argument("--train-years", default="")

    args = parser.parse_args()

    try:
        from moneyball.db.readers import read_ridge_team_dataset_for_year
        from moneyball.models.predicted_auction_share_of_pool import (
            predict_auction_share_of_pool,
        )
        from moneyball.db.writers.silver_writers import (
            write_predicted_market_share_with_run,
        )

        excluded_entry_name = str(args.excluded_entry_name)
        ctx = _read_calcutta_context(calcutta_id=str(args.calcutta_id))
        tournament_id = ctx["tournament_id"]
        season_year = int(ctx["season_year"])
        team_id_map = _read_team_id_map(tournament_id=tournament_id)

        if args.train_years.strip():
            train_years = [
                int(y.strip())
                for y in args.train_years.split(",")
                if y.strip()
            ]
        else:
            train_years = _default_train_years(target_year=season_year)

        exclude_entry_names: Optional[List[str]] = (
            [excluded_entry_name] if excluded_entry_name else None
        )

        train_frames = []
        for y in train_years:
            try:
                df = read_ridge_team_dataset_for_year(
                    y,
                    exclude_entry_names=exclude_entry_names,
                    include_target=True,
                )
            except Exception:
                continue
            train_frames.append(df)

        if not train_frames:
            raise ValueError("no training frames loaded")

        import pandas as pd

        train_ds = pd.concat(train_frames, ignore_index=True)
        if "team_share_of_pool" in train_ds.columns:
            train_ds = train_ds[train_ds["team_share_of_pool"].notna()].copy()
        if train_ds.empty:
            raise ValueError("no training rows (team_share_of_pool all NULL)")

        predict_ds = read_ridge_team_dataset_for_year(
            season_year,
            exclude_entry_names=None,
            include_target=False,
        )

        predictions = predict_auction_share_of_pool(
            train_team_dataset=train_ds,
            predict_team_dataset=predict_ds,
            ridge_alpha=float(args.ridge_alpha),
            feature_set=str(args.feature_set),
        )

        predictions["team_key"] = (
            predictions["team_key"].astype(str).str.split(":").str[-1]
        )

        git_sha = os.getenv("GIT_SHA")
        params: Dict[str, Any] = {
            "ridge_alpha": float(args.ridge_alpha),
            "feature_set": str(args.feature_set),
            "excluded_entry_name": excluded_entry_name,
            "train_years": train_years,
            "source": "python_runner",
        }

        run_id, inserted = write_predicted_market_share_with_run(
            predictions_df=predictions,
            team_id_map=team_id_map,
            calcutta_id=str(args.calcutta_id),
            tournament_id=None,
            algorithm_name=str(args.algorithm_name),
            params=params,
            git_sha=git_sha,
        )

        out = {
            "ok": True,
            "run_id": run_id,
            "rows_inserted": inserted,
            "calcutta_id": str(args.calcutta_id),
            "tournament_id": tournament_id,
            "season_year": season_year,
            "excluded_entry_name": excluded_entry_name,
            "algorithm_name": str(args.algorithm_name),
            "ridge_alpha": float(args.ridge_alpha),
            "feature_set": str(args.feature_set),
        }
        sys.stdout.write(json.dumps(out) + "\n")
        return 0

    except Exception as e:
        out = {
            "ok": False,
            "error": str(e),
        }
        sys.stdout.write(json.dumps(out) + "\n")
        return 1


if __name__ == "__main__":
    raise SystemExit(main())
