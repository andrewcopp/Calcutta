# Generate Predicted Game Outcomes

Generates and writes `predicted_game_outcomes` for a tournament season.

## Run

From the repo root:

```bash
go run ./backend/cmd/generate-predicted-game-outcomes --season 2025
```

## Flags

- `--season` (required) Tournament season/year (e.g. `2025`)
- `--n-sims` (default: `5000`) Number of Monte Carlo simulations used to estimate matchup probabilities
- `--seed` (default: `42`) Base RNG seed
- `--kenpom-scale` (default: `10.0`) KenPom scale parameter
- `--model-version` (default: `kenpom-v1-go`) Stored on `derived.predicted_game_outcomes`

## Configuration

Uses `backend/internal/platform.LoadConfigFromEnv()`.

- `DATABASE_URL` (preferred)
- Or: `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_HOST`, `DB_PORT`, `DB_SSLMODE`

Note: config loading also enforces auth env vars (e.g. `JWT_SECRET` when `AUTH_MODE != cognito`).

## Side effects

Writes predicted game outcome rows to Postgres.
