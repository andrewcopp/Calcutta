# March Madness Investment Pool

A platform for running March Madness investment pools (Calcutta-style tournaments) where participants bid on teams in a blind auction format rather than filling out traditional brackets.

## Overview

This application manages March Madness investment pools where players:
- Receive a fixed budget of in-game points (e.g. 100 points) to invest in teams
- Participate in blind auctions for team ownership
- Earn points based on their teams' tournament performance
- Compete for the highest total points

Points are the in-game currency used for bidding and scoring. Real-world dollars/cents (entry fees and payouts) are tracked separately.

## Technical Stack

### Frontend
- **React 18** with TypeScript
- **Vite** for build tooling
- **React Router** for navigation
- **Recharts** for data visualization
- **Tailwind CSS** for styling
- Modern React patterns (hooks, context)

### Backend
- **Go 1.24**
- **Gorilla Mux** for routing
- **PostgreSQL** driver (pgx/v5)
- Clean architecture principles (in transition)
- Multiple command binaries for different operations

### Database
- **PostgreSQL 16** with UUID extension
- Structured data model for users, teams, bids, and scoring
- Migration-based schema management
- Soft deletes with `deleted_at` timestamps

## Architecture Notes

### Backend structure

The backend is a standard Go module. The `backend/cmd/*` directories contain the runnable binaries (API server, migrations, and tools).

### Current vs. Future Architecture

**Current state:** Services and repositories are in `pkg/services/`. This works but doesn't follow strict clean architecture.

**Future state (documented in [docs/standards/engineering.md](docs/standards/engineering.md)):** 
- Interfaces in `internal/ports/`
- Implementations in `internal/adapters/postgres/`
- Services depend on ports, not concrete implementations

**For new code:** Follow the current structure for now. The migration is planned but not yet underway.

## Documentation

### Essential Reading
- **[Engineering Standards](docs/standards/engineering.md)** - Architecture principles and coding standards
- **[Testing Guidelines](docs/standards/bracket_testing_guidelines.md)** - **CRITICAL**: Strict testing conventions (one assert per test, TestThat naming)
- [Complete Rules and Examples](docs/reference/rules.md) - Business logic and game rules
- [Data Science Sandbox](docs/runbooks/data_science_sandbox.md) - Sandbox workflow + two-model framing (predict returns, predict investment, compute ROI)

### Domain Documentation
- [Bracket Management](docs/reference/bracket_management.md)
- [Bracket State Management](docs/reference/bracket_state_management.md)
- [Tournament Modeling](docs/reference/tournament_modeling.md)
- [Analytics](docs/design/analytics_future_enhancements.md)

### API Reference
See [API Endpoints](#api-endpoints) section below.

### Data science workflow

The Python utilities for snapshot ingestion, canonical dataset derivation, evaluation, and backtesting live in `data-science/`.

See: `data-science/README.md`

## Project Status

🚧 Active development - Core bracket and scoring logic implemented, analytics in progress.

## Getting Started

The preferred way to run this project is using Docker Compose, which will set up all necessary services automatically.

### Using Docker Compose (Recommended)

1. Make sure you have Docker and Docker Compose installed
2. Copy `.env.example` to `.env` and update any necessary settings
3. Start the application:
   ```bash
   docker compose up
   ```

The services start in this order:
- Database starts and becomes healthy
- Backend and frontend services start

Database schema migrations are run as an explicit ops step:
```bash
make ops-migrate
```

Initial data is imported via the admin UI (Bundles upload):
- Visit `http://localhost:3000/admin/bundles`
- Upload a bundle archive

The application will be available at `http://localhost:3000`

To start fresh, tear down the environment and restart:
```bash
docker compose down -v  # -v removes volumes including database data
docker compose up
```

### Manual Setup (Alternative)

If you prefer to run the services manually, follow these steps:

#### Prerequisites

- **Go 1.24 or later**
- **PostgreSQL 16 or later**
- **Node.js 18 or later**
- **npm 9 or later**

#### Environment Configuration

Copy `.env.example` to `.env` and configure for local development:

```bash
cp .env.example .env
```

**Important:** For local development (not Docker), update these values:
```bash
DB_HOST=localhost        # Change from 'db' to 'localhost'
API_URL=http://localhost:8080  # Change from 'http://backend:8080'
```

#### Database Setup

1. Create a PostgreSQL database:
   ```bash
   createdb calcutta
   ```

2. Run migrations:
   ```bash
   make ops-migrate
   ```

3. Import initial data:
   - Visit `http://localhost:3000/admin/bundles`
   - Upload a bundle archive

#### Running the Application

1. Start the backend server:
   ```bash
   cd backend
   go run ./cmd/api
   ```
   Backend will be available at `http://localhost:8080`

2. In a new terminal, start the frontend:
   ```bash
   cd frontend
   npm install
   npm run dev
   ```
   Frontend will be available at `http://localhost:3000`

### Troubleshooting

#### Docker Issues

**Problem:** Migrations fail on startup
- **Solution:** Check database logs: `docker compose logs db`
- Ensure PostgreSQL is fully started before migrations run
- Try: `docker compose down -v && docker compose up`

**Problem:** Port conflicts (3000, 8080, 5432)
- **Solution:** Stop other services using these ports
- Or modify ports in `docker-compose.yml`

#### Database Issues

**Problem:** Connection refused to PostgreSQL
- Local dev: Ensure PostgreSQL is running and `DB_HOST=localhost`
- Docker: Use `DB_HOST=db` (service name)

**Problem:** Migration version conflicts
- Migrations are timestamped and run in order
- Check applied migrations: `SELECT * FROM schema_migrations;`
- Never modify existing migrations; create new ones

### Running Tests

```bash
# Run all backend tests
cd backend
go test ./...

# Run specific package tests
go test ./pkg/services/...

# Run with verbose output
go test -v ./pkg/services/...

# Run specific test
go test -v ./pkg/services/ -run TestThatBracketBuilderGeneratesSameGameIDs
```

**Note:** Tests follow strict conventions (see [Testing Guidelines](docs/standards/bracket_testing_guidelines.md))

## API Endpoints

### Authentication
- `POST /api/auth/login` - User login
- `POST /api/auth/signup` - User registration

### Basic
- `GET /api/health` - Health check
- `GET /api/schools` - List all schools

### Tournaments
- `GET /api/tournaments` - List all tournaments
- `GET /api/tournaments/{id}` - Get tournament details
- `POST /api/tournaments` - Create tournament
- `GET /api/tournaments/{id}/teams` - Get tournament teams
- `POST /api/tournaments/{id}/teams` - Add team to tournament
- `PATCH /api/tournaments/{tournamentId}/teams/{teamId}` - Update team
- `POST /api/tournaments/{id}/recalculate-portfolios` - Recalculate portfolio scores

### Bracket Management
- `GET /api/tournaments/{id}/bracket` - Get bracket structure
- `GET /api/tournaments/{id}/bracket/validate` - Validate bracket setup
- `POST /api/tournaments/{tournamentId}/bracket/games/{gameId}/winner` - Select game winner
- `DELETE /api/tournaments/{tournamentId}/bracket/games/{gameId}/winner` - Unselect game winner

### Calcuttas
- `GET /api/calcuttas` - List all calcuttas
- `POST /api/calcuttas` - Create calcutta
- `GET /api/calcuttas/{id}` - Get calcutta details
- `GET /api/calcuttas/{id}/entries` - Get calcutta entries
- `GET /api/calcuttas/{calcuttaId}/entries/{entryId}/teams` - Get entry teams

### Portfolios
- `GET /api/entries/{id}/portfolios` - Get portfolios for entry
- `GET /api/portfolios/{id}/teams` - Get portfolio teams
- `POST /api/portfolios/{id}/calculate-scores` - Calculate portfolio scores
- `PUT /api/portfolios/{id}/teams/{teamId}/scores` - Update team scores
- `PUT /api/portfolios/{id}/maximum-score` - Update maximum score

### Analytics
- `GET /api/analytics` - Get general analytics
- `GET /api/analytics/seeds` - Seed performance analytics
- `GET /api/analytics/regions` - Regional analytics
- `GET /api/analytics/teams` - Team analytics
- `GET /api/analytics/variance` - Seed variance analytics

## Contributing

### Before You Start

**Required Reading:**
1. **[Engineering Standards](docs/standards/engineering.md)** - Architecture and coding principles
2. **[Testing Guidelines](docs/standards/bracket_testing_guidelines.md)** - Testing conventions (non-negotiable)

### Code Review Expectations

#### Testing Standards (Strictly Enforced)

All tests **must** follow these conventions:

1. **One assertion per test** - Each test has exactly one reason to fail
2. **TestThat{Scenario} naming** - Example: `TestThatBracketBuilderGeneratesSameGameIDsAcrossMultipleBuilds`
3. **GIVEN/WHEN/THEN structure** - Clear setup, action, and assertion

**Example:**
```go
func TestThatFirstFourGameForElevenSeedHasDeterministicID(t *testing.T) {
    // GIVEN a region with two 11-seeds
    teams := createRegionWithDuplicateSeeds("East", 11, 16)
    builder := NewBracketBuilder()
    bracket := &models.BracketStructure{
        TournamentID: "test",
        Games:        make(map[string]*models.BracketGame),
    }

    // WHEN building the regional bracket
    _, err := builder.buildRegionalBracket(bracket, "East", teams)
    if err != nil {
        t.Fatalf("Failed to build regional bracket: %v", err)
    }

    // THEN First Four game for 11-seeds has ID 'East-first_four-11'
    if bracket.Games["East-first_four-11"] == nil {
        t.Error("Expected First Four game with ID 'East-first_four-11'")
    }
}
```

#### Code Style

- **Handlers:** Parse input → call service → return response DTO
- **Services:** Business logic only, no logging, return errors
- **No upward imports:** Lower layers never import higher layers
- **File size:** Keep files under 200-400 LOC; split by responsibility
- **Function size:** Aim for 20-60 LOC; orchestration may be longer

#### Pull Request Process

1. Ensure all tests pass: `go test ./...`
2. Follow existing file organization patterns
3. Add tests for new functionality (one assert per test!)
4. Update documentation if adding features
5. Keep PRs focused on a single concern

### Development Workflow

1. Create a feature branch from `main`
2. Make changes following the coding standards
3. Write tests (remember: one assertion per test)
4. Run tests locally
5. Submit PR with clear description
6. Address review feedback

### Common Pitfalls for New Contributors

- **Multiple assertions in tests:** Will be rejected in code review
- **Wrong test naming:** Must use `TestThat{Scenario}` format
- **Adding logic to handlers:** Business logic belongs in services
- **Modifying existing migrations:** Always create new migration files

## License

Coming soon...