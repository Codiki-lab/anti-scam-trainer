-include backend/.env
-include frontend/.env

export
export PROJECT_ROOT := $(CURDIR)

COMPOSE := docker compose --env-file backend/.env -f deploy/docker-compose.yml
COMPOSE_OLLAMA := docker compose \
	--env-file backend/.env \
	-f deploy/docker-compose.yml \
	-f deploy/docker-compose.ollama.yml

.PHONY: help env setup build up down logs lint test \
	build-ollama up-ollama down-ollama logs-ollama ollama-init ollama-reset \
	migrate-create migrate-up migrate-down clean

help:
	@echo "Available commands:"
	@echo "  make env                    Create backend/.env and add missing template variables"
	@echo "  make setup                  Initial project setup"
	@echo "  make build                  Build images without starting containers"
	@echo "  make build-ollama           Build images with the Ollama configuration"
	@echo "  make up                     Start previously built infrastructure without Ollama"
	@echo "  make up-ollama              Start previously built infrastructure with Ollama"
	@echo "  make down                   Stop infrastructure without Ollama"
	@echo "  make down-ollama            Stop infrastructure with Ollama"
	@echo "  make logs                   Show infrastructure logs"
	@echo "  make logs-ollama             Show Ollama logs"
	@echo "  make lint                   Run linters"
	@echo "  make test                   Run tests"
	@echo "  make migrate-create seq=xx  Create migration"
	@echo "  make migrate-up             Apply migrations"
	@echo "  make migrate-down           Rollback migrations"
	@echo "  make clean                  Remove containers and volumes"

env:
	@if [ ! -f backend/.env.example ]; then \
		echo "Missing backend/.env.example"; exit 1; \
	fi
	@touch backend/.env
	@while IFS= read -r line || [ -n "$$line" ]; do \
		case "$$line" in ''|\#*) continue ;; esac; \
		key=$${line%%=*}; \
		if ! grep -q "^[[:space:]]*$$key=" backend/.env; then \
			printf '%s\n' "$$line" >> backend/.env; \
		fi; \
	done < backend/.env.example

setup: env
	@if [ ! -f frontend/.env ]; then \
		if [ -f frontend/.env.example ]; then cp frontend/.env.example frontend/.env; \
		else echo "Missing frontend/.env.example"; exit 1; fi; \
	fi
	@if [ -f backend/go.mod ]; then cd backend && go mod download; else echo "backend/go.mod not found; Go setup skipped"; fi
	@if [ -f frontend/package.json ]; then cd frontend && npm install; else echo "frontend/package.json not found; frontend setup skipped"; fi

build: env
	@$(COMPOSE) build

build-ollama: env
	@$(COMPOSE_OLLAMA) build

up: env
	@$(COMPOSE) up -d --no-build

up-ollama: env
	@$(COMPOSE_OLLAMA) up -d --no-build

down:
	@$(COMPOSE) down

down-ollama:
	@$(COMPOSE_OLLAMA) down

logs:
	@$(COMPOSE) logs -f

logs-ollama:
	@$(COMPOSE_OLLAMA) logs -f anti-scam-trainer-ollama

lint:
	@cd backend && golangci-lint run ./...
	@cd frontend && npm run lint

test:
	@cd backend && go test ./...
	@cd frontend && npm test

migrate-create:
	@if [ -z "$(seq)" ]; then \
		echo "Usage: make migrate-create seq=create_users"; \
		exit 1; \
	fi
	@$(COMPOSE) run --rm anti-scam-trainer-migrate \
		create \
		-ext sql \
		-dir /migrations \
		-seq "$(seq)"

migrate-up:
	@$(COMPOSE) run --rm anti-scam-trainer-migrate \
		-path=/migrations \
		-database="postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@anti-scam-trainer-postgres:5432/$(POSTGRES_DB)?sslmode=disable" \
		up

migrate-down:
	@$(COMPOSE) run --rm anti-scam-trainer-migrate \
		-path=/migrations \
		-database="postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@anti-scam-trainer-postgres:5432/$(POSTGRES_DB)?sslmode=disable" \
		down 1

clean:
	@$(COMPOSE_OLLAMA) down -v --remove-orphans

ollama-init:
	@$(COMPOSE_OLLAMA) run --rm anti-scam-trainer-ollama-init

ollama-logs:
	@$(COMPOSE_OLLAMA) logs -f anti-scam-trainer-ollama

ollama-reset:
	@$(COMPOSE_OLLAMA) rm -sf anti-scam-trainer-ollama-init
	@$(COMPOSE_OLLAMA) run --rm anti-scam-trainer-ollama-init
