ENV = set -a; [ -f .env ] && . ./.env; [ -f .env.local ] && . ./.env.local; set +a;
ENV_DOCKER = set -a; [ -f .env ] && . ./.env; set +a;

DOCKER_PROJECT ?= calcutta
DC = docker compose -p $(DOCKER_PROJECT)
DC_PROD = docker compose -f docker-compose.local-prod.yml -p $(DOCKER_PROJECT)
DC_AIRFLOW = docker compose -f data-science/docker-compose.airflow.yml

.PHONY: env-init bootstrap dev dev-up dev-down up-d logs ps stats
.PHONY: prod-up prod-down prod-reset prod-ops-migrate
.PHONY: up down reset ops-migrate backend-test sqlc-generate
.PHONY: reset-derived db-shell db-query db query query-file query-csv
.PHONY: logs-backend logs-worker logs-db logs-frontend logs-search logs-tail
.PHONY: restart-backend restart-worker restart-frontend restart-db
.PHONY: db-ping db-sizes db-activity db-vacuum api-health api-test curl
.PHONY: airflow-up airflow-down airflow-logs airflow-reset
.PHONY: register-models

env-init:
	@if [ ! -f .env ]; then cp .env.example .env; fi
	@if [ ! -f .env.local ]; then \
		cp .env.example .env.local; \
		perl -pi -e 's/^DB_HOST=.*/DB_HOST=localhost/' .env.local; \
		perl -pi -e 's|^API_URL=.*|API_URL=http://localhost:8080|' .env.local; \
		perl -pi -e 's/^SMTP_HOST=.*/SMTP_HOST=localhost/' .env.local; \
	fi

dev-up: up-d

dev-down: down

dev: dev-up ops-migrate

bootstrap: env-init dev

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
	$(ENV_DOCKER) $(DC) logs --tail=100 | grep -i "$(PATTERN)"

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

prod-up:
	$(ENV_DOCKER) $(DC_PROD) up -d --build

prod-down:
	$(ENV_DOCKER) $(DC_PROD) down

prod-reset:
	$(ENV_DOCKER) $(DC_PROD) down -v

prod-ops-migrate:
	$(ENV_DOCKER) $(DC_PROD) --profile ops run --rm migrate

ops-migrate:
	$(ENV_DOCKER) $(DC) --profile ops run --rm migrate

reset-derived:
	$(ENV) psql "postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable" -v ON_ERROR_STOP=1 -f backend/ops/db/maintenance/reset_derived_data.sql

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
	@curl -s http://localhost:8080/api/health | jq . || curl -s http://localhost:8080/api/health

api-test:
	@if [ -z "$(ENDPOINT)" ]; then echo "Usage: make api-test ENDPOINT=\"/api/calcuttas\" [METHOD=GET] [DATA='{}']"; exit 1; fi
	@METHOD=$${METHOD:-GET}; \
	if [ -z "$(DATA)" ]; then \
		curl -s -X $$METHOD http://localhost:8080$(ENDPOINT) | jq . || curl -s -X $$METHOD http://localhost:8080$(ENDPOINT); \
	else \
		curl -s -X $$METHOD -H "Content-Type: application/json" -d '$(DATA)' http://localhost:8080$(ENDPOINT) | jq . || \
		curl -s -X $$METHOD -H "Content-Type: application/json" -d '$(DATA)' http://localhost:8080$(ENDPOINT); \
	fi

curl:
	@if [ -z "$(URL)" ]; then echo "Usage: make curl URL=\"http://localhost:8080/api/health\""; exit 1; fi
	@curl -s $(URL) | jq . || curl -s $(URL)

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

# Airflow (full stack - heavyweight)
airflow-up:
	@echo "Starting Airflow (this may take 30-60 seconds)..."
	$(DC_AIRFLOW) up -d
	@echo "Airflow UI: http://localhost:8081 (airflow/airflow)"

airflow-down:
	$(DC_AIRFLOW) down

airflow-logs:
	$(DC_AIRFLOW) logs -f airflow-worker

airflow-reset:
	$(DC_AIRFLOW) down -v
	@echo "Airflow data volumes removed. Run 'make airflow-up' to reinitialize."

# Data science
register-models:
	@echo "Registering investment models..."
	@cd data-science && \
		. .venv/bin/activate && \
		DB_HOST=localhost DB_USER=$${DB_USER:-calcutta} DB_PASSWORD=$${DB_PASSWORD:-calcutta} \
		DB_NAME=$${DB_NAME:-calcutta} DB_PORT=$${DB_PORT:-5432} \
		python scripts/register_investment_models.py
