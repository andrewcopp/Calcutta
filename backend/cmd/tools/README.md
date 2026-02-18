# Backend CLI Tools

Operational tools for bundle management.

## Quick Reference

```bash
# Import historical data (schools, tournaments, calcuttas)
make import-bundles

# Run complete dry-run flow
make dry-run
```

## Tools

### import-bundles

Import schools, tournaments, and calcuttas from JSON bundles.

**Usage:**
```bash
make import-bundles
go run ./cmd/tools/import-bundles -in=./exports/bundles -dry-run=false
```

Uses database transaction (all-or-nothing import). Validates bundles before importing.

### export-bundles

Export schools, tournaments, and calcuttas to JSON bundles.

**Usage:**
```bash
go run ./cmd/tools/export-bundles -out=./exports/bundles -season=2025
```

Use for backups or creating seed data for new environments.

### verify-bundles

Verify bundle integrity without importing.

**Usage:**
```bash
go run ./cmd/tools/verify-bundles -in=./exports/bundles
```

Checks JSON schema, referential integrity, and required fields.

See individual tool directories for detailed README files.
