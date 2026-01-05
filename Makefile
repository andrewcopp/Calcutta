ENV = set -a; [ -f .env ] && . ./.env; [ -f .env.local ] && . ./.env.local; set +a;
ENV_DOCKER = set -a; [ -f .env ] && . ./.env; set +a;

DOCKER_PROJECT ?= calcutta
DC = docker compose -p $(DOCKER_PROJECT)
DC_PROD = docker compose -f docker-compose.local-prod.yml -p $(DOCKER_PROJECT)
DC_AIRFLOW = docker compose -f data-science/docker-compose.airflow.yml

.PHONY: env-init bootstrap dev dev-up dev-down up-d logs ps
.PHONY: prod-up prod-down prod-reset prod-ops-migrate
.PHONY: up down reset ops-migrate backend-test sqlc-generate
.PHONY: reset-derived
.PHONY: airflow-up airflow-down airflow-logs airflow-reset

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

ps:
	$(ENV_DOCKER) $(DC) ps

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
