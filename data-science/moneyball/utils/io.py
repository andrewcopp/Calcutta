from pathlib import Path
from typing import Dict, Optional

import pandas as pd


def read_parquet(p: Path) -> pd.DataFrame:
    if not p.exists():
        raise FileNotFoundError(f"missing required file: {p}")
    return pd.read_parquet(p)


def choose_calcutta_key(df: pd.DataFrame, requested: Optional[str]) -> str:
    if "calcutta_key" not in df.columns:
        raise ValueError("team_dataset missing calcutta_key")

    keys = [k for k in sorted(df["calcutta_key"].dropna().unique())]
    if requested:
        if requested not in keys:
            raise ValueError(f"calcutta_key not found: {requested}")
        return requested

    if len(keys) != 1:
        raise ValueError(
            "multiple calcutta_key values found; pass --calcutta-key"
        )
    return str(keys[0])


def load_snapshot_tables(snapshot_dir: Path) -> Dict[str, pd.DataFrame]:
    tables: Dict[str, pd.DataFrame] = {}

    # Required
    tables["team_dataset"] = read_parquet(
        snapshot_dir / "derived" / "team_dataset.parquet"
    )
    tables["teams"] = read_parquet(snapshot_dir / "teams.parquet")
    tables["entries"] = read_parquet(snapshot_dir / "entries.parquet")
    tables["entry_bids"] = read_parquet(snapshot_dir / "entry_bids.parquet")
    tables["payouts"] = read_parquet(snapshot_dir / "payouts.parquet")

    # Optional
    round_scoring_path = snapshot_dir / "round_scoring.parquet"
    if round_scoring_path.exists():
        tables["round_scoring"] = read_parquet(round_scoring_path)

    games_path = snapshot_dir / "games.parquet"
    if games_path.exists():
        tables["games"] = read_parquet(games_path)

    return tables
