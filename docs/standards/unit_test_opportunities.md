# Unit test opportunities

- [x] Add unit tests for `backend/internal/auth/jwt.go` (`TokenManager`): issue/verify roundtrip with fixed `now`, invalid format, signature mismatch, expired token, required args, invalid manager configuration
- [x] Add unit test for `backend/internal/auth/refresh.go` `HashRefreshToken` (deterministic hash for known input)
- [x] Refactor `backend/internal/auth/refresh.go` + `backend/internal/auth/api_key.go` to allow injecting randomness (e.g. `io.Reader`) so token generation can be unit tested deterministically
- [x] Add unit tests for `backend/internal/platform/config.go` beyond JWT secret defaults: database URL construction from `DB_*`, TTL parsing, default port/origin behavior
- [x] Add unit tests for numeric conversion helpers in `backend/internal/adapters/db/ml_analytics_repository.go` (`floatPtrFromPgNumeric`, `floatFromPgNumeric`)
- [x] Extract pure ownership math from `backend/internal/app/calcutta/calcutta_service_ownership.go` so portfolio-team ownership calculation is testable without repository calls
- [x] Add unit tests for DTO mappers (pure mapping) in `backend/internal/transport/httpserver/dtos/mappers_analytics.go` (e.g. `ToAnalyticsResponse` handles nil input and maps fields correctly)

- [x] Add unit tests for `backend/internal/transport/httpserver/http_helpers.go` (`writeError` status/code mapping)
- [x] Add unit tests for `backend/internal/app/bracket/builder.go` pure helper methods (e.g. matching/ordering, First Four identification)
- [x] Add unit tests for calcutta payout calculation logic in `backend/internal/app/calcutta/calcutta_service_accessors.go`


