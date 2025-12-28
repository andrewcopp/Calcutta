from __future__ import annotations

from pathlib import Path
from typing import List, Optional

import pandas as pd


def zscore(s: pd.Series) -> pd.Series:
    v = pd.to_numeric(s, errors="coerce")
    mu = float(v.mean()) if v.notna().any() else 0.0
    sd = float(v.std(ddof=0)) if v.notna().any() else 0.0
    if sd <= 0:
        return v * 0.0
    return (v - mu) / sd


def parse_snapshot_year(name: str) -> Optional[int]:
    try:
        return int(str(name).strip())
    except Exception:
        return None


def choose_train_snapshots(
    *,
    all_snapshot_dirs: List[Path],
    predict_snapshot: str,
    train_mode: str,
) -> List[str]:
    if train_mode == "loo":
        return [
            s.name
            for s in all_snapshot_dirs
            if s.name != predict_snapshot
        ]

    if train_mode != "past_only":
        raise ValueError(f"unknown train_mode: {train_mode}")

    y_pred = parse_snapshot_year(predict_snapshot)
    if y_pred is None:
        return [
            s.name
            for s in all_snapshot_dirs
            if s.name != predict_snapshot
        ]

    out: List[str] = []
    for s in all_snapshot_dirs:
        if s.name == predict_snapshot:
            continue
        ys = parse_snapshot_year(s.name)
        if ys is None:
            continue
        if ys < y_pred:
            out.append(s.name)
    return sorted(out)
