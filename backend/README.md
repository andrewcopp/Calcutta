# Backend
Go server for March Madness Investment Pool

## Structure

- `cmd/api`: HTTP API server
- `cmd/migrate`: migration runner
- `cmd/tools/*`: one-off operational tools
- `cmd/workers`: background workers (placeholder)

See `cmd/README.md` for the full list of binaries and per-command docs.

Core code lives under `internal/`.

## Running

### With Docker Compose

From the repo root:

```bash
make up
make ops-migrate
```

### Locally

From the repo root:

```bash
go run ./backend/cmd/api
```

## Tests

From the repo root:

```bash
make backend-test
```

## sqlc

This repo uses sqlc to generate typed query wrappers.

From the repo root:

```bash
make sqlc-generate
```

Conventions:
- Generated files live under `internal/adapters/db/sqlc`.
- Do not hand-edit generated `.sql.go` files; change the `.sql` query definitions and regenerate.

## Transactions

Repository methods that mutate multiple tables should:
- create a transaction (`BeginTx`)
- use sqlc `WithTx(tx)` for all queries
- commit on success
- rollback on error