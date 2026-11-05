# База общая с домашкой: контейнер из homework/docker-compose.yml, отдельная БД tasktracker.
COMPOSE := docker compose -f ../../homework/docker-compose.yml

.PHONY: db-up db-create run build test vet psql

db-up:        ## поднять PostgreSQL (порт 55432)
	$(COMPOSE) up -d postgres

db-create: db-up  ## создать базу tasktracker (один раз)
	$(COMPOSE) exec postgres createdb -U goschool tasktracker || true

run:          ## запустить API с переменными из .env (миграции применятся на старте)
	set -a && . ./.env && set +a && go run ./cmd/api

build:
	go build -o bin/api ./cmd/api

test:         ## юнит и интеграционные тесты (нужна поднятая база)
	go test -race ./...

vet:
	go vet ./...

psql:
	$(COMPOSE) exec postgres psql -U goschool tasktracker
