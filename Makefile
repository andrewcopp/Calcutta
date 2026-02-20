ENV = set -a; [ -f .env ] && . ./.env; [ -f .env.local ] && . ./.env.local; set +a;
ENV_DOCKER = set -a; [ -f .env ] && . ./.env; set +a;

DOCKER_PROJECT ?= calcutta
DC = docker compose -p $(DOCKER_PROJECT)

.PHONY: env-init bootstrap bootstrap-admin dev up-d logs ps stats
.PHONY: up down reset ops-migrate backend-test sqlc-generate
.PHONY: db-shell db-query db query query-file query-csv
.PHONY: logs-backend logs-worker logs-db logs-frontend logs-search logs-tail
.PHONY: restart-backend restart-worker restart-frontend restart-db
.PHONY: db-ping db-sizes db-activity db-vacuum api-health api-test
.PHONY: register-models export-bundles import-bundles

env-init:
	@if [ ! -f .env ]; then cp .env.example .env; fi
	@if [ ! -f .env.local ]; then \
		cp .env.example .env.local; \
		perl -pi -e 's/^DB_HOST=.*/DB_HOST=localhost/' .env.local; \
		perl -pi -e 's|^API_URL=.*|API_URL=http://localhost:8080|' .env.local; \
		perl -pi -e 's/^SMTP_HOST=.*/SMTP_HOST=localhost/' .env.local; \
	fi

dev: up-d ops-migrate

bootstrap: env-init dev

bootstrap-admin:
	$(ENV_DOCKER) $(DC) --profile ops run --rm migrate -bootstrap

logs:
	$(ENV_DOCKER) $(DC) logs -f

logs-backend:
	$(ENV_DOCKER) $(DC) logs -f backend

logs-worker:
	$(ENV_DOCKER) $(DC) logs -f worker

logs-db:
	$(ENV_DOCKER) $(DC) logs -f db

logs-frontend:
	$(ENV_DOCKER) $(DC) logs -f frontend

logs-search:
	@if [ -z "$(PATTERN)" ]; then echo "Usage: make logs-search PATTERN=\"error\""; exit 1; fi
	@LINES=$${LINES:-1000}; $(ENV_DOCKER) $(DC) logs --tail=$$LINES | grep -i "$(PATTERN)"

logs-tail:
	@LINES=$${LINES:-50}; $(ENV_DOCKER) $(DC) logs --tail=$$LINES

ps:
	$(ENV_DOCKER) $(DC) ps

stats:
	$(ENV_DOCKER) docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}" $$($(DC) ps -q)

backend-test:
	$(ENV) go -C backend test ./...

sqlc-generate:
	$(ENV) sqlc generate -f backend/sqlc.yaml

up:
	$(ENV_DOCKER) $(DC) up --build

up-d:
	$(ENV_DOCKER) $(DC) up -d --build

down:
	$(ENV_DOCKER) $(DC) down

reset:
	$(ENV_DOCKER) $(DC) down -v

ops-migrate:
	$(ENV_DOCKER) $(DC) --profile ops run --rm migrate

# Service restart commands
restart-backend:
	$(ENV_DOCKER) $(DC) restart backend

restart-worker:
	$(ENV_DOCKER) $(DC) restart worker

restart-frontend:
	$(ENV_DOCKER) $(DC) restart frontend

restart-db:
	$(ENV_DOCKER) $(DC) restart db

# API testing helpers
api-health:
	@OUT=$$(curl -s http://localhost:8080/api/health); echo "$$OUT" | jq . 2>/dev/null || echo "$$OUT"

api-test:
	@if [ -z "$(ENDPOINT)" ]; then echo "Usage: make api-test ENDPOINT=\"/api/calcuttas\" [METHOD=GET] [DATA='{}']"; exit 1; fi
	@METHOD=$${METHOD:-GET}; \
	if [ -z "$(DATA)" ]; then \
		OUT=$$(curl -s -X $$METHOD http://localhost:8080$(ENDPOINT)); echo "$$OUT" | jq . 2>/dev/null || echo "$$OUT"; \
	else \
		OUT=$$(curl -s -X $$METHOD -H "Content-Type: application/json" -d '$(DATA)' http://localhost:8080$(ENDPOINT)); echo "$$OUT" | jq . 2>/dev/null || echo "$$OUT"; \
	fi

# Database access (via docker exec - works regardless of host psql)
db-shell:
	$(ENV_DOCKER) docker exec -it $$($(DC) ps -q db) psql -U $${DB_USER} -d $${DB_NAME}

# Alias for db-shell (shorter command)
db: db-shell

db-query:
	@if [ -z "$(SQL)" ]; then echo "Usage: make db-query SQL=\"SELECT ...\""; exit 1; fi
	$(ENV_DOCKER) docker exec $$($(DC) ps -q db) psql -U $${DB_USER} -d $${DB_NAME} -c "$(SQL)"

# Alias for db-query (shorter command)
query: db-query

# Run SQL from a file
query-file:
	@if [ -z "$(FILE)" ]; then echo "Usage: make query-file FILE=\"path/to/file.sql\""; exit 1; fi
	@if [ ! -f "$(FILE)" ]; then echo "Error: File $(FILE) not found"; exit 1; fi
	$(ENV_DOCKER) docker exec -i $$($(DC) ps -q db) psql -U $${DB_USER} -d $${DB_NAME} -v ON_ERROR_STOP=1 < "$(FILE)"

# Export query results as CSV
query-csv:
	@if [ -z "$(SQL)" ]; then echo "Usage: make query-csv SQL=\"SELECT ...\""; exit 1; fi
	$(ENV_DOCKER) docker exec $$($(DC) ps -q db) psql -U $${DB_USER} -d $${DB_NAME} -c "COPY ($(SQL)) TO STDOUT WITH CSV HEADER"

# Database health and diagnostics
db-ping:
	@$(ENV_DOCKER) docker exec $$($(DC) ps -q db) pg_isready -U $${DB_USER} || echo "Database not ready"

db-sizes:
	@$(ENV_DOCKER) docker exec $$($(DC) ps -q db) psql -U $${DB_USER} -d $${DB_NAME} -c "\
	SELECT \
	  schemaname, \
	  tablename, \
	  pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size \
	FROM pg_tables \
	WHERE schemaname NOT IN ('pg_catalog', 'information_schema') \
	ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC \
	LIMIT 20;"

db-activity:
	@$(ENV_DOCKER) docker exec $$($(DC) ps -q db) psql -U $${DB_USER} -d $${DB_NAME} -c "\
	SELECT \
	  pid, \
	  usename, \
	  application_name, \
	  client_addr, \
	  state, \
	  query_start, \
	  state_change, \
	  LEFT(query, 100) AS query \
	FROM pg_stat_activity \
	WHERE state != 'idle' \
	ORDER BY query_start;"

db-vacuum:
	@echo "Running VACUUM ANALYZE (this may take a few seconds)..."
	@$(ENV_DOCKER) docker exec $$($(DC) ps -q db) psql -U $${DB_USER} -d $${DB_NAME} -c "VACUUM ANALYZE;"
	@echo "Done."

# Data science
register-models:
	@echo "Registering investment models..."
	@$(ENV) cd data-science && \
		. .venv/bin/activate && \
		python scripts/register_investment_models.py

export-bundles:
	@$(ENV) cd backend && go run ./cmd/tools/export-bundles -out=./exports/bundles

import-bundles:
	@DRY_RUN=$${DRY_RUN:-true}; \
	$(ENV) cd backend && go run ./cmd/tools/import-bundles -in=./exports/bundles -dry-run=$$DRY_RUN
