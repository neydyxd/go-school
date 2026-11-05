GO    ?= go
PKG   ?= ./...
IMAGE ?= goschool-diploma:dev
# Переменные из .env для локального запуска (cp .env.example .env)
ENV   := set -a; [ ! -f .env ] || . ./.env; set +a;

.DEFAULT_GOAL := help
.PHONY: help db-up psql run migrate migration test cover fmt-check vet lint build docker up down logs ci

help: ## показать цели
	@grep -E '^[a-z-]+:.*## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*## "} {printf "  %-10s %s\n", $$1, $$2}'

db-up: ## поднять только PostgreSQL и Redis (для make run и тестов)
	docker compose up -d --wait postgres redis

psql: ## psql в базе стенда (сид, повышение до admin, проверки)
	docker compose exec postgres psql -U diploma diploma

run: ## API локально с переменными из .env
	$(ENV) $(GO) run ./cmd/api

migrate: ## применить миграции из migrations/ к базе из .env
	$(ENV) $(GO) run ./cmd/api -migrate

migration: ## новая миграция: make migration name=create_rooms
	@test -n "$(name)" || (echo "укажи name=..."; exit 1)
	$(GO) run github.com/pressly/goose/v3/cmd/goose@v3.29.0 -s -dir migrations create $(name) sql

test: ## юнит и интеграционные тесты с детектором гонок (нужен make db-up)
	$(ENV) $(GO) test -race -count=1 $(PKG)

cover: ## итог покрытия
	$(ENV) $(GO) test -count=1 -coverprofile=coverage.out $(PKG) && $(GO) tool cover -func=coverage.out | tail -1

fmt-check: ## упасть, если есть неотформатированные файлы
	@out=$$(gofmt -l .); if [ -n "$$out" ]; then echo "gofmt не прогнан:"; echo "$$out"; exit 1; fi

vet: ## go vet
	$(GO) vet $(PKG)

lint: ## golangci-lint
	golangci-lint run $(PKG)

build: ## бинарник в bin/api
	CGO_ENABLED=0 $(GO) build -trimpath -o bin/api ./cmd/api

docker: ## собрать образ
	docker build -t $(IMAGE) .

up: ## весь стенд в Docker: база, Redis, миграции, API на :8080
	docker compose up -d --build --wait

down: ## остановить стенд (данные остаются в томе; стереть: docker compose down -v)
	docker compose down

logs: ## логи API
	docker compose logs -f api

ci: fmt-check vet lint test ## то же, что CI, локально
