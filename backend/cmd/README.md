# Commands

Runnable binaries live under `backend/cmd/*`.

## Binaries

- `cmd/api`
  - HTTP API server.
- `cmd/migrate`
  - Migration runner.
- `cmd/simulate-tournaments`
  - Monte Carlo tournament simulation writer.
- `cmd/generate-predicted-game-outcomes`
  - Generates and writes `predicted_game_outcomes`.
- `cmd/generate-recommended-entry-bids`
  - Runs the portfolio optimizer and writes recommended bids.
- `cmd/calculate-simulated-calcuttas`
  - Calculates simulated calcutta results.
- `cmd/evaluate-calcutta`
  - Runs calcutta evaluation using an existing simulation batch.
- `cmd/benchmark-tournament-simulations`
  - End-to-end benchmark for tournament simulation + DB writes.
- `cmd/tools/*`
  - One-off operational tools.
- `cmd/workers`
  - Background workers (placeholder).

Each command directory has a `README.md` describing how to run it, configuration, and side effects.
