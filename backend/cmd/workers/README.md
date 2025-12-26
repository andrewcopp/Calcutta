# Workers

This directory contains background worker binaries.

## Status

This is currently a placeholder entrypoint to reserve the structure and conventions.

## Running

```bash
go run ./cmd/workers
```

## Intent

Workers are intended for long-running async processing such as:
- importing bundle uploads
- recalculating portfolios
- generating analytics snapshots
- scheduled maintenance
