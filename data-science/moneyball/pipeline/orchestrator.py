"""
Pipeline orchestrator for running stages in sequence.

Separates concerns:
- Orchestrator: Runs stages in order
- ArtifactStore: Manages paths and locations
- Stage functions: Just do the work
"""
from __future__ import annotations

from pathlib import Path
from typing import Any, Dict, List, Optional

from moneyball.pipeline.artifacts import (
    build_run_dir,
    default_artifacts_root,
    ensure_dir,
    utc_now_iso,
    write_json,
)


class ArtifactStore:
    """Manages artifact paths and locations."""

    def __init__(
        self,
        snapshot_dir: Path,
        snapshot_name: str,
        artifacts_root: Optional[Path] = None,
        run_id: Optional[str] = None,
    ):
        self.snapshot_dir = Path(snapshot_dir)
        self.snapshot_name = snapshot_name

        self.artifacts_root = (
            Path(artifacts_root)
            if artifacts_root is not None
            else default_artifacts_root(self.snapshot_dir)
        )

        self.run_id = (
            str(run_id)
            if run_id is not None
            else utc_now_iso().replace(":", "").replace("-", "")
        )

        self.run_dir = build_run_dir(
            artifacts_root=self.artifacts_root,
            snapshot_name=self.snapshot_name,
            run_id=self.run_id,
        )
        ensure_dir(self.run_dir)

        # Canonical artifact directory (no timestamp)
        self.canonical_dir = self.snapshot_dir / "derived"
        ensure_dir(self.canonical_dir)

    def get_canonical_path(self, artifact_name: str) -> Path:
        """Get path for canonical artifact (no timestamp)."""
        return self.canonical_dir / f"{artifact_name}.parquet"

    def get_canonical_manifest_path(self, artifact_name: str) -> Path:
        """Get manifest path for canonical artifact."""
        return self.canonical_dir / f"{artifact_name}_manifest.json"

    def get_versioned_path(self, artifact_name: str) -> Path:
        """Get path for versioned artifact (timestamped)."""
        return self.run_dir / f"{artifact_name}.parquet"

    def get_snapshot_path(self, filename: str) -> Path:
        """Get path to snapshot data file."""
        return self.snapshot_dir / filename


class PipelineOrchestrator:
    """Orchestrates pipeline stage execution."""

    def __init__(
        self,
        artifact_store: ArtifactStore,
        use_cache: bool = True,
    ):
        self.store = artifact_store
        self.use_cache = use_cache
        self.manifest: Dict[str, Any] = {
            "moneyball": {
                "snapshot_dir": str(self.store.snapshot_dir),
                "snapshot_name": self.store.snapshot_name,
                "run_id": self.store.run_id,
                "created_at": utc_now_iso(),
            },
            "stages": {},
        }
        self.results: Dict[str, Any] = {
            "run_dir": str(self.store.run_dir),
        }

    def run_stage(
        self,
        stage_name: str,
        stage_func: Any,
        **kwargs: Any,
    ) -> Path:
        """
        Run a single stage.

        Args:
            stage_name: Name of the stage
            stage_func: Stage function to call
            **kwargs: Additional arguments to pass to stage function

        Returns:
            Path to the output artifact
        """
        out_path, self.manifest = stage_func(
            snapshot_dir=self.store.snapshot_dir,
            out_dir=self.store.run_dir,
            use_cache=self.use_cache,
            manifest=self.manifest,
            **kwargs,
        )
        return out_path

    def finalize(self) -> Dict[str, Any]:
        """Write manifest and return results."""
        manifest_path = self.store.run_dir / "manifest.json"
        write_json(manifest_path, self.manifest)
        self.results["manifest"] = str(manifest_path)
        return self.results
