import argparse
import json
import os
import sys
import tempfile
import urllib.parse
import urllib.request
import zipfile
from pathlib import Path
from typing import Optional

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


def _download_snapshot_zip(
    base_url: str,
    tournament_id: str,
    calcutta_id: str,
    api_key: str,
    out_path: Path,
):
    q = urllib.parse.urlencode(
        {"tournamentId": tournament_id, "calcuttaId": calcutta_id}
    )
    url = base_url.rstrip("/") + "/api/admin/analytics/export?" + q

    req = urllib.request.Request(url)
    req.add_header("Authorization", f"Bearer {api_key}")

    with urllib.request.urlopen(req, timeout=60) as resp:
        if resp.status != 200:
            raise RuntimeError(
                f"download failed: {resp.status} {resp.reason}"
            )
        out_path.write_bytes(resp.read())


def main() -> int:
    parser = argparse.ArgumentParser(
        description=(
            "Convert Calcutta analytics snapshot zip (CSVs) "
            "into Parquet-per-table"
        )
    )
    parser.add_argument(
        "zip_path",
        nargs="?",
        help="Path to downloaded analytics snapshot zip",
    )
    parser.add_argument(
        "--base-url",
        dest="base_url",
        help="Base URL for API, e.g. http://localhost:8080",
    )
    parser.add_argument(
        "--tournament-id",
        dest="tournament_id",
        help="Tournament ID for snapshot export",
    )
    parser.add_argument(
        "--calcutta-id",
        dest="calcutta_id",
        help="Calcutta ID for snapshot export",
    )
    parser.add_argument(
        "--api-key",
        dest="api_key",
        default=os.getenv("CALCUTTA_API_KEY"),
        help="API key for Authorization: Bearer (or set CALCUTTA_API_KEY)",
    )
    parser.add_argument(
        "--out",
        dest="out_dir",
        default="./out",
        help="Output directory for parquet files",
    )
    args = parser.parse_args()

    out_dir = Path(args.out_dir)
    out_dir.mkdir(parents=True, exist_ok=True)

    zip_path: Optional[Path]
    if args.zip_path:
        zip_path = Path(args.zip_path)
        if not zip_path.exists():
            print(f"zip not found: {zip_path}", file=sys.stderr)
            return 2
    else:
        if not args.base_url:
            print(
                "--base-url is required when zip_path is omitted",
                file=sys.stderr,
            )
            return 2
        if not args.tournament_id:
            print(
                "--tournament-id is required when zip_path is omitted",
                file=sys.stderr,
            )
            return 2
        if not args.calcutta_id:
            print(
                "--calcutta-id is required when zip_path is omitted",
                file=sys.stderr,
            )
            return 2
        if not args.api_key:
            print(
                "--api-key (or CALCUTTA_API_KEY) is required when zip_path is "
                "omitted",
                file=sys.stderr,
            )
            return 2

        tmp_dir = Path(tempfile.mkdtemp(prefix="calcutta-analytics-"))
        zip_path = tmp_dir / "analytics_snapshot.zip"
        try:
            _download_snapshot_zip(
                args.base_url,
                args.tournament_id,
                args.calcutta_id,
                args.api_key,
                zip_path,
            )
        except Exception as e:
            print(str(e), file=sys.stderr)
            return 4

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
