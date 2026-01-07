import json
import os
import subprocess
import sys
import time
from dataclasses import dataclass
from datetime import datetime, timedelta, timezone
from typing import Any, Dict, Optional, Tuple

import psycopg2


@dataclass
class Job:
    id: str
    run_id: str
    run_key: str
    params_json: Dict[str, Any]


def _utc_now() -> datetime:
    return datetime.now(timezone.utc)


def _env_int(name: str, default: int) -> int:
    try:
        v = int(str(os.getenv(name, str(default))).strip())
        return v
    except Exception:
        return default


def _env_seconds(name: str, default_seconds: int) -> int:
    return _env_int(name, default_seconds)


def _connect() -> psycopg2.extensions.connection:
    # Prefer DB_* (matches moneyball.db.connection expectations)
    host = os.getenv("DB_HOST", "db").strip() or "db"
    port = int(os.getenv("DB_PORT", "5432"))
    dbname = os.getenv("DB_NAME", "calcutta").strip() or "calcutta"
    user = os.getenv("DB_USER", "calcutta").strip() or "calcutta"
    password = os.getenv("DB_PASSWORD", "").strip() or "calcutta"

    return psycopg2.connect(
        host=host,
        port=port,
        dbname=dbname,
        user=user,
        password=password,
    )


def _claim_next_job(
    conn: psycopg2.extensions.connection,
    *,
    worker_id: str,
    stale_before: datetime,
) -> Optional[Job]:
    with conn.cursor() as cur:
        cur.execute(
            """
            WITH candidate AS (
                SELECT id
                FROM derived.run_jobs
                WHERE run_kind = 'market_share'
                  AND (
                        status = 'queued'
                     OR (
                        status = 'running'
                        AND claimed_at IS NOT NULL
                        AND claimed_at < %s
                     )
                  )
                ORDER BY created_at ASC
                LIMIT 1
                FOR UPDATE SKIP LOCKED
            )
            UPDATE derived.run_jobs j
            SET status = 'running',
                attempt = j.attempt + 1,
                claimed_at = %s,
                claimed_by = %s,
                started_at =
                    COALESCE(
                    j.started_at,
                    %s
                ),
                finished_at = NULL,
                error_message = NULL,
                updated_at = NOW()
            FROM candidate
            WHERE j.id = candidate.id
            RETURNING
                j.id::text,
                j.run_id::text,
                j.run_key::text,
                j.params_json
            """,
            (
                stale_before,
                _utc_now(),
                worker_id,
                _utc_now(),
            ),
        )
        row = cur.fetchone()
        if not row:
            return None
        job_id, run_id, run_key, params_json = row
        if params_json is None:
            params_json = {}
        return Job(
            id=str(job_id),
            run_id=str(run_id),
            run_key=str(run_key),
            params_json=dict(params_json),
        )


def _upsert_run_artifact_metrics(
    conn: psycopg2.extensions.connection,
    *,
    run_id: str,
    run_key: Optional[str],
    summary: Dict[str, Any],
) -> None:
    summary_json = json.dumps(summary)
    with conn.cursor() as cur:
        cur.execute(
            """
            INSERT INTO derived.run_artifacts (
                run_kind,
                run_id,
                run_key,
                artifact_kind,
                schema_version,
                storage_uri,
                summary_json
            )
            VALUES (
                'market_share',
                %s::uuid,
                %s::uuid,
                'metrics',
                'v1',
                NULL,
                %s::jsonb
            )
            ON CONFLICT (run_kind, run_id, artifact_kind)
                WHERE deleted_at IS NULL
            DO UPDATE
            SET run_key = EXCLUDED.run_key,
                schema_version = EXCLUDED.schema_version,
                storage_uri = EXCLUDED.storage_uri,
                summary_json = EXCLUDED.summary_json,
                updated_at = NOW(),
                deleted_at = NULL
            """,
            (
                run_id,
                run_key,
                summary_json,
            ),
        )


def _mark_job_succeeded(
    conn: psycopg2.extensions.connection,
    *,
    run_id: str,
) -> None:
    with conn.cursor() as cur:
        cur.execute(
            """
            UPDATE derived.run_jobs
            SET status = 'succeeded',
                finished_at = NOW(),
                error_message = NULL,
                updated_at = NOW()
            WHERE run_kind = 'market_share'
              AND run_id = %s::uuid
            """,
            (run_id,),
        )


def _mark_job_failed(
    conn: psycopg2.extensions.connection,
    *,
    run_id: str,
    error_message: str,
) -> None:
    with conn.cursor() as cur:
        cur.execute(
            """
            UPDATE derived.run_jobs
            SET status = 'failed',
                finished_at = NOW(),
                error_message = %s,
                updated_at = NOW()
            WHERE run_kind = 'market_share'
              AND run_id = %s::uuid
            """,
            (error_message, run_id),
        )


def _run_market_share_runner(
    *,
    run_id: str,
) -> Tuple[bool, Dict[str, Any], str]:
    python_bin = os.getenv("PYTHON_BIN", "python3").strip() or "python3"
    runner = os.getenv("PYTHON_MARKET_SHARE_RUNNER", "").strip()
    if not runner:
        runner = os.path.join(
            os.path.dirname(__file__),
            "run_market_share_runner.py",
        )

    start = time.time()
    try:
        cp = subprocess.run(
            [python_bin, runner, "--run-id", str(run_id)],
            capture_output=True,
            text=True,
            env=os.environ.copy(),
        )
    except Exception as e:
        return False, {}, str(e)

    dur_ms = int((time.time() - start) * 1000)
    stdout = ((cp.stdout or "").strip() or "{}")
    stderr = (cp.stderr or "").strip()

    parsed: Dict[str, Any] = {}
    try:
        parsed = json.loads(stdout)
    except Exception:
        parsed = {}

    if cp.returncode != 0:
        msg = (
            parsed.get("error")
            or stderr
            or f"python runner failed (exit={cp.returncode})"
        )
        return (
            False,
            {"durationMs": dur_ms, "stdout": stdout, "stderr": stderr},
            str(msg),
        )

    if not bool(parsed.get("ok")):
        msg = parsed.get("error") or "python runner returned ok=false"
        return (
            False,
            {"durationMs": dur_ms, "stdout": stdout, "stderr": stderr},
            str(msg),
        )

    return True, {"durationMs": dur_ms, "result": parsed}, ""


def main() -> int:
    poll_interval = _env_seconds("MARKET_SHARE_POLL_INTERVAL_SECONDS", 2)
    stale_after_seconds = _env_seconds(
        "MARKET_SHARE_STALE_AFTER_SECONDS",
        30 * 60,
    )

    worker_id = os.getenv("HOSTNAME", "market-share-python-worker").strip()
    if not worker_id:
        worker_id = "market-share-python-worker"

    while True:
        conn = None
        try:
            conn = _connect()
            conn.autocommit = False

            stale_before = _utc_now() - timedelta(
                seconds=stale_after_seconds
            )

            with conn:
                job = _claim_next_job(
                    conn,
                    worker_id=worker_id,
                    stale_before=stale_before,
                )

            if job is None:
                time.sleep(max(1, poll_interval))
                continue

            ok, meta, err = _run_market_share_runner(run_id=job.run_id)

            # Always include params in summary, since UI expects it sometimes.
            summary: Dict[str, Any] = {
                "runId": job.run_id,
                "runKey": job.run_key,
                "durationMs": int(
                    meta.get("durationMs")
                    or 0
                ),
                "params": job.params_json,
            }

            if ok:
                runner_result = meta.get("result") or {}
                rows_inserted = runner_result.get("rows_inserted")
                if rows_inserted is None:
                    rows_inserted = runner_result.get("RowsInserted")
                summary["status"] = "succeeded"
                if rows_inserted is not None:
                    summary["rowsInserted"] = rows_inserted

                with conn:
                    _mark_job_succeeded(conn, run_id=job.run_id)
                    _upsert_run_artifact_metrics(
                        conn,
                        run_id=job.run_id,
                        run_key=job.run_key,
                        summary=summary,
                    )

            else:
                summary["status"] = "failed"
                summary["errorMessage"] = err

                with conn:
                    _mark_job_failed(
                        conn,
                        run_id=job.run_id,
                        error_message=err,
                    )
                    _upsert_run_artifact_metrics(
                        conn,
                        run_id=job.run_id,
                        run_key=job.run_key,
                        summary=summary,
                    )

        except Exception as e:
            # Best effort: avoid crashing the worker loop.
            sys.stderr.write(f"market_share_python_worker error: {e}\n")
            time.sleep(max(1, poll_interval))
        finally:
            try:
                if conn is not None:
                    conn.close()
            except Exception:
                pass


if __name__ == "__main__":
    raise SystemExit(main())
