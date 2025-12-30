from __future__ import annotations

from pathlib import Path
from typing import Any, Dict, List, Optional, Tuple

import pandas as pd

from moneyball.models.predicted_auction_share_of_pool import (
    predict_auction_share_of_pool_from_out_root,
)
from moneyball.models.predicted_game_outcomes import (
    predict_game_outcomes_from_snapshot,
)
from moneyball.models.recommended_entry_bids import (
    recommend_entry_bids,
)
from moneyball.models.simulated_tournaments import (
    simulate_tournaments,
)
from moneyball.models.simulated_entry_outcomes import (
    simulate_entry_outcomes,
)
from moneyball.models.investment_report import (
    generate_investment_report,
)
from moneyball.pipeline.orchestrator import (
    ArtifactStore,
    PipelineOrchestrator,
)
from moneyball.pipeline.artifacts import (
    ensure_dir,
    fingerprint_file,
    fingerprints_to_dict,
    load_manifest,
    manifest_matches,
    sha256_jsonable,
    utc_now_iso,
    write_json,
)


def _list_snapshots_under_out_root(out_root: Path) -> Dict[str, Path]:
    if not out_root.exists():
        return {}
    out: Dict[str, Path] = {}
    for p in sorted(out_root.iterdir()):
        if not p.is_dir():
            continue
        if (p / "derived" / "team_dataset.parquet").exists():
            out[p.name] = p
    return out


def _stage_predicted_game_outcomes(
    *,
    snapshot_dir: Path,
    out_dir: Path,
    calcutta_key: Optional[str],
    kenpom_scale: float,
    n_sims: int,
    seed: int,
    use_cache: bool,
    manifest: Dict[str, Any],
) -> Tuple[Path, Dict[str, Any]]:
    stage = "predicted_game_outcomes"

    sd = Path(snapshot_dir)
    games_path = sd / "games.parquet"
    teams_path = sd / "teams.parquet"

    if not games_path.exists():
        raise FileNotFoundError(f"missing required file: {games_path}")
    if not teams_path.exists():
        raise FileNotFoundError(f"missing required file: {teams_path}")

    # Store in canonical location (no timestamp)
    canonical_dir = sd / "derived"
    ensure_dir(canonical_dir)
    out_path = canonical_dir / "predicted_game_outcomes.parquet"
    manifest_path = canonical_dir / "predicted_game_outcomes_manifest.json"

    input_fps = {
        "games": fingerprint_file(games_path),
        "teams": fingerprint_file(teams_path),
    }

    stage_config: Dict[str, Any] = {
        "calcutta_key": calcutta_key,
        "kenpom_scale": float(kenpom_scale),
        "n_sims": int(n_sims),
        "seed": int(seed),
    }

    # Check cache using canonical manifest
    if use_cache and out_path.exists():
        existing_manifest = load_manifest(manifest_path)
        if existing_manifest is not None and manifest_matches(
            existing=existing_manifest,
            stage=stage,
            stage_config=stage_config,
            input_fingerprints=input_fps,
        ):
            print("✓ Using cached predicted_game_outcomes")
            stage_manifest = {
                "stage_config_hash": sha256_jsonable(stage_config),
                "stage_config": stage_config,
                "inputs": fingerprints_to_dict(input_fps),
                "outputs": {
                    "predicted_game_outcomes": str(out_path),
                },
                "cached": True,
            }
            stages = manifest.setdefault("stages", {})
            if isinstance(stages, dict):
                stages[stage] = stage_manifest
            return out_path, manifest

    # Generate predictions
    print("⚙ Generating predicted_game_outcomes...")
    df = predict_game_outcomes_from_snapshot(
        snapshot_dir=snapshot_dir,
        calcutta_key=calcutta_key,
        kenpom_scale=float(kenpom_scale),
        n_sims=int(n_sims),
        seed=int(seed),
    )

    df.to_parquet(out_path, index=False)

    stage_manifest = {
        "stage_config_hash": sha256_jsonable(stage_config),
        "stage_config": stage_config,
        "inputs": fingerprints_to_dict(input_fps),
        "outputs": {
            "predicted_game_outcomes": str(out_path),
        },
    }

    # Save canonical manifest
    canonical_manifest = {"stages": {stage: stage_manifest}}
    write_json(manifest_path, canonical_manifest)

    stages = manifest.setdefault("stages", {})
    if isinstance(stages, dict):
        stages[stage] = stage_manifest
    return out_path, manifest


def _stage_predicted_auction_share_of_pool(
    *,
    snapshot_dir: Path,
    out_dir: Path,
    train_snapshots: Optional[List[str]],
    ridge_alpha: float,
    feature_set: str,
    exclude_entry_names: Optional[List[str]],
    use_cache: bool,
    manifest: Dict[str, Any],
) -> Tuple[Path, Dict[str, Any]]:
    stage = "predicted_auction_share_of_pool"

    sd = Path(snapshot_dir)
    out_root = sd.parent
    predict_snapshot = sd.name

    snapshot_dirs = _list_snapshots_under_out_root(out_root)
    if predict_snapshot not in snapshot_dirs:
        raise FileNotFoundError(
            "predict snapshot not found under out_root: "
            f"{predict_snapshot}"
        )

    if train_snapshots is None:
        train_names = [
            k for k in snapshot_dirs.keys() if k != predict_snapshot
        ]
    else:
        train_names = [
            str(s).strip() for s in train_snapshots if str(s).strip()
        ]

    prev_name = ""
    try:
        prev_name = str(int(str(predict_snapshot)) - 1)
    except Exception:
        prev_name = ""

    snapshots_for_inputs = set(train_names + [predict_snapshot])
    if prev_name and prev_name in snapshot_dirs:
        snapshots_for_inputs.add(prev_name)

    input_fps: Dict[str, Any] = {}
    for name in sorted(snapshots_for_inputs):
        sdir = snapshot_dirs.get(name)
        if sdir is None:
            continue
        required = {
            "team_dataset": sdir / "derived" / "team_dataset.parquet",
            "entries": sdir / "entries.parquet",
            "entry_bids": sdir / "entry_bids.parquet",
            "teams": sdir / "teams.parquet",
            "payouts": sdir / "payouts.parquet",
        }
        for k, p in required.items():
            if not p.exists():
                raise FileNotFoundError(f"missing required file: {p}")
            input_fps[f"{name}:{k}"] = fingerprint_file(p)

    stage_config: Dict[str, Any] = {
        "predict_snapshot": str(predict_snapshot),
        "train_snapshots": list(train_names),
        "ridge_alpha": float(ridge_alpha),
        "feature_set": str(feature_set),
        "exclude_entry_names": [
            str(n) for n in (exclude_entry_names or []) if str(n).strip()
        ],
    }

    out_path = out_dir / "predicted_auction_share_of_pool.parquet"

    if use_cache:
        existing = load_manifest(out_dir / "manifest.json")
        if existing is not None and manifest_matches(
            existing=existing,
            stage=stage,
            stage_config=stage_config,
            input_fingerprints=input_fps,
        ):
            if out_path.exists():
                return out_path, existing

    df = predict_auction_share_of_pool_from_out_root(
        out_root=out_root,
        predict_snapshot=str(predict_snapshot),
        train_snapshots=list(train_names),
        ridge_alpha=float(ridge_alpha),
        feature_set=str(feature_set),
        exclude_entry_names=list(stage_config["exclude_entry_names"]),
    )

    ensure_dir(out_path.parent)
    df.to_parquet(out_path, index=False)

    stage_manifest = {
        "stage_config_hash": sha256_jsonable(stage_config),
        "stage_config": stage_config,
        "inputs": fingerprints_to_dict(input_fps),
        "outputs": {
            "predicted_auction_share_of_pool": str(out_path),
        },
    }

    stages = manifest.setdefault("stages", {})
    if isinstance(stages, dict):
        stages[stage] = stage_manifest
    return out_path, manifest


def _stage_recommended_entry_bids(
    *,
    snapshot_dir: Path,
    out_dir: Path,
    budget_points: int,
    min_teams: int,
    max_teams: int,
    max_per_team_points: int,
    min_bid_points: int,
    predicted_total_pool_bids_points: Optional[float],
    strategy: str = "greedy",
    use_cache: bool,
    manifest: Dict[str, Any],
) -> Tuple[Path, Dict[str, Any]]:
    stage = "recommended_entry_bids"

    # Load from canonical location (no timestamp)
    predicted_game_outcomes_path = Path(snapshot_dir) / "derived" / "predicted_game_outcomes.parquet"
    predicted_share_path = out_dir / "predicted_auction_share_of_pool.parquet"
    entries_path = Path(snapshot_dir) / "entries.parquet"

    if not predicted_game_outcomes_path.exists():
        raise FileNotFoundError(
            "missing required artifact: "
            f"{predicted_game_outcomes_path}"
        )
    if not predicted_share_path.exists():
        raise FileNotFoundError(
            "missing required artifact: "
            f"{predicted_share_path}"
        )
    if not entries_path.exists():
        raise FileNotFoundError(f"missing required file: {entries_path}")

    input_fps = {
        "predicted_game_outcomes": fingerprint_file(
            predicted_game_outcomes_path
        ),
        "predicted_auction_share_of_pool": fingerprint_file(
            predicted_share_path
        ),
        "entries": fingerprint_file(entries_path),
    }

    if predicted_total_pool_bids_points is None:
        entries = pd.read_parquet(entries_path)
        n_entries = int(entries["entry_key"].astype(str).nunique())
        predicted_total_pool_bids_points = float(n_entries) * float(
            int(budget_points)
        )

    stage_config: Dict[str, Any] = {
        "budget_points": int(budget_points),
        "min_teams": int(min_teams),
        "max_teams": int(max_teams),
        "max_per_team_points": int(max_per_team_points),
        "min_bid_points": int(min_bid_points),
        "predicted_total_pool_bids_points": float(
            predicted_total_pool_bids_points
        ),
        "strategy": str(strategy),
    }

    out_path = out_dir / "recommended_entry_bids.parquet"

    if use_cache:
        existing = load_manifest(out_dir / "manifest.json")
        if existing is not None and manifest_matches(
            existing=existing,
            stage=stage,
            stage_config=stage_config,
            input_fingerprints=input_fps,
        ):
            if out_path.exists():
                return out_path, existing

    predicted_game_outcomes = pd.read_parquet(predicted_game_outcomes_path)
    predicted_share = pd.read_parquet(predicted_share_path)

    df = recommend_entry_bids(
        predicted_auction_share_of_pool=predicted_share,
        predicted_game_outcomes=predicted_game_outcomes,
        predicted_total_pool_bids_points=predicted_total_pool_bids_points,
        budget_points=int(budget_points),
        min_teams=int(min_teams),
        max_teams=int(max_teams),
        max_per_team_points=int(max_per_team_points),
        min_bid_points=int(min_bid_points),
        strategy=str(strategy),
    )

    ensure_dir(out_path.parent)
    df.to_parquet(out_path, index=False)

    stage_manifest = {
        "stage_config_hash": sha256_jsonable(stage_config),
        "stage_config": stage_config,
        "inputs": fingerprints_to_dict(input_fps),
        "outputs": {
            "recommended_entry_bids": str(out_path),
        },
    }

    stages = manifest.setdefault("stages", {})
    if isinstance(stages, dict):
        stages[stage] = stage_manifest
    return out_path, manifest


def _stage_simulated_tournaments(
    *,
    snapshot_dir: Path,
    out_dir: Path,
    n_sims: int,
    seed: int,
    regenerate: bool,
    use_cache: bool,
    manifest: Dict[str, Any],
) -> Tuple[Path, Dict[str, Any]]:
    stage = "simulated_tournaments"

    sd = Path(snapshot_dir)
    games_path = sd / "games.parquet"

    if not games_path.exists():
        raise FileNotFoundError(f"missing required file: {games_path}")

    # Load from canonical location (no timestamp)
    predicted_game_outcomes_path = sd / "derived" / "predicted_game_outcomes.parquet"

    if not predicted_game_outcomes_path.exists():
        raise FileNotFoundError(
            f"missing required artifact: {predicted_game_outcomes_path}"
        )

    # Store tournaments in canonical location (independent of Calcutta runs)
    tournaments_dir = sd / "derived"
    ensure_dir(tournaments_dir)
    out_path = tournaments_dir / "tournaments.parquet"
    manifest_path = tournaments_dir / "tournaments_manifest.json"

    # Cache key only depends on games data, not predicted_game_outcomes
    # (since predicted_game_outcomes is regenerated each run)
    input_fps = {
        "games": fingerprint_file(games_path),
    }

    stage_config = {
        "n_sims": n_sims,
        "seed": seed,
    }

    # Check cache unless regenerate flag is set
    if use_cache and not regenerate and out_path.exists():
        existing_manifest = load_manifest(manifest_path)
        if existing_manifest is not None and manifest_matches(
            existing=existing_manifest,
            stage=stage,
            stage_config=stage_config,
            input_fingerprints=input_fps,
        ):
            print(f"✓ Using cached tournaments ({n_sims} sims)")
            stage_manifest = {
                "stage_config_hash": sha256_jsonable(stage_config),
                "stage_config": stage_config,
                "inputs": fingerprints_to_dict(input_fps),
                "outputs": {
                    "simulated_tournaments": str(out_path),
                },
                "cached": True,
            }
            stages = manifest.setdefault("stages", {})
            if isinstance(stages, dict):
                stages[stage] = stage_manifest
            return out_path, manifest

    # Run simulation
    print(f"⚙ Generating tournaments ({n_sims} sims)...")
    games = pd.read_parquet(games_path)
    predicted_game_outcomes = pd.read_parquet(predicted_game_outcomes_path)

    result = simulate_tournaments(
        games=games,
        predicted_game_outcomes=predicted_game_outcomes,
        n_sims=n_sims,
        seed=seed,
    )

    result.to_parquet(out_path, index=False)

    stage_manifest = {
        "stage_config_hash": sha256_jsonable(stage_config),
        "stage_config": stage_config,
        "inputs": fingerprints_to_dict(input_fps),
        "outputs": {
            "simulated_tournaments": str(out_path),
        },
    }

    # Save manifest for future cache checks
    tournament_manifest = {"stages": {stage: stage_manifest}}
    write_json(manifest_path, tournament_manifest)

    stages = manifest.setdefault("stages", {})
    if isinstance(stages, dict):
        stages[stage] = stage_manifest

    return out_path, manifest


def _stage_simulated_entry_outcomes(
    *,
    snapshot_dir: Path,
    out_dir: Path,
    calcutta_key: Optional[str],
    n_sims: int,
    seed: int,
    budget_points: int,
    keep_sims: bool,
    use_cache: bool,
    manifest: Dict[str, Any],
) -> Tuple[Path, Dict[str, Any]]:
    stage = "simulated_entry_outcomes"

    sd = Path(snapshot_dir)
    games_path = sd / "games.parquet"
    teams_path = sd / "teams.parquet"
    payouts_path = sd / "payouts.parquet"
    entry_bids_path = sd / "entry_bids.parquet"

    if not games_path.exists():
        raise FileNotFoundError(f"missing required file: {games_path}")
    if not teams_path.exists():
        raise FileNotFoundError(f"missing required file: {teams_path}")
    if not payouts_path.exists():
        raise FileNotFoundError(f"missing required file: {payouts_path}")
    if not entry_bids_path.exists():
        raise FileNotFoundError(f"missing required file: {entry_bids_path}")

    # Load from canonical location (no timestamp)
    predicted_game_outcomes_path = Path(snapshot_dir) / "derived" / "predicted_game_outcomes.parquet"
    recommended_entry_bids_path = out_dir / "recommended_entry_bids.parquet"

    if not predicted_game_outcomes_path.exists():
        raise FileNotFoundError(
            "missing required artifact: "
            f"{predicted_game_outcomes_path}"
        )
    if not recommended_entry_bids_path.exists():
        raise FileNotFoundError(
            "missing required artifact: "
            f"{recommended_entry_bids_path}"
        )

    # Load tournaments from canonical location
    tournaments_path = sd / "derived" / "tournaments.parquet"

    simulated_tournaments_df = None
    if tournaments_path.exists():
        simulated_tournaments_df = pd.read_parquet(tournaments_path)
        print("✓ Loaded tournaments for entry outcomes")
    else:
        print("⚠ No tournaments found - run simulate-tournaments first")

    input_fps = {
        "games": fingerprint_file(games_path),
        "teams": fingerprint_file(teams_path),
        "payouts": fingerprint_file(payouts_path),
        "entry_bids": fingerprint_file(entry_bids_path),
        "predicted_game_outcomes": fingerprint_file(
            predicted_game_outcomes_path
        ),
        "recommended_entry_bids": fingerprint_file(
            recommended_entry_bids_path
        ),
    }
    
    # Add simulated_tournaments to inputs if it exists
    if simulated_tournaments_df is not None:
        input_fps["simulated_tournaments"] = fingerprint_file(
            tournaments_path
        )

    ck = calcutta_key
    if ck is None:
        entry_bids_df = pd.read_parquet(entry_bids_path)
        if "calcutta_key" in entry_bids_df.columns and not entry_bids_df.empty:
            ck = str(entry_bids_df["calcutta_key"].iloc[0])
        else:
            teams_df = pd.read_parquet(teams_path)
            if "calcutta_key" in teams_df.columns and not teams_df.empty:
                ck = str(teams_df["calcutta_key"].iloc[0])
            else:
                raise ValueError(
                    "calcutta_key not provided and cannot be inferred "
                    "from entry_bids or teams"
                )

    teams_df = pd.read_parquet(teams_path)

    stage_config: Dict[str, Any] = {
        "calcutta_key": str(ck),
        "n_sims": int(n_sims),
        "seed": int(seed),
        "budget_points": int(budget_points),
        "keep_sims": bool(keep_sims),
    }

    out_summary_path = out_dir / "simulated_entry_outcomes.parquet"
    out_sims_path = out_dir / "simulated_entry_outcomes_sims.parquet"

    if use_cache:
        existing = load_manifest(out_dir / "manifest.json")
        if existing is not None and manifest_matches(
            existing=existing,
            stage=stage,
            stage_config=stage_config,
            input_fingerprints=input_fps,
        ):
            if out_summary_path.exists():
                if not keep_sims:
                    return out_summary_path, existing
                if out_sims_path.exists():
                    return out_summary_path, existing

    games_df = pd.read_parquet(games_path)
    payouts_df = pd.read_parquet(payouts_path)
    entry_bids_df = pd.read_parquet(entry_bids_path)
    predicted_game_outcomes_df = pd.read_parquet(predicted_game_outcomes_path)
    recommended_entry_bids_df = pd.read_parquet(recommended_entry_bids_path)

    summary_df, sims_df = simulate_entry_outcomes(
        games=games_df,
        teams=teams_df,
        payouts=payouts_df,
        entry_bids=entry_bids_df,
        predicted_game_outcomes=predicted_game_outcomes_df,
        recommended_entry_bids=recommended_entry_bids_df,
        simulated_tournaments=simulated_tournaments_df,
        calcutta_key=str(ck),
        n_sims=int(n_sims),
        seed=int(seed),
        budget_points=int(budget_points),
        keep_sims=bool(keep_sims),
    )

    ensure_dir(out_summary_path.parent)
    summary_df.to_parquet(out_summary_path, index=False)
    if keep_sims and sims_df is not None:
        sims_df.to_parquet(out_sims_path, index=False)

    outputs: Dict[str, Any] = {
        "simulated_entry_outcomes": str(out_summary_path),
    }
    if keep_sims:
        outputs["simulated_entry_outcomes_sims"] = str(out_sims_path)

    stage_manifest = {
        "stage_config_hash": sha256_jsonable(stage_config),
        "stage_config": stage_config,
        "inputs": fingerprints_to_dict(input_fps),
        "outputs": outputs,
    }

    stages = manifest.setdefault("stages", {})
    if isinstance(stages, dict):
        stages[stage] = stage_manifest
    return out_summary_path, manifest


def _stage_investment_report(
    *,
    snapshot_dir: Path,
    out_dir: Path,
    snapshot_name: str,
    budget_points: int,
    n_sims: int,
    seed: int,
    use_cache: bool,
    manifest: Dict[str, Any],
) -> Tuple[Path, Dict[str, Any]]:
    stage = "investment_report"

    # Load canonical artifacts (no timestamp)
    pgo_path = Path(snapshot_dir) / "derived" / "predicted_game_outcomes.parquet"
    # Load timestamped Calcutta artifacts
    pas_path = out_dir / "predicted_auction_share_of_pool.parquet"
    reb_path = out_dir / "recommended_entry_bids.parquet"
    seo_path = out_dir / "simulated_entry_outcomes.parquet"

    if not pgo_path.exists():
        raise FileNotFoundError(
            f"missing required artifact: {pgo_path}"
        )
    if not pas_path.exists():
        raise FileNotFoundError(
            f"missing required artifact: {pas_path}"
        )
    if not reb_path.exists():
        raise FileNotFoundError(
            f"missing required artifact: {reb_path}"
        )
    if not seo_path.exists():
        raise FileNotFoundError(
            f"missing required artifact: {seo_path}"
        )

    input_fps = {
        "predicted_game_outcomes": fingerprint_file(pgo_path),
        "predicted_auction_share_of_pool": fingerprint_file(pas_path),
        "recommended_entry_bids": fingerprint_file(reb_path),
        "simulated_entry_outcomes": fingerprint_file(seo_path),
    }

    stage_config: Dict[str, Any] = {
        "snapshot_name": str(snapshot_name),
        "budget_points": int(budget_points),
        "n_sims": int(n_sims),
        "seed": int(seed),
    }

    out_path = out_dir / "investment_report.parquet"

    if use_cache:
        existing = load_manifest(out_dir / "manifest.json")
        if existing is not None and manifest_matches(
            existing=existing,
            stage=stage,
            stage_config=stage_config,
            input_fingerprints=input_fps,
        ):
            if out_path.exists():
                return out_path, existing

    pgo_df = pd.read_parquet(pgo_path)
    pas_df = pd.read_parquet(pas_path)
    reb_df = pd.read_parquet(reb_path)
    seo_df = pd.read_parquet(seo_path)

    report_df = generate_investment_report(
        recommended_entry_bids=reb_df,
        simulated_entry_outcomes=seo_df,
        predicted_game_outcomes=pgo_df,
        predicted_auction_share_of_pool=pas_df,
        snapshot_name=str(snapshot_name),
        budget_points=int(budget_points),
        n_sims=int(n_sims),
        seed=int(seed),
    )

    ensure_dir(out_path.parent)
    report_df.to_parquet(out_path, index=False)

    stage_manifest = {
        "stage_config_hash": sha256_jsonable(stage_config),
        "stage_config": stage_config,
        "inputs": fingerprints_to_dict(input_fps),
        "outputs": {"investment_report": str(out_path)},
    }

    stages = manifest.setdefault("stages", {})
    if isinstance(stages, dict):
        stages[stage] = stage_manifest
    return out_path, manifest


def run(
    *,
    snapshot_dir: Path,
    snapshot_name: Optional[str] = None,
    artifacts_root: Optional[Path] = None,
    run_id: Optional[str] = None,
    stages: Optional[List[str]] = None,
    calcutta_key: Optional[str] = None,
    kenpom_scale: float = 10.0,
    n_sims: int = 5000,
    seed: int = 123,
    auction_train_snapshots: Optional[List[str]] = None,
    auction_ridge_alpha: float = 1.0,
    auction_feature_set: str = "expanded_last_year_expected",
    auction_exclude_entry_names: Optional[List[str]] = None,
    bids_budget_points: int = 100,
    bids_min_teams: int = 3,
    bids_max_teams: int = 10,
    bids_max_per_team_points: int = 50,
    bids_min_bid_points: int = 1,
    bids_predicted_total_pool_bids_points: Optional[float] = None,
    bids_strategy: str = "greedy",
    sim_n_sims: int = 5000,
    sim_seed: int = 123,
    sim_budget_points: int = 100,
    sim_keep_sims: bool = False,
    regenerate_tournaments: bool = False,
    use_cache: bool = True,
) -> Dict[str, Any]:
    """
    Run pipeline stages using orchestrator pattern.

    Refactored to use ArtifactStore and PipelineOrchestrator for
    cleaner separation of concerns.
    """
    # Validate inputs
    sd = Path(snapshot_dir)
    if not sd.exists():
        raise FileNotFoundError(f"snapshot_dir not found: {sd}")

    sname = str(snapshot_name) if snapshot_name is not None else sd.name

    # Initialize artifact store and orchestrator
    store = ArtifactStore(
        snapshot_dir=sd,
        snapshot_name=sname,
        artifacts_root=artifacts_root,
        run_id=run_id,
    )

    orchestrator = PipelineOrchestrator(
        artifact_store=store,
        use_cache=use_cache,
    )

    # Determine which stages to run
    wanted = (
        list(stages) if stages is not None else ["predicted_game_outcomes"]
    )

    # Run stages in order
    if "predicted_game_outcomes" in wanted:
        out_path = orchestrator.run_stage(
            stage_name="predicted_game_outcomes",
            stage_func=_stage_predicted_game_outcomes,
            calcutta_key=calcutta_key,
            kenpom_scale=float(kenpom_scale),
            n_sims=int(n_sims),
            seed=int(seed),
        )
        orchestrator.results["predicted_game_outcomes_parquet"] = str(
            out_path
        )

    if "predicted_auction_share_of_pool" in wanted:
        out_path = orchestrator.run_stage(
            stage_name="predicted_auction_share_of_pool",
            stage_func=_stage_predicted_auction_share_of_pool,
            train_snapshots=auction_train_snapshots,
            ridge_alpha=float(auction_ridge_alpha),
            feature_set=str(auction_feature_set),
            exclude_entry_names=auction_exclude_entry_names,
        )
        orchestrator.results["predicted_auction_share_of_pool_parquet"] = str(
            out_path
        )

    if "recommended_entry_bids" in wanted:
        out_path = orchestrator.run_stage(
            stage_name="recommended_entry_bids",
            stage_func=_stage_recommended_entry_bids,
            budget_points=int(bids_budget_points),
            min_teams=int(bids_min_teams),
            max_teams=int(bids_max_teams),
            max_per_team_points=int(bids_max_per_team_points),
            min_bid_points=int(bids_min_bid_points),
            predicted_total_pool_bids_points=(
                bids_predicted_total_pool_bids_points
            ),
            strategy=str(bids_strategy),
        )
        orchestrator.results["recommended_entry_bids_parquet"] = str(out_path)

    if "simulated_tournaments" in wanted:
        out_path = orchestrator.run_stage(
            stage_name="simulated_tournaments",
            stage_func=_stage_simulated_tournaments,
            n_sims=int(sim_n_sims),
            seed=int(sim_seed),
            regenerate=bool(regenerate_tournaments),
        )
        orchestrator.results["simulated_tournaments_parquet"] = str(out_path)

    if "simulated_entry_outcomes" in wanted:
        out_path = orchestrator.run_stage(
            stage_name="simulated_entry_outcomes",
            stage_func=_stage_simulated_entry_outcomes,
            calcutta_key=calcutta_key,
            n_sims=int(sim_n_sims),
            seed=int(sim_seed),
            budget_points=int(sim_budget_points),
            keep_sims=bool(sim_keep_sims),
        )
        orchestrator.results["simulated_entry_outcomes_parquet"] = str(
            out_path
        )

    if "investment_report" in wanted:
        out_path = orchestrator.run_stage(
            stage_name="investment_report",
            stage_func=_stage_investment_report,
            snapshot_name=sname,
            budget_points=int(sim_budget_points),
            n_sims=int(sim_n_sims),
            seed=int(sim_seed),
        )
        orchestrator.results["investment_report_parquet"] = str(out_path)

    # Finalize and return results
    return orchestrator.finalize()
