# Database Scripts

This directory contains scripts for database operations.

## Available Scripts

- **run_migrations.sh**: Runs database migrations (up or down)
- **test_db.sh**: Tests the database connection
- **seed_schools.sh**: Seeds the database with schools from the master JSON file

## Usage

### Running Migrations

```bash
# Run migrations up
./run_migrations.sh

# Run migrations down
./run_migrations.sh down
```

### Testing Database Connection

```bash
./test_db.sh
```

### Seeding Schools

```bash
./seed_schools.sh
``` 