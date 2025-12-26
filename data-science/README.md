# Data Science

Small Python utilities used to ingest/export snapshots and generate canonical datasets.

## Setup

```bash
python -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
```

## Scripts

- `ingest_snapshot.py`
  - Ingest a snapshot export from the backend admin tools.
- `derive_canonical.py`
  - Build a canonical dataset from ingested files.

## Notes

- This folder intentionally stays lightweight.
- For more context, see `docs/data_science_sandbox.md`.
