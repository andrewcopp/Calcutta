# Engineering Standards

This repo follows simple, practical rules that keep code clear and change-safe without heavy process.

## Layering and dependencies
- **Dependency direction**: cmd/* (handlers/CLIs) → services → ports (interfaces) → adapters (implementations)
- **No upward imports**: lower layers never import higher layers.
- **Contracts at boundaries**: anything crossing boundaries uses typed structs/interfaces, not loose maps.

## Handlers: parse → service → response
- Handlers should:
  - parse inputs and validate
  - call a service method
  - map domain results to a response DTO and serialize
- Keep business logic out of handlers.

## Services: narrow interfaces, deterministic behavior
- Depend on small interfaces from `internal/ports` (1–3 method groups).
- Return errors; do not log internally. Logging happens at boundaries (handlers/CLIs).
- Prefer composition and strategy maps over large switch statements.

## Repositories (adapters)
- Concrete DB code lives in adapters (initially existing in `pkg/services/*_repository.go`; will move under `internal/adapters/postgres`).
- Implement `internal/ports` interfaces.

## Testing
- Unit tests only for now. No DB/integration tests.
- Unit tests should cover pure business logic (deterministic inputs → outputs).
- Do not write unit tests that assert filesystem/DB/HTTP behavior.
- Unit tests must not read/write external state (DB, network/API calls, filesystem, or any external storage).
- Prefer a “pure core + thin runner” structure:
  - Core functions: no I/O, return values.
  - Runners/handlers: load/save/log, call core, translate errors.
- If you need a dependency for unit tests, inject an interface and use a fake/in-memory implementation.
- Structure tests as GIVEN / WHEN / THEN with one reason to fail.
- Determinism: fix time, randomness, and sort before comparing.

## Function/file size guidelines
- Functions: aim for 20–60 LOC; orchestration may be longer if mostly wiring.
- Complexity: prefer strategy maps/state machines over deep branching.
- Files: keep under ~200–400 LOC; split by responsibility (types/logic/formatting) or by feature slice.

## Error handling and logging
- Wrap errors with context at the layer that understands it.
- Log only at boundaries (HTTP handlers/CLIs); inner layers return errors.

## Naming and layout
- Prefer verb-based interface names (Fetcher, Calculator, Formatter) and capability-scoped packages.
- Organize by responsibility: `internal/ports`, `internal/adapters`, `internal/services`, `cmd/server`, `pkg/models`.

## Migration notes for this repo
- Introduce `internal/ports/*` interfaces and have services depend on them.
- Keep current repository implementations where they are temporarily; move to `internal/adapters/postgres` in a later pass.
- Gradually update handlers to delegate to services and use request/response DTOs.
