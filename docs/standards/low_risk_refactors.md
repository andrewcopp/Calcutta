# Low-risk refactors

## Checklist

- [x] Centralize `models.BracketRound` metadata (order / min-progress / reverse-index) in `backend/pkg/models` and update call sites to remove duplicated round `switch` ladders.
- [x] Extract a shared helper in `backend/internal/adapters/db/analytics_repository.go` to resolve strategy generation run selection (explicit run ID vs latest run vs legacy fallback).
- [x] Frontend: replace `switch`-based sorting in `frontend/src/pages/TournamentViewPage.tsx` with a comparator map keyed by `SortField`.
