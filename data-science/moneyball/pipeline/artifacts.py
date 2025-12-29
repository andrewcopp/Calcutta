from __future__ import annotations

import hashlib
import json
from dataclasses import asdict, dataclass
from datetime import datetime, timezone
from pathlib import Path
from typing import Any, Dict, Optional


@dataclass(frozen=True)
class FileFingerprint:
    path: str
    size: int
    mtime_ns: int


def fingerprint_file(p: Path) -> FileFingerprint:
    st = p.stat()
    return FileFingerprint(
        path=str(p),
        size=int(st.st_size),
        mtime_ns=int(st.st_mtime_ns),
    )


def sha256_jsonable(obj: Any) -> str:
    payload = json.dumps(obj, sort_keys=True, separators=(",", ":")).encode(
        "utf-8"
    )
    return hashlib.sha256(payload).hexdigest()


def utc_now_iso() -> str:
    return datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")


def ensure_dir(p: Path) -> None:
    p.mkdir(parents=True, exist_ok=True)


def write_json(p: Path, payload: Dict[str, Any]) -> None:
    ensure_dir(p.parent)
    p.write_text(json.dumps(payload, indent=2) + "\n", encoding="utf-8")


def read_json(p: Path) -> Dict[str, Any]:
    return json.loads(p.read_text(encoding="utf-8"))


def default_artifacts_root(snapshot_dir: Path) -> Path:
    return snapshot_dir / "derived" / "moneyball"


def build_run_dir(
    *,
    artifacts_root: Path,
    snapshot_name: str,
    run_id: str,
) -> Path:
    return artifacts_root / str(snapshot_name) / str(run_id)


def load_manifest(manifest_path: Path) -> Optional[Dict[str, Any]]:
    if not manifest_path.exists():
        return None
    try:
        return read_json(manifest_path)
    except Exception:
        return None


def manifest_matches(
    *,
    existing: Dict[str, Any],
    stage: str,
    stage_config: Dict[str, Any],
    input_fingerprints: Dict[str, FileFingerprint],
) -> bool:
    stages = existing.get("stages")
    if not isinstance(stages, dict):
        return False
    s = stages.get(stage)
    if not isinstance(s, dict):
        return False

    if s.get("stage_config_hash") != sha256_jsonable(stage_config):
        return False

    existing_inputs = s.get("inputs")
    if not isinstance(existing_inputs, dict):
        return False

    for k, fp in input_fingerprints.items():
        e = existing_inputs.get(k)
        if not isinstance(e, dict):
            return False
        if int(e.get("size", -1)) != int(fp.size):
            return False
        if int(e.get("mtime_ns", -1)) != int(fp.mtime_ns):
            return False

    return True


def fingerprints_to_dict(
    fps: Dict[str, FileFingerprint],
) -> Dict[str, Dict[str, Any]]:
    return {k: asdict(v) for k, v in fps.items()}
