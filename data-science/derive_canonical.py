import argparse
from pathlib import Path

import pandas as pd


def _read_parquet(snapshot_dir: Path, name: str) -> pd.DataFrame:
    p = snapshot_dir / f"{name}.parquet"
    if not p.exists():
        raise FileNotFoundError(f"missing required file: {p}")
    return pd.read_parquet(p)


def _write_parquet(out_dir: Path, name: str, df: pd.DataFrame) -> Path:
    out_dir.mkdir(parents=True, exist_ok=True)
    p = out_dir / f"{name}.parquet"
    df.to_parquet(p, index=False)
    return p


def _build_team_features(teams: pd.DataFrame) -> pd.DataFrame:
    cols = [
        "tournament_key",
        "team_key",
        "school_slug",
        "school_name",
        "seed",
        "region",
        "byes",
        "wins",
        "eliminated",
        "kenpom_net",
        "kenpom_o",
        "kenpom_d",
        "kenpom_adj_t",
    ]
    missing = [c for c in cols if c not in teams.columns]
    if missing:
        raise ValueError(f"teams.parquet missing columns: {missing}")

    df = teams[cols].copy()

    for c in ["seed", "byes", "wins"]:
        df[c] = pd.to_numeric(df[c], errors="coerce")

    for c in ["kenpom_net", "kenpom_o", "kenpom_d", "kenpom_adj_t"]:
        df[c] = pd.to_numeric(df[c], errors="coerce")

    if "eliminated" in df.columns:
        df["eliminated"] = df["eliminated"].astype("bool")

    return df


def _build_team_market(entry_bids: pd.DataFrame) -> pd.DataFrame:
    cols = ["calcutta_key", "entry_key", "team_key", "bid_amount"]
    missing = [c for c in cols if c not in entry_bids.columns]
    if missing:
        raise ValueError(f"entry_bids.parquet missing columns: {missing}")

    df = entry_bids[cols].copy()
    df["bid_amount"] = (
        pd.to_numeric(df["bid_amount"], errors="coerce")
        .fillna(0.0)
    )

    grouped = (
        df.groupby(
            ["calcutta_key", "team_key"],
            dropna=False,
        )
        .agg(
            total_bid_amount=("bid_amount", "sum"),
            num_bids=("bid_amount", "size"),
            num_entries=("entry_key", "nunique"),
            min_bid_amount=("bid_amount", "min"),
            max_bid_amount=("bid_amount", "max"),
        )
        .reset_index()
    )

    def _avg_bid_amount(r) -> float:
        num_bids = r["num_bids"]
        if not num_bids:
            return 0.0
        return r["total_bid_amount"] / num_bids

    grouped["avg_bid_amount"] = grouped.apply(
        _avg_bid_amount,
        axis=1,
    )

    pool_sizes = (
        df.groupby(["calcutta_key"], dropna=False)
        .agg(pool_total_bid_amount=("bid_amount", "sum"))
        .reset_index()
    )
    res = grouped.merge(
        pool_sizes,
        on="calcutta_key",
        how="left",
    )
    res["team_share_of_pool"] = res.apply(
        lambda r: (r["total_bid_amount"] / r["pool_total_bid_amount"])
        if r["pool_total_bid_amount"]
        else 0.0,
        axis=1,
    )

    return res


def _build_team_dataset(
    team_features: pd.DataFrame,
    team_market: pd.DataFrame,
) -> pd.DataFrame:
    if "team_key" not in team_features.columns:
        raise ValueError("team_features missing team_key")
    if (
        "team_key" not in team_market.columns
        or "calcutta_key" not in team_market.columns
    ):
        raise ValueError("team_market missing required keys")

    df = team_market.merge(team_features, on="team_key", how="left")
    return df


def main() -> int:
    parser = argparse.ArgumentParser(
        description=(
            "Derive canonical modeling tables from an ingested "
            "analytics snapshot. "
            "Reads <snapshot_dir>/*.parquet and writes "
            "<snapshot_dir>/derived/*.parquet."
        )
    )
    parser.add_argument(
        "snapshot_dir",
        help="Path to the ingested snapshot directory (contains *.parquet)",
    )
    parser.add_argument(
        "--out",
        dest="out_dir",
        default=None,
        help="Override output directory (default: <snapshot_dir>/derived)",
    )

    args = parser.parse_args()

    snapshot_dir = Path(args.snapshot_dir)
    if not snapshot_dir.exists():
        raise FileNotFoundError(f"snapshot_dir not found: {snapshot_dir}")

    out_dir = (
        Path(args.out_dir)
        if args.out_dir
        else (snapshot_dir / "derived")
    )

    teams = _read_parquet(snapshot_dir, "teams")
    entry_bids = _read_parquet(snapshot_dir, "entry_bids")

    team_features = _build_team_features(teams)
    team_market = _build_team_market(entry_bids)
    team_dataset = _build_team_dataset(team_features, team_market)

    _write_parquet(out_dir, "team_features", team_features)
    _write_parquet(out_dir, "team_market", team_market)
    _write_parquet(out_dir, "team_dataset", team_dataset)

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
