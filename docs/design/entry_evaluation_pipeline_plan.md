# Entry Evaluation Pipeline Plan

## Goals

- Separate Python-owned artifacts from Go-owned derived computations.
- Python submits entry evaluation work via API (no direct writes to `core`/`derived`).
- Go runs `simulate tournaments + evaluate calcutta` as a single job.
- Python can only request `starting_state_key: "post_first_four"`.
- Exclude a real entry by `excluded_entry_name` for now.

## Status
- Completed: consolidated runtime schemas to `core` + `derived` via migration `20260104211500_collapse_lab_schemas_into_derived`.
- Completed: added `models` schema and Python-owned artifact tables via migration `20260104223000_add_models_schema`.
- Completed: added `derived.entry_evaluation_requests` queue table via migration `20260104223500_add_entry_evaluation_requests`.
- Completed: added request submission permission `analytics.entry_evaluation_requests.write` via migration `20260104224000_add_entry_evaluation_request_permission`.
- Completed: implemented `POST /api/entry-evaluation-requests` (permission-gated).
- Completed: added entry evaluation worker loop and wired it into `cmd/workers`.
- Completed: implemented `starting_state_key='post_first_four'` behavior for simulations.

## Data ownership

- Python writes: `models.*`
- Go writes: `core.*`, `derived.*`
- Python calls API to create: `derived.entry_evaluation_requests` (via Go)

## DAGs

### Entry Experimentation

- Predict returns (Python)
- Predict investments (Python; excludes `excluded_entry_name`)
- Allocate entry (Python; MINLP)
- Submit entry evaluation request (Python -> Go API)
- Execute evaluation (Go worker: simulate + evaluate)

### Calcutta Predictions

- Submit calcutta evaluation request (Go or UI -> Go API)
- Execute evaluation (Go worker: simulate + evaluate)

## Schema contract

### `models` schema (Python-owned)

- `models.runs`
- `models.entry_candidates`
- `models.entry_candidate_bids`

Notes:

- These tables exist to support reproducibility and later extraction into a separate Python codebase.
- The UI may read these via Go API for convenience.

### `derived.entry_evaluation_requests` (Go-owned queue)

- One request corresponds to evaluating exactly one `models.entry_candidate_id` against one `core.calcutta_id`.
- `starting_state_key` is a string enum. Python only uses `post_first_four`.
- `excluded_entry_name` is supported for now.
- Requests can be grouped using `experiment_key`.

## API contract

### Create request

- `POST /api/entry-evaluation-requests`
- Auth: API key (Bearer token) + permission `analytics.entry_evaluation_requests.write`

Request body (shape):

- `calcuttaId` (uuid)
- `entryCandidateId` (uuid; `models.entry_candidates.id`)
- `excludedEntryName` (string, optional)
- `startingStateKey` (string; must be `post_first_four` for Python)
- `nSims` (int)
- `seed` (int)
- `experimentKey` (string, optional)
- `requestSource` (string, optional)

Response body (shape):

- `id` (uuid)
- `status` (string)

## Work plan

- [x] Add migration: create schema `models`
- [x] Add migration: create `models.runs`
- [x] Add migration: create `models.entry_candidates`
- [x] Add migration: create `models.entry_candidate_bids`
- [x] Add migration: create `derived.entry_evaluation_requests`
- [x] Add permission key for request submission (`analytics.entry_evaluation_requests.write`)
- [x] Add API endpoint: `POST /api/entry-evaluation-requests`
- [x] Add DTO: `CreateEntryEvaluationRequest`
- [ ] Add Go repository for `derived.entry_evaluation_requests`:
- [ ] - Insert (for API handler)
- [ ] - Claim next queued request (worker)
- [ ] - Mark running/succeeded/failed (worker)
- [ ] Add Go worker loop to claim queued requests
- [ ] Extend Go evaluation job to handle `starting_state_key`:
- [x] Add Go worker loop to claim queued requests
- [x] Extend Go evaluation job to handle `starting_state_key`:
- [x] - Resolve/create simulation state for `post_first_four`
- [x] - Run `simulate + evaluate` in a single job
- [ ] - Persist evaluation results to existing `derived.*` evaluation tables
- [ ] Update request status transitions: `queued -> running -> succeeded/failed`
- [ ] Harden worker semantics:
- [ ] - Ensure request ID is not reused as `run_id` for downstream writes
- [ ] - Deduplicate simulation batches when a compatible batch already exists
- [ ] - Add better logging + metrics for worker throughput
- [ ] Optional: add read endpoints for UI:
- [ ] - list requests by `experiment_key`
- [ ] - fetch request + linked evaluation run
- [ ] Decide what to do with the Go heuristic allocator (`recommended_entry_bids`) (debug-only vs remove)
