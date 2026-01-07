# internal/app

This package subtree is the backend **composition root**.

## What belongs here

- Application wiring (constructing services, repositories, and shared dependencies)
- Bootstrap/config loading and dependency graph assembly
- The top-level `App` struct used by transports/CLIs

## What does NOT belong here

- New feature logic
- New reusable domain logic
- New service implementations intended for consumption by other packages

## Import conventions (Option A)

- **Outside of `internal/app/**`**, code should depend directly on feature packages under:
  - `backend/internal/app/<feature>`

- `backend/internal/features/*` is deprecated and should not be used.

### Examples

Preferred:

- `github.com/andrewcopp/Calcutta/backend/internal/app/analytics`
- `github.com/andrewcopp/Calcutta/backend/internal/app/auth`

Avoid:

- `github.com/andrewcopp/Calcutta/backend/internal/features/analytics`
- `github.com/andrewcopp/Calcutta/backend/internal/features/auth`
