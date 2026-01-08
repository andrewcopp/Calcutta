# Contributing

## Quick start

### Backend (Go)

- Copy `.env.example` to `.env` and fill in required values.
- Run the API from the repo root:

```bash
go run ./backend/cmd/api
```

### Frontend

```bash
cd frontend
npm install
npm run dev
```

### Data-science

Use a virtual environment:

```bash
python3 -m venv data-science/.venv
source data-science/.venv/bin/activate
pip install -r data-science/requirements.txt
```

## Tests

- Go:

```bash
go test ./...
```

- Python (from the venv):

```bash
pytest
```

## Code style and expectations

- Keep PRs small and focused.
- Prefer deterministic logic and deterministic tests.
- Avoid hard exits (`log.Fatal`, `os.Exit`) outside of `main` packages.
- Never commit secrets. Use `.env.example` placeholders only.
- Prefer lean, flexible APIs (list + filters, composable primitives). Only add specialized/optimized endpoints when a real performance or UX constraint demands it.

## Database changes

- Schema changes must be done via committed, versioned migrations (no ad-hoc DDL on shared DBs).
