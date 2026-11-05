# Дипломный проект Go School: каркас

Стартовая раскладка для любой из трёх тем диплома (бронирование переговорок, сервис
фича-флагов, заказы с outbox и вебхуками). ТЗ, этапы и чеклист приёмки: страница «Диплом» на сайте курса
(`site/content/diploma.md`).

Каркас уже умеет всё, что не зависит от темы: конфиг из окружения, JSON-логи `slog` с request_id,
`/healthz`, `/readyz` с параллельной проверкой PostgreSQL и Redis, `/metrics` для Prometheus,
graceful shutdown с паузой drain, миграции goose внутри бинарника, Docker multi-stage
на distroless, compose, Makefile, CI. Твоя работа: доменная модель, бизнес-логика, API и тесты.

Когда возьмёшь каркас в свой репозиторий, переименуй модуль под свой GitHub:
`go mod edit -module github.com/<ник>/<проект>` и замени импорты `goschool/diploma` (плюс
`local-prefixes` в `.golangci.yml`).

## Старт

```bash
cp .env.example .env
make db-up      # PostgreSQL на :55433, Redis на :56380
make migrate    # применить migrations/
make run        # API на :8080
curl -i localhost:8080/readyz
make test       # тесты с -race
make lint       # golangci-lint v2 (brew install golangci-lint)
```

Весь стенд в контейнерах, как на собесе «покажи, как запускается»:

```bash
make up         # build образа, база, Redis, миграции отдельным сервисом, API на :8080
curl localhost:8080/healthz
make down       # остановить; docker compose down -v ещё и сотрёт данные
```

Порт API занят: `API_PORT=18080 make up`. Консоль базы стенда: `make psql`
(первый админ, сид справочников, проверки из чеклиста).

## Раскладка

```
cmd/api/main.go           wiring и жизненный цикл; флаги -migrate и -healthcheck
internal/config/          конфиг из env с валидацией, секреты скрыты в логах
internal/http/            Gin: роутер, middleware (логи, request_id, метрики, recovery),
                          единый формат ошибок, /healthz, /readyz, /metrics, пример теста
internal/service/         доменные ошибки; сюда бизнес-правила и интерфейсы репозиториев
internal/repository/      DBTX (пул или транзакция), TxManager.InTx, MapErr; сюда SQL на pgx
migrations/               goose SQL, встроены через embed; новая: make migration name=...
Dockerfile                multi-stage: golang:1.27-alpine, затем distroless/static nonroot
docker-compose.yml        postgres, redis, migrate (тот же образ с -migrate), api
.github/workflows/ci.yml  lint, тесты против Postgres и Redis в service containers, docker build
```

## Архитектура

```
   клиенты: HTTP (JSON)              внутренние сервисы: gRPC (если тема требует)
            │                                    │
   ┌────────▼────────────────────────────────────▼───────┐
   │ internal/http: Gin, middleware, DTO, коды ошибок     │
   └────────┬─────────────────────────────────────────────┘
            │ интерфейсы сервисов
   ┌────────▼─────────────────────────────────────────────┐
   │ internal/service: правила, транзакции (TxManager)     │
   └────────┬──────────────────┬──────────────────┬───────┘
            │ репозитории      │ кеш, лимиты      │ InsertTx в той же транзакции
   ┌────────▼─────────┐  ┌─────▼─────┐   ┌────────▼────────────┐
   │ repository: pgx  │  │  Redis    │   │ River: воркеры      │
   └────────┬─────────┘  └───────────┘   │ (письма, outbox,    │
            │                            │  слив статистики)   │
   ┌────────▼────────────────────────┐   └────────┬────────────┘
   │ PostgreSQL: данные + river_job  │◄───────────┘
   └─────────────────────────────────┘
```

Зависимости смотрят только вниз: `http` знает про `service`, `service` не знает ни про Gin,
ни про pgx. Сборка всех слоёв руками в `cmd/api/main.go`.

## Что добавить по ходу

- **River** (все три темы): `go get github.com/riverqueue/river@v0.51.0 github.com/riverqueue/river/riverdriver/riverpgxv5@v0.51.0`.
  Таблицы River мигрируй через `rivermigrate` в `-migrate` рядом с goose. Клиент создаёшь в
  `main.go`, `Start` после старта HTTP, `Stop` при остановке до закрытия пула.
- **gRPC** (тема с фича-флагами): `api/proto/*.proto`, генерация через buf, сгенерированные `*.pb.go`
  коммить. Второй сервер на своём порту, останавливается `GracefulStop` в том же `select`.
- **Интеграционные тесты репозиториев**: в настоящую базу из `DATABASE_URL`, каждый тест в своей
  схеме (`CREATE SCHEMA test_<имя>`, `search_path` в строке подключения, миграции goose,
  `DROP SCHEMA ... CASCADE` в `t.Cleanup`), как `internal/testdb` из домашки. Без моков базы.

## Проверки перед сдачей

```bash
make ci                       # gofmt, vet, golangci-lint, тесты
make up && curl -s localhost:8080/readyz && make down
docker image ls goschool-diploma:dev   # размер образа для README
```
