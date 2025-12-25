ENV = set -a; [ -f .env ] && . ./.env; [ -f .env.local ] && . ./.env.local; set +a;
ENV_DOCKER = set -a; [ -f .env ] && . ./.env; set +a;

DOCKER_PROJECT ?= calcutta
DC = docker compose -p $(DOCKER_PROJECT)

.PHONY: up down reset ops-migrate backend-test

backend-test:
	$(ENV) go -C backend test ./...

up:
	$(ENV_DOCKER) $(DC) up --build

down:
	$(ENV_DOCKER) $(DC) down

reset:
	$(ENV_DOCKER) $(DC) down -v

ops-migrate:
	$(ENV_DOCKER) $(DC) --profile ops run --rm migrate
