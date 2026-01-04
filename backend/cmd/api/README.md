# API

HTTP API server.

## Run

From the repo root:

```bash
go run ./backend/cmd/api
```

## Configuration

Uses `backend/internal/platform.LoadConfigFromEnv()`.

- `DATABASE_URL` (preferred)
- Or: `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_HOST`, `DB_PORT`, `DB_SSLMODE`
- `PORT` (default: `8080`)
- `ALLOWED_ORIGIN` (default: `http://localhost:3000`)

Auth:

- `AUTH_MODE` (`legacy` | `cognito` | `dev`; default: `legacy`)
- If `AUTH_MODE != cognito`: `JWT_SECRET` (defaults to `dev-jwt-secret` when `NODE_ENV=development`)
- If `AUTH_MODE = cognito`: `COGNITO_REGION`, `COGNITO_USER_POOL_ID`, `COGNITO_APP_CLIENT_ID`

Token TTLs:

- `ACCESS_TOKEN_TTL_SECONDS` (default: `900`)
- `REFRESH_TOKEN_TTL_HOURS` (default: `720`)
