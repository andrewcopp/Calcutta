# Tools

One-off operational tools.

## Tools

- `tools/export-bundles`
  - Exports bundle archives from the database to a local directory.
- `tools/import-bundles`
  - Imports bundle archives from a local directory into the database.
- `tools/verify-bundles`
  - Verifies that bundle archives on disk match the current database state.
- `tools/retain-simulation-runs`
  - Deletes old simulation runs and their associated artifacts (DB rows + filesystem `file://` artifacts).

Each tool directory has a `README.md` describing how to run it, configuration, and side effects.
