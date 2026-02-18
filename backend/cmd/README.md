# Commands

Runnable binaries live under `backend/cmd/*`.

## Binaries

- `cmd/api`
  - HTTP API server.
- `cmd/migrate`
  - Migration runner.
- `cmd/workers`
  - Background worker system (bundle import worker, lab pipeline worker).
- `cmd/tools/export-bundles`
  - Exports data bundles for backup/transfer.
- `cmd/tools/import-bundles`
  - Imports data bundles.
- `cmd/tools/verify-bundles`
  - Verifies data bundle integrity.

Each command directory has a `README.md` describing how to run it, configuration, and side effects.
