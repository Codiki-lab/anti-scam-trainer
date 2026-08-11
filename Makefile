-include backend/.env
-include frontend/.env

export
export PROJECT_ROOT := $(CURDIR)

COMPOSE := docker compose \
	--project-name anti-scam-trainer \
	--env-file backend/.env \
	-f deploy/docker-compose.yml \
	-f deploy/docker-compose.ollama.yml
COMPOSE_OLLAMA := $(COMPOSE) --profile ollama

.PHONY: help env setup build up down logs lint test gateway-regression demo-reset db-reset \
	build-ollama up-ollama down-ollama logs-ollama ollama-init ollama-reset \
	start-ollama migrate-create migrate-up migrate-down clean

help:
	@echo "Available commands:"
	@echo "  make env                    Create backend/.env and add missing template variables"
	@echo "  make setup                  Initial project setup"
	@echo "  make build                  Build images without starting containers"
	@echo "  make build-ollama           Build images with the Ollama configuration"
	@echo "  make start-ollama           Build and start all services with Ollama, then pull the model"
	@echo "  make up                     Start previously built infrastructure without Ollama"
	@echo "  make up-ollama              Start previously built infrastructure with Ollama"
	@echo "  make down                   Stop infrastructure without Ollama"
	@echo "  make down-ollama            Stop infrastructure with Ollama"
	@echo "  make logs                   Show infrastructure logs"
	@echo "  make logs-ollama             Show Ollama logs"
	@echo "  make lint                   Run linters"
	@echo "  make test                   Run tests"
	@echo "  make gateway-regression     Verify Register → Login → Dashboard through the gateway"
	@echo "  make demo-reset             Recreate the deterministic seller demo account"
	@echo "  make db-reset               Delete all local PostgreSQL data; migrations restore seed content on next start"
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

start-ollama: build-ollama up-ollama ollama-init

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
	@cd backend && golangci-lint run --config .golangci.yaml ./...
	@if [ -f frontend/package.json ]; then cd frontend && npm run lint; else echo "frontend/package.json not found; frontend lint skipped"; fi

test:
	@cd backend && go test ./...
	@if [ -f frontend/package.json ]; then cd frontend && npm test; else echo "frontend/package.json not found; frontend tests skipped"; fi

gateway-regression:
	@./scripts/gateway-regression.sh

demo-reset:
	@cd backend && go run ./cmd/demo-reset

db-reset:
	@$(COMPOSE_OLLAMA) down --remove-orphans
	@rm -rf "$(PROJECT_ROOT)/out/pgdata"

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
