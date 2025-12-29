from __future__ import annotations

from pathlib import Path
from typing import Dict

import pandas as pd

from moneyball.utils import io


def load_snapshot_tables(snapshot_dir: Path) -> Dict[str, pd.DataFrame]:
    return io.load_snapshot_tables(snapshot_dir)
