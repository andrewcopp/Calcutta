ENV = set -a; [ -f .env ] && . ./.env; [ -f .env.local ] && . ./.env.local; set +a;
ENV_DOCKER = set -a; [ -f .env ] && . ./.env; set +a;

DOCKER_PROJECT ?= calcutta
DC = docker compose -p $(DOCKER_PROJECT)
DC_AIRFLOW = docker compose -f data-science/docker-compose.airflow.yml

.PHONY: up down reset ops-migrate backend-test sqlc-generate
.PHONY: airflow-up airflow-down airflow-logs airflow-reset

backend-test:
	$(ENV) go -C backend test ./...

sqlc-generate:
	$(ENV) sqlc generate -f backend/sqlc.yaml

up:
	$(ENV_DOCKER) $(DC) up --build

down:
	$(ENV_DOCKER) $(DC) down

reset:
	$(ENV_DOCKER) $(DC) down -v

ops-migrate:
	$(ENV_DOCKER) $(DC) --profile ops run --rm migrate

# Airflow (full stack - heavyweight)
airflow-up:
	@echo "Starting Airflow (this may take 30-60 seconds)..."
	$(DC_AIRFLOW) up -d
	@echo "Airflow UI: http://localhost:8080 (airflow/airflow)"

airflow-down:
	$(DC_AIRFLOW) down

airflow-logs:
	$(DC_AIRFLOW) logs -f airflow-worker

airflow-reset:
	$(DC_AIRFLOW) down -v
	@echo "Airflow data volumes removed. Run 'make airflow-up' to reinitialize."
