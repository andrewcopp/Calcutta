# Commands

Runnable binaries live under `backend/cmd/*`.

## Binaries

- `cmd/api`
  - HTTP API server.
- `cmd/migrate`
  - Migration runner.
- `cmd/simulate-tournaments`
  - Monte Carlo tournament simulation writer.
- `cmd/evaluate-calcutta`
  - Runs calcutta evaluation using an existing simulation batch.
- `cmd/tools/create-admin`
  - Creates an admin user.
- `cmd/tools/export-bundles`
  - Exports data bundles for backup/transfer.
- `cmd/tools/import-bundles`
  - Imports data bundles.
- `cmd/tools/verify-bundles`
  - Verifies data bundle integrity.
- `cmd/tools/retain-simulation-runs`
  - Retains selected simulation runs and cleans up old ones.
- `cmd/tools/seed-simulations`
  - Seeds simulation runs for the worker to process.
- `cmd/workers`
  - Background worker system (simulation worker, game outcome worker).

Each command directory has a `README.md` describing how to run it, configuration, and side effects.
