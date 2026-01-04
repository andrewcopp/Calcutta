# 12-Factor Readiness Plan (Full-Stack)

## Goals

- Make the system runnable via the same build artifacts across environments (dev/staging/prod)
- Centralize runtime configuration in environment variables and load once at process start
- Ensure processes are stateless, disposable, and horizontally scalable
- Improve dev/prod parity and operational ergonomics (health checks, one-off tasks, logging)

## Principles (12-Factor aligned)

- **Config**: All runtime config comes from env vars; no runtime env-file coupling in production
- **Build / Release / Run**: Build artifacts are immutable; runtime behavior changes only via env
- **Backing services**: DB and other services are addressed via URLs and are swappable
- **Logs**: Write logs to stdout/stderr as an event stream
- **Disposability**: Fast startup, graceful shutdown, clean readiness signaling

## Work Plan

### Backend (Go)

- [ ] Define and document the canonical backend env var contract (required vs optional)
- [x] Ensure config is loaded once at startup and passed down (no per-request `os.Getenv`)
- [x] Make CORS configuration fully config-driven (prefer `ALLOWED_ORIGINS`)
- [ ] Standardize on `DATABASE_URL` (keep DB_* only as a dev convenience if needed)
- [x] Add `GET /healthz` and `GET /readyz` aliases (keep existing routes for compatibility)
- [ ] Ensure graceful shutdown covers:
  - [ ] HTTP server shutdown with timeout
  - [ ] DB pool close
- [ ] Add request/trace correlation:
  - [x] Preserve inbound `X-Request-ID` and always respond with it
  - [x] Ensure it’s included in all request logs
- [x] Standardize structured logging format (JSON or consistent key/value)
- [ ] Confirm background worker command(s) share the same config contract

### Frontend (Vite/React)

- [x] Confirm the canonical frontend env var contract (`VITE_API_URL` preferred)
- [x] Remove implicit fallback to `http://localhost:8080` in non-dev builds (fail fast or require explicit env)
- [ ] Document how frontend discovers backend in:
  - [ ] local dev
  - [ ] docker-compose
  - [ ] deployed environment
- [ ] Decide whether frontend needs runtime config injection (separate from build-time Vite env)

### Deployment model (frontend)

- [x] Decide frontend deployment model: S3 + CloudFront (static SPA)
- [ ] CloudFront behaviors: route `/api/*` to backend (ALB/ECS) and `/*` to S3

### Docker Compose / Local Dev Parity

- [ ] Ensure docker-compose env vars match the documented env contracts
- [ ] Avoid “install deps on container start” patterns for non-dev images (keep only for dev)
- [ ] Add/confirm container health checks (backend `ready` endpoint already exists)
- [ ] Verify `migrate` is a true one-off process (no coupling to web boot)
- [x] Add a local prod-like compose path (built artifacts) via `docker-compose.local-prod.yml`
- [x] Backend Dockerfile supports separate `dev` and `prod` targets

### CI / Deploy / Ops

- [ ] Confirm build artifacts are reproducible and versioned (commit SHA injected into backend binary)
- [ ] Ensure migrations run as an explicit release step in deploy workflow
- [ ] Confirm secrets are never committed and are injected via environment/secrets manager
- [ ] Document the minimal production env var set

### Data-Science / Pipelines (Non-web processes)

- [ ] Treat batch jobs/CLI tools as one-off processes with the same env contract style
- [ ] Ensure artifacts are written to explicit destinations configured via env (no hidden relative paths)
- [ ] Confirm logs are stdout/stderr and that jobs fail fast on missing config

## Execution Order (recommended)

- [x] Backend: config-loaded-once enforcement (CORS first)
- [ ] Backend: health/ready endpoint polish and logging consistency
- [x] Frontend: API URL contract tightening
- [ ] Compose/CI: align env contracts and one-off processes
