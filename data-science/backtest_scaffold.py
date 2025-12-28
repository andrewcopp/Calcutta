from pathlib import Path
from typing import Dict, List, Optional, Tuple

import numpy as np
import pandas as pd

from calcutta_ds import allocation, bracket, io, points, sim, standings


def _read_parquet(p: Path) -> pd.DataFrame:
    return io.read_parquet(p)


def _choose_calcutta_key(df: pd.DataFrame, requested: Optional[str]) -> str:
    return io.choose_calcutta_key(df, requested)


def _waterfill_equal(
    k: int,
    budget: float,
    max_per_team: float,
) -> List[float]:
    return allocation.waterfill_equal(
        k=k,
        budget=budget,
        max_per_team=max_per_team,
    )


def _load_snapshot_tables(snapshot_dir: Path) -> Dict[str, pd.DataFrame]:
    return io.load_snapshot_tables(snapshot_dir)


def _round_order(round_name: str) -> int:
    return bracket.round_order(round_name)


def _sigmoid(x: float) -> float:
    return bracket.sigmoid(x)


def _win_prob(
    net1: float,
    net2: float,
    scale: float,
) -> float:
    return bracket.win_prob(net1, net2, scale)


def _prepare_bracket_graph(
    games: pd.DataFrame,
) -> Tuple[pd.DataFrame, Dict[str, Dict[int, str]]]:
    return bracket.prepare_bracket_graph(games)


def _expected_simulation(
    tables: Dict[str, pd.DataFrame],
    calcutta_key: str,
    market_bids: pd.DataFrame,
    sim_rows: pd.DataFrame,
    sim_entry_key: str,
    points_mode: str,
    points_by_round: Optional[Dict[int, float]],
    n_sims: int,
    seed: int,
    kenpom_scale: float,
    budget: float,
    use_historical_winners: bool,
    competitor_entry_keys: Optional[List[str]] = None,
) -> Dict[str, object]:
    return sim.expected_simulation(
        tables=tables,
        calcutta_key=calcutta_key,
        market_bids=market_bids,
        sim_rows=sim_rows,
        sim_entry_key=sim_entry_key,
        points_mode=points_mode,
        points_by_round=points_by_round,
        n_sims=n_sims,
        seed=seed,
        kenpom_scale=kenpom_scale,
        budget=budget,
        use_historical_winners=use_historical_winners,
        competitor_entry_keys=competitor_entry_keys,
    )


def _simulate_team_points_scenarios(
    tables: Dict[str, pd.DataFrame],
    calcutta_key: str,
    points_mode: str,
    points_by_round: Optional[Dict[int, float]],
    n_sims: int,
    seed: int,
    kenpom_scale: float,
    use_historical_winners: bool,
) -> Tuple[List[str], np.ndarray]:
    return sim.simulate_team_points_scenarios(
        tables=tables,
        calcutta_key=calcutta_key,
        points_mode=points_mode,
        points_by_round=points_by_round,
        n_sims=n_sims,
        seed=seed,
        kenpom_scale=kenpom_scale,
        use_historical_winners=use_historical_winners,
    )


def _ensure_entry_points_include_competitors(
    entry_points: pd.DataFrame,
    competitor_entry_keys: List[str],
) -> pd.DataFrame:
    return standings.ensure_entry_points_include_competitors(
        entry_points,
        competitor_entry_keys,
    )


def _team_points_fixed(progress: int) -> float:
    return points.team_points_fixed(progress)


def _team_points_from_round_scoring(
    progress: int,
    points_by_round: Dict[int, float],
) -> float:
    return points.team_points_from_round_scoring(progress, points_by_round)


def _build_points_by_team(
    teams: pd.DataFrame,
    calcutta_key: str,
    points_mode: str,
    round_scoring: Optional[pd.DataFrame],
) -> pd.DataFrame:
    return points.build_points_by_team(
        teams=teams,
        calcutta_key=calcutta_key,
        points_mode=points_mode,
        round_scoring=round_scoring,
    )


def _compute_entry_points(
    entry_bids: pd.DataFrame,
    points_by_team: pd.DataFrame,
    calcutta_key: str,
) -> pd.DataFrame:
    return standings.compute_entry_points(
        entry_bids=entry_bids,
        points_by_team=points_by_team,
        calcutta_key=calcutta_key,
    )


def _compute_finish_positions_and_payouts(
    entry_points: pd.DataFrame,
    payouts: pd.DataFrame,
    calcutta_key: str,
    epsilon: float = 0.0001,
) -> pd.DataFrame:
    return standings.compute_finish_positions_and_payouts(
        entry_points=entry_points,
        payouts=payouts,
        calcutta_key=calcutta_key,
        epsilon=epsilon,
    )


def main() -> int:
    from calcutta_ds.cli_backtest_scaffold import main as cli_main

    return int(cli_main())


if __name__ == "__main__":
    raise SystemExit(main())
