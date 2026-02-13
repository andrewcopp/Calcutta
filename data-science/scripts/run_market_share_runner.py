from __future__ import annotations

import argparse
import json
import os
import sys
from pathlib import Path
from typing import Any, Dict, List, Optional


def _split_entry_names(v: object) -> List[str]:
    if v is None:
        return []
    if isinstance(v, list):
        out: List[str] = []
        for x in v:
            s = str(x).strip()
            if s:
                out.append(s)
        return out
    s = str(v).strip()
    if not s:
        return []
    if "," in s:
        return [p.strip() for p in s.split(",") if p.strip()]
    return [s]


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


def _apply_train_years_window(
    *,
    train_years: List[int],
    train_years_window: int,
) -> List[int]:
    w = int(train_years_window)
    if w <= 0:
        return list(train_years)
    ys = sorted({int(y) for y in train_years})
    if not ys:
        return []
    if w >= len(ys):
        return ys
    return ys[-w:]


def _parse_train_years(v: Any) -> Optional[List[int]]:
    if v is None:
        return None
    if isinstance(v, list):
        out: List[int] = []
        for x in v:
            try:
                out.append(int(x))
            except Exception:
                continue
        return out
    if isinstance(v, str):
        s = v.strip()
        if not s:
            return None
        out = []
        for part in s.split(","):
            p = part.strip()
            if not p:
                continue
            try:
                out.append(int(p))
            except Exception:
                continue
        return out
    return None


def _read_market_share_run(*, run_id: str) -> Dict[str, Any]:
    from moneyball.db.connection import get_db_connection

    with get_db_connection() as conn:
        with conn.cursor() as cur:
            cur.execute(
                """
                SELECT
                    r.id::text AS run_id,
                    r.calcutta_id::text AS calcutta_id,
                    r.params_json,
                    a.name AS algorithm_name
                FROM derived.market_share_runs r
                JOIN derived.algorithms a
                  ON a.id = r.algorithm_id
                 AND a.deleted_at IS NULL
                WHERE r.id = %s::uuid
                  AND r.deleted_at IS NULL
                LIMIT 1
                """,
                (run_id,),
            )
            row = cur.fetchone()
            if not row:
                raise ValueError(f"market_share_run not found: {run_id}")
            return {
                "run_id": str(row[0]),
                "calcutta_id": str(row[1]),
                "params_json": row[2] or {},
                "algorithm_name": str(row[3]),
            }


def _compute_oracle_market_share(
    *,
    calcutta_id: str,
    tournament_id: str,
    excluded_entry_names: Optional[List[str]],
) -> Any:
    import pandas as pd
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
            s.slug AS team_key,
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
        s.slug AS team_key,
        CASE
            WHEN (SELECT total FROM denom) > 0 THEN
                COALESCE(ttt.team_total, 0)::float8 / (SELECT total FROM denom)
            ELSE 0::float8
        END AS predicted_auction_share_of_pool
    FROM core.teams tt
    JOIN core.schools s
      ON s.id = tt.school_id
     AND s.deleted_at IS NULL
    LEFT JOIN team_totals ttt ON ttt.team_key = s.slug
    WHERE tt.tournament_id = %s::uuid
      AND tt.deleted_at IS NULL
    ORDER BY tt.seed ASC, s.name ASC
    """

    params = [*params, tournament_id]
    with get_db_connection() as conn:
        df = pd.read_sql_query(query, conn, params=tuple(params))

    df["predicted_auction_share_of_pool"] = pd.to_numeric(
        df["predicted_auction_share_of_pool"], errors="coerce"
    ).fillna(0.0)

    ssum = float(df["predicted_auction_share_of_pool"].sum() or 0.0)
    if ssum <= 0 and len(df) > 0:
        df["predicted_auction_share_of_pool"] = 1.0 / float(len(df))
    elif ssum > 0:
        df["predicted_auction_share_of_pool"] = (
            df["predicted_auction_share_of_pool"] / ssum
        )

    return df[["team_key", "predicted_auction_share_of_pool"]].copy()


def run_market_share_for_calcutta(
    *,
    calcutta_id: str,
    algorithm_name: str,
    excluded_entry_name: str,
    train_years: Optional[List[int]] = None,
    train_years_window: int = 0,
    ridge_alpha: float = 1.0,
    feature_set: str = "optimal",
    target_transform: str = "none",
    seed_prior_monotone: Optional[bool] = None,
    seed_prior_k: float = 0.0,
    program_prior_k: float = 0.0,
    run_id: Optional[str] = None,
) -> Dict[str, Any]:
    _ensure_project_root_on_path()

    from moneyball.db.readers import read_ridge_team_dataset_for_year
    from moneyball.models.predicted_auction_share_of_pool import (
        predict_auction_share_of_pool,
    )
    from moneyball.db.writers.silver_writers import (
        write_predicted_market_share_with_run,
    )

    ctx = _read_calcutta_context(calcutta_id=str(calcutta_id))
    tournament_id = ctx["tournament_id"]
    season_year = int(ctx["season_year"])
    team_id_map = _read_team_id_map(tournament_id=tournament_id)

    excluded_entry_name = str(excluded_entry_name or "").strip()
    if not excluded_entry_name:
        raise ValueError("excluded_entry_name is required")

    excluded_entry_names = _split_entry_names(excluded_entry_name)

    algorithm_name = str(algorithm_name or "").strip()
    if not algorithm_name:
        raise ValueError("algorithm_name is required")

    if algorithm_name == "oracle_actual_market":
        predictions = _compute_oracle_market_share(
            calcutta_id=str(calcutta_id),
            tournament_id=str(tournament_id),
            excluded_entry_names=excluded_entry_names,
        )
        params_out: Dict[str, Any] = {
            "excluded_entry_name": excluded_entry_name,
        }
    else:
        is_underbid_1sigma = str(algorithm_name) == "ridge-v2-underbid-1sigma"

        train_years_eff = train_years
        if not train_years_eff:
            train_years_eff = _default_train_years(target_year=season_year)
            if int(train_years_window or 0) > 0:
                train_years_eff = _apply_train_years_window(
                    train_years=train_years_eff,
                    train_years_window=int(train_years_window),
                )

        train_frames = []
        for y in train_years_eff:
            try:
                df = read_ridge_team_dataset_for_year(
                    int(y),
                    exclude_entry_names=excluded_entry_names,
                    include_target=True,
                )
            except Exception:
                continue
            train_frames.append(df)

        if not train_frames:
            raise ValueError("no training frames loaded")

        import pandas as pd
        import numpy as np

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
            ridge_alpha=float(ridge_alpha),
            feature_set=str(feature_set or "optimal"),
            target_transform=str(target_transform or "none"),
            seed_prior_monotone=seed_prior_monotone,
            seed_prior_k=float(seed_prior_k or 0.0),
            program_prior_k=float(program_prior_k or 0.0),
        )
        predictions["team_key"] = (
            predictions["team_key"].astype(str).str.split(":").str[-1]
        )

        underbid_sigma_global: Optional[float] = None
        underbid_sigma_by_seed: Dict[int, float] = {}
        if is_underbid_1sigma:
            eps = 1e-12
            residuals_all: List[float] = []
            residuals_by_seed: Dict[int, List[float]] = {}

            for i, df_y in enumerate(train_frames):
                if df_y is None or df_y.empty:
                    continue
                df_y = df_y.copy()

                oof_frames = [f for j, f in enumerate(train_frames) if j != i]
                if not oof_frames:
                    continue
                oof_train = pd.concat(oof_frames, ignore_index=True)
                if "team_share_of_pool" in oof_train.columns:
                    oof_train = oof_train[
                        oof_train["team_share_of_pool"].notna()
                    ].copy()
                if oof_train.empty:
                    continue

                pred_y = predict_auction_share_of_pool(
                    train_team_dataset=oof_train,
                    predict_team_dataset=df_y,
                    ridge_alpha=float(ridge_alpha),
                    feature_set=str(feature_set or "optimal"),
                    target_transform=str(target_transform or "none"),
                    seed_prior_monotone=seed_prior_monotone,
                    seed_prior_k=float(seed_prior_k or 0.0),
                    program_prior_k=float(program_prior_k or 0.0),
                )
                if "team_key" in pred_y.columns:
                    pred_y["team_key"] = (
                        pred_y["team_key"].astype(str).str.split(":").str[-1]
                    )
                if "team_key" not in df_y.columns:
                    continue
                df_y["team_key"] = (
                    df_y["team_key"].astype(str).str.split(":").str[-1]
                )

                merged = df_y[
                    ["team_key", "team_share_of_pool", "seed"]
                ].merge(
                    pred_y[["team_key", "predicted_auction_share_of_pool"]],
                    on="team_key",
                    how="inner",
                )
                if merged.empty:
                    continue
                a = pd.to_numeric(
                    merged["team_share_of_pool"],
                    errors="coerce",
                ).fillna(0.0)
                p = pd.to_numeric(
                    merged["predicted_auction_share_of_pool"],
                    errors="coerce",
                ).fillna(0.0)
                s = pd.to_numeric(merged["seed"], errors="coerce")
                res = np.log(a + eps) - np.log(p + eps)
                valid = np.isfinite(res)
                res = res[valid]
                s = s[valid]
                if len(res) == 0:
                    continue
                residuals_all.extend(res.astype(float).tolist())
                for seed_val, rv in zip(s.tolist(), res.tolist()):
                    try:
                        seed_i = int(seed_val)
                    except Exception:
                        continue
                    if seed_i < 1 or seed_i > 16:
                        continue
                    residuals_by_seed.setdefault(seed_i, []).append(float(rv))

            if len(residuals_all) >= 2:
                underbid_sigma_global = float(
                    np.std(np.asarray(residuals_all, dtype=float), ddof=1)
                )
                if (
                    not np.isfinite(underbid_sigma_global)
                    or underbid_sigma_global <= 0
                ):
                    underbid_sigma_global = None

            for seed_i, vals in residuals_by_seed.items():
                if len(vals) < 2:
                    continue
                sig = float(np.std(np.asarray(vals, dtype=float), ddof=1))
                if np.isfinite(sig) and sig > 0:
                    underbid_sigma_by_seed[int(seed_i)] = sig

        if is_underbid_1sigma:
            v = pd.to_numeric(
                predictions["predicted_auction_share_of_pool"],
                errors="coerce",
            ).fillna(0.0)
            if "seed" in predictions.columns:
                seed_series = pd.to_numeric(
                    predictions["seed"],
                    errors="coerce",
                )
            else:
                seed_series = pd.Series(
                    [np.nan] * int(len(predictions)),
                    index=predictions.index,
                    dtype=float,
                )
            base_sigma = float(underbid_sigma_global or 0.0)
            adj: List[float] = []
            for mu, seed_val in zip(v.tolist(), seed_series.tolist()):
                try:
                    seed_i = int(seed_val)
                except Exception:
                    seed_i = 0
                sig = float(underbid_sigma_by_seed.get(seed_i, base_sigma))
                adj.append(float(mu) * float(np.exp(-sig)))
            predictions["predicted_auction_share_of_pool"] = pd.Series(
                adj,
                index=predictions.index,
                dtype=float,
            )

        params_out = {
            "excluded_entry_name": excluded_entry_name,
            "train_years": train_years_eff,
            "train_years_window": int(train_years_window or 0),
            "ridge_alpha": float(ridge_alpha),
            "feature_set": str(feature_set or "optimal"),
            "target_transform": str(target_transform or "none"),
            "seed_prior_monotone": seed_prior_monotone,
            "seed_prior_k": float(seed_prior_k or 0.0),
            "program_prior_k": float(program_prior_k or 0.0),
        }

        if is_underbid_1sigma:
            params_out["underbid_mode"] = "seed_log_1sigma_unscaled"
            params_out["underbid_sigma_global"] = underbid_sigma_global
            params_out["underbid_sigma_by_seed"] = {
                str(k): float(v)
                for k, v in sorted(underbid_sigma_by_seed.items())
            }

    git_sha = os.getenv("GIT_SHA")

    run_id_out, inserted = write_predicted_market_share_with_run(
        predictions_df=predictions,
        team_id_map=team_id_map,
        calcutta_id=str(calcutta_id),
        tournament_id=None,
        algorithm_name=str(algorithm_name),
        run_id=str(run_id) if run_id else None,
        params=params_out,
        git_sha=git_sha,
    )

    return {
        "ok": True,
        "run_id": run_id_out,
        "rows_inserted": inserted,
        "calcutta_id": str(calcutta_id),
        "tournament_id": tournament_id,
        "season_year": season_year,
        "excluded_entry_name": excluded_entry_name,
        "algorithm_name": str(algorithm_name),
    }


def run_market_share_runner(*, run_id: str) -> Dict[str, Any]:
    _ensure_project_root_on_path()

    run = _read_market_share_run(run_id=str(run_id))
    params = run.get("params_json") or {}

    excluded_entry_name = str(params.get("excluded_entry_name") or "").strip()
    if not excluded_entry_name:
        raise ValueError("excluded_entry_name is required in params_json")

    train_years = _parse_train_years(params.get("train_years"))
    train_years_window = int(params.get("train_years_window") or 0)

    return run_market_share_for_calcutta(
        calcutta_id=str(run["calcutta_id"]),
        algorithm_name=str(run["algorithm_name"]),
        excluded_entry_name=excluded_entry_name,
        train_years=train_years,
        train_years_window=train_years_window,
        ridge_alpha=float(params.get("ridge_alpha") or 1.0),
        feature_set=str(params.get("feature_set") or "optimal"),
        target_transform=str(params.get("target_transform") or "none"),
        seed_prior_monotone=(
            None
            if params.get("seed_prior_monotone") is None
            else bool(params.get("seed_prior_monotone"))
        ),
        seed_prior_k=float(params.get("seed_prior_k") or 0.0),
        program_prior_k=float(params.get("program_prior_k") or 0.0),
        run_id=str(run["run_id"]),
    )


def main() -> int:
    parser = argparse.ArgumentParser(
        description=(
            "Run market-share model for a specific calcutta and persist "
            "artifacts."
        )
    )
    parser.add_argument("--run-id")
    parser.add_argument("--calcutta-id")
    parser.add_argument("--excluded-entry-name")
    parser.add_argument("--algorithm-name")

    args = parser.parse_args()

    try:
        if args.run_id:
            out = run_market_share_runner(run_id=str(args.run_id))
        else:
            calcutta_id = str(args.calcutta_id or "").strip()
            excluded = str(args.excluded_entry_name or "").strip()
            algorithm_name = str(args.algorithm_name or "").strip()
            if not calcutta_id:
                raise ValueError(
                    "--calcutta-id is required when --run-id is not provided"
                )
            if not excluded:
                raise ValueError(
                    "--excluded-entry-name is required when --run-id is not "
                    "provided"
                )
            if not algorithm_name:
                raise ValueError(
                    "--algorithm-name is required when --run-id is not "
                    "provided"
                )

            out = run_market_share_for_calcutta(
                calcutta_id=calcutta_id,
                algorithm_name=algorithm_name,
                excluded_entry_name=excluded,
            )
        sys.stdout.write(json.dumps(out) + "\n")
        return 0

    except Exception as e:
        import traceback

        out = {
            "ok": False,
            "error": str(e),
            "error_type": type(e).__name__,
            "traceback": traceback.format_exc(),
        }
        sys.stdout.write(json.dumps(out) + "\n")
        return 1


if __name__ == "__main__":
    raise SystemExit(main())
