# Backend CLI Tools

Operational tools for database seeding, admin tasks, and bundle management.

## Quick Reference

```bash
# Create admin user
make create-admin EMAIL=admin@example.com

# Import historical data (schools, tournaments, calcuttas)
make import-bundles

# Generate simulation data for ML features
make seed-simulations

# Run complete dry-run flow
make dry-run
```

## Tools

### create-admin (NEW)

Create an admin user in the database.

**Usage:**
```bash
make create-admin EMAIL=admin@example.com NAME="Admin User"
go run ./cmd/tools/create-admin -email=admin@example.com -name="Admin User"
go run ./cmd/tools/create-admin -email=admin@example.com -dry-run
```

Creates a new user with role=admin and status=active. Idempotent (safe to run multiple times).

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

### seed-simulations (NEW)

Generate game outcome predictions and tournament simulations for all seasons.

**Usage:**
```bash
make seed-simulations
make seed-simulations NSIMS=50000 SEED=123
go run ./cmd/tools/seed-simulations -n-sims=10000 -seed=42
```

Requires tournaments in database. Populates derived schema for ML features.

### retain-simulation-runs

Mark specific simulation runs for retention (prevents cleanup).

**Usage:**
```bash
go run ./cmd/tools/retain-simulation-runs -batch-id=abc123 -reason="baseline"
```

Use to preserve important simulation runs for ML experiments.

## Seeding Workflow for New Environment

```bash
make prod-up
make prod-ops-migrate
make import-bundles
make seed-simulations
make create-admin EMAIL=admin@example.com
make api-test ENDPOINT=/api/health
```

See individual tool directories for detailed README files.
