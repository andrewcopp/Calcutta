import argparse
import json
import os
import sys
import zipfile
from pathlib import Path

import pandas as pd


def _read_manifest(zf: zipfile.ZipFile):
    try:
        with zf.open("manifest.json") as f:
            return json.loads(f.read().decode("utf-8"))
    except KeyError:
        return None


def _csv_members(zf: zipfile.ZipFile):
    for name in zf.namelist():
        if name.endswith("/"):
            continue
        base = os.path.basename(name)
        if base.lower().endswith(".csv"):
            yield name


def main() -> int:
    parser = argparse.ArgumentParser(
        description=(
            "Convert Calcutta analytics snapshot zip (CSVs) "
            "into Parquet-per-table"
        )
    )
    parser.add_argument(
        "zip_path",
        help="Path to downloaded analytics snapshot zip",
    )
    parser.add_argument(
        "--out",
        dest="out_dir",
        default="./out",
        help="Output directory for parquet files",
    )
    args = parser.parse_args()

    zip_path = Path(args.zip_path)
    if not zip_path.exists():
        print(f"zip not found: {zip_path}", file=sys.stderr)
        return 2

    out_dir = Path(args.out_dir)
    out_dir.mkdir(parents=True, exist_ok=True)

    with zipfile.ZipFile(zip_path, "r") as zf:
        manifest = _read_manifest(zf)
        if manifest is not None:
            (out_dir / "manifest.json").write_text(
                json.dumps(manifest, indent=2) + "\n",
                encoding="utf-8",
            )

        csv_names = list(_csv_members(zf))
        if not csv_names:
            print("no csv files found in zip", file=sys.stderr)
            return 3

        for name in csv_names:
            table_name = Path(os.path.basename(name)).stem
            with zf.open(name) as f:
                df = pd.read_csv(f)

            parquet_path = out_dir / f"{table_name}.parquet"
            df.to_parquet(parquet_path, index=False)

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
