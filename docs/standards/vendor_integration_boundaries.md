# Vendor Integration Boundaries

Quick reference for where third-party libraries are isolated, so future swaps touch minimal surface area.

## PostgreSQL / pgx

- **Role:** Primary data store
- **Seam:** Repository adapters in `backend/internal/adapters/db/`
- **Contract:** Port interfaces in `backend/internal/ports/` (e.g., `PoolRepository`, `PortfolioRepository`)
- **Swap scope:** Replace adapter implementations; services depend only on port interfaces
- **Known exception:** `calcutta_evaluations/resolver_data.go` uses raw `pgxpool.Pool` queries (predates port extraction)

## SQLC

- **Role:** Type-safe SQL code generation
- **Seam:** Generated code in `backend/internal/adapters/db/sqlc/`; queries in `sqlc/queries/*.sql`
- **Config:** `backend/sqlc.yaml`
- **Swap scope:** Replace generated package; adapters wrap SQLC types into domain models

## Sentry

- **Role:** Error tracking and performance monitoring
- **Seam:** Initialized in `backend/internal/platform/sentry.go`; middleware in `transport/httpserver/middleware/`
- **Usage:** `httperr.WriteFromErr` captures unhandled errors; middleware injects hub into request context
- **Swap scope:** Replace `platform/sentry.go` init + middleware + the capture call in `httperr.go`

## AWS Cognito

- **Role:** Production authentication (when `AUTH_MODE=cognito`)
- **Seam:** `backend/internal/auth/cognito_authenticator.go`
- **Contract:** `auth.Authenticator` interface used by `ChainAuthenticator`
- **Swap scope:** Implement a new `Authenticator`; no service-layer changes needed

## SMTP / Email

- **Role:** Transactional email (invitations, notifications)
- **Seam:** `backend/internal/platform/mailer.go`
- **Contract:** `Mailer` interface in `internal/ports/`
- **Swap scope:** Replace mailer implementation (e.g., SendGrid, SES)
