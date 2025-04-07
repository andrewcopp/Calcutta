# March Madness Investment Pool

A platform for running March Madness investment pools (Calcutta-style tournaments) where participants bid on teams in a blind auction format rather than filling out traditional brackets.

## Overview

This application manages March Madness investment pools where players:
- Receive virtual currency ($100) to invest in teams
- Participate in blind auctions for team ownership
- Earn points based on their teams' tournament performance
- Compete for the highest total points

## Technical Stack

### Frontend
- React - Single page application for user interface
- Modern React patterns (hooks, context)
- Responsive design for mobile compatibility

### Backend
- Go - RESTful API server
- Clean architecture principles
- JWT authentication

### Database
- PostgreSQL - Relational database
- Structured data model for users, teams, bids, and scoring

## Documentation

- [Complete Rules and Examples](docs/rules.md)
- [Technical Documentation](docs/technical/) (Coming Soon)
- [API Documentation](docs/technical/api.md) (Coming Soon)

## Project Status

🚧 Currently in initial planning and development phase.

## Getting Started

### Prerequisites

- Go 1.24 or later
- PostgreSQL 16 or later
- Node.js 18 or later
- npm 9 or later

### Database Setup

1. Create a PostgreSQL database named `calcutta`
2. Copy `.env.example` to `.env` and update the database connection settings
3. Run the database migrations:
   ```bash
   cd backend
   ./scripts/run_migrations.sh -up
   ```
4. Seed the database with initial data:
   ```bash
   cd backend
   ./scripts/run_seed.sh
   ```

### Running the Application

1. Start the backend server:
   ```bash
   cd backend
   go run cmd/server/main.go
   ```

2. Start the frontend development server:
   ```bash
   cd frontend
   npm install
   npm start
   ```

The application should now be running at `http://localhost:3000`

## Contributing

Coming soon...

## License

Coming soon...