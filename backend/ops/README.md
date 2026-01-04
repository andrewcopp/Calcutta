# Backend Ops

This directory contains operational scripts intended to be run manually against a Postgres instance (typically via `psql`).

## Directory layout

- `db/admin/`
  - Role/user setup and privilege management.
- `db/maintenance/`
  - Destructive or semi-destructive maintenance tasks (e.g. truncating derived tables).
- `db/audits/`
  - Read-only sanity checks and verification queries.

## Running scripts

These scripts are designed to be run with `psql` and `ON_ERROR_STOP=1`.

Example:

```sh
psql "postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable" \
  -v ON_ERROR_STOP=1 \
  -f backend/ops/db/audits/core_sanity_checks.sql
```

## Scripts

### db/admin

- `db_roles_grants.sql`
  - Creates roles `app_writer` and `lab_pipeline` (idempotent), and applies schema/table privileges.

- `verify_db_roles_grants.sql`
  - Read-only checks that roles exist and key privileges are present.

### db/maintenance

- `reset_derived_data.sql`
  - Truncates derived/lab tables used for simulations/evaluations.
  - This is destructive; use only in dev environments.

- `freeze_invariants.sql`
  - Read-only “baseline” counts/orphan checks intended to be captured in a text output for later comparisons.

### db/audits

- `core_sanity_checks.sql`
  - Read-only sanity checks (counts + orphan checks) for core backfills/cutovers.

## Makefile integration

- `make reset-derived` runs `backend/ops/db/maintenance/reset_derived_data.sql`.
