---
title: Диплом
outline: [2, 3]
---

# Диплом: свой сервис со всем стеком курса

Диплом это один сервис, который ты делаешь сам от пустой папки до `docker compose up`, и в котором сходится всё, что было в курсе: PostgreSQL с честными ограничениями, слои handler/service/repository, тесты в настоящей базе, фоновые задачи на River, Redis, Docker, метрики, CI. Не учебная задача с готовыми тестами, а проект, который можно открыть на собесе и полчаса про него разговаривать.

Зачем он нужен. У джуна без коммерческого опыта на собесе почти нет фактов о себе: «прошёл курс» ничего не говорит интервьюеру. Репозиторий с работающим сервисом, зелёным CI и README со схемой говорит много: ты умеешь довести дело до конца, понимаешь, зачем нужны миграции и readiness, и можешь объяснить, почему бронь не задвоится при двух одновременных запросах. Хороший диплом превращает собес из экзамена в разговор о твоём проекте, где ты знаешь ответы лучше всех.

«Трекер задач» из модуля про Gin уже научил тебя базовому REST API с JWT, пагинацией и IDOR. Диплом не повторяет его: у каждой темы своя техническая изюминка, о которой спрашивают на собесах (конкурентные записи, кеш и gRPC, транзакционный outbox). Код аутентификации из трекера можешь переносить как есть: это не то, на чём ты будешь блистать.

## Как устроен диплом

- **Срок** при 8-12 часах в неделю: бронирование 4 недели, фича-флаги и заказы по 5 недель. Это оценка для человека, который делает такое впервые: если этап занял полторы недели, всё в порядке. У тем 2 и 3 отмечено, что можно перенести в «со звёздочкой», если время кончается; обязательный минимум всё равно покрывает изюминку темы. Этапы разбиты по неделям, у каждого свой результат, который можно показать.
- **Старт**: общий каркас `projects/diploma/` (отдельный Go-модуль). В нём уже готово всё, что не зависит от темы: конфиг из окружения, `slog`, `/healthz`, `/readyz`, `/metrics`, graceful shutdown, миграции goose внутри бинарника, Dockerfile, compose с PostgreSQL и Redis, Makefile, CI, `.golangci.yml`. Скопируй папку в свой репозиторий на GitHub и переименуй модуль (инструкция в README каркаса).
- **Эталона нет.** Проверка по чеклисту приёмки своей темы и общим критериям. Самопроверка честная: каждый пункт чеклиста подтверждается командой.
- **Одна тема.** Выбирай ту, о которой хочется рассказывать. Если сомневаешься: бронирование проще всего объяснить, фича-флаги дают больше всего разных технологий, заказы ближе всего к реальному продакшну.

```bash
cd projects/diploma
cp .env.example .env
make db-up && make migrate && make run   # API на :8080
curl -s localhost:8080/readyz             # {"checks":{"postgres":"ok","redis":"ok"},"status":"ready"}
make test                                 # пример теста роутера зелёный
make up                                   # весь стенд в Docker: база, Redis, миграции, API
```

## Три темы на выбор

| Тема | Изюминка | Что разберёте на собесе |
|------|----------|--------------------------|
| [Бронирование переговорок](#booking), 4 недели | Конкурентные брони: `EXCLUDE`-ограничение на пересечение интервалов, напоминания и автоотмена через River | Гонки и где их ловить, ограничения в БД против проверок в коде, отложенные задачи и их отмена |
| [Сервис фича-флагов](#flags), 5 недель | Горячий путь чтения: Redis-кеш с инвалидацией, gRPC для внутренних сервисов, rate limiting по ключу, агрегация статистики фоновой задачей | Кеширование и согласованность, gRPC против REST, идемпотентная агрегация, детерминированный раскат на процент |
| [Заказы с outbox и вебхуками](#orders), 5 недель | Транзакционный outbox на River, идемпотентность `POST` по `Idempotency-Key`, резерв остатков, доставка вебхуков с ретраями и подписью | Dual write, at-least-once, идемпотентность, дедлоки, наблюдаемость фоновой обработки |

::: details Почему здесь нет сокращателя ссылок
Сокращатель ссылок был проектом модуля про HTTP на стандартной библиотеке и разобран как задача system design в [уроке про проектирование](/m10/l05-system-design). Третий раз та же предметная область не даст нового опыта, а на собесе прозвучит как учебный проект. Его технические изюминки (Redis-кеш, rate limiting, агрегация кликов, gRPC для внутренних клиентов) целиком перешли в тему фича-флагов: там они нужны по делу, а не ради галочки, и есть новое (инвалидация кеша при записи, раскат на процент, аудит).
:::

## Обязательный стек и критерии

Эти требования общие для всех тем. Без них диплом не засчитывается, даже если бизнес-логика работает.

### Код и архитектура

- **Слои** как в [уроке про архитектуру](/m08/l04-architecture): `internal/http` (Gin, DTO, коды ошибок) зовёт `internal/service` (правила, транзакции) через интерфейсы, `internal/service` зовёт `internal/repository` (SQL на pgx) через интерфейсы, объявленные у потребителя. Сервис не импортирует ни Gin, ни pgx. Вся сборка руками в `cmd/api/main.go`.
- **Единый формат ошибок** `{"error":{"code":"...","message":"...","fields":{...}}}` (уже есть в каркасе). Текст ошибок базы клиенту не уходит. Коды: 400 `bad_request`, 401, 403, 404, 409 `conflict`, 422 `validation_failed`, 429 `rate_limited`, 500 `internal`.
- **Конфиг только из окружения**, валидация при старте, все ошибки сразу. Секреты не попадают в логи (`LogValue` в каркасе).
- **Время через параметр.** Сервис получает `now func() time.Time`, а не зовёт `time.Now()` внутри: иначе правила вроде «нельзя бронировать в прошлом» не протестировать.

### База

- **Миграции goose** только новыми файлами (`make migration name=...`), применяются отдельным шагом (`-migrate`, сервис `migrate` в compose).
- **Целостность в базе, а не только в коде**: `NOT NULL`, `CHECK`, `UNIQUE`, внешние ключи, а для темы бронирования `EXCLUDE`. Проверка в Go нужна для понятного сообщения пользователю, ограничение в базе нужно для корректности при гонках.
- **Индексы под реальные запросы**, и для одного главного запроса в README приведён `EXPLAIN ANALYZE` с объяснением, почему индекс используется.
- Деньги в целых копейках (`bigint`) или `numeric`, время в `timestamptz`, наружу в RFC 3339.

### Тесты

- **Юнит-тесты сервисного слоя** с fake-репозиториями: все бизнес-правила, граничные случаи, переходы статусов.
- **Хендлеры через `httptest`** с настоящим роутером и fake-сервисами: коды ответов 200/201/400/401/404/409/422, формат ошибок. Образец: `internal/http/router_test.go` в каркасе.
- **Интеграционные тесты репозиториев в настоящей PostgreSQL**, своя схема на тест, без моков базы. Главная гарантия темы (нет двойной брони, нет двойного списания остатка, нет двойного счёта статистики) доказана конкурентным тестом: N горутин одновременно, проверяется итог в базе.
- **Фоновые задачи** через `rivertest`: задача поставлена в той же транзакции, с нужными аргументами и временем.
- `go test -race -count=1 ./...` зелёный, тесты не зависят от порядка и не спят `time.Sleep` в ожидании результата.

### Продакшн

- **Docker multi-stage** на distroless nonroot (готово в каркасе), размер образа записан в README.
- **`docker compose up`** поднимает всё с нуля на чистой машине: база, Redis, миграции, API, воркеры River.
- **CI на GitHub Actions** (готово в каркасе): gofmt, `go vet`, golangci-lint, тесты против PostgreSQL и Redis в service containers, сборка образа. Бейдж статуса в README, последний прогон зелёный.
- **`/metrics`**: HTTP-метрики (готово) плюс 2-4 бизнес-метрики своей темы с осмысленными метками без высокой кардинальности.
- **`/healthz` и `/readyz`** (готово): liveness без зависимостей, readiness проверяет базу и Redis параллельно с таймаутом.
- **Graceful shutdown** (готово для HTTP): на SIGTERM `/readyz` отвечает 503, через паузу останавливается HTTP-сервер, затем River `Stop` дожидается текущих задач, потом закрывается пул.
- **Логи `slog`** в JSON с `request_id`, без паролей, токенов и query-строк. `request_id` доходит до логов фоновых задач, которые запрос поставил.

### README

- Что делает сервис, в двух абзацах.
- **Архитектурная схема**: компоненты, хранилища, фоновые процессы, куда идут запросы. ASCII или Mermaid, главное, чтобы читалась за минуту. Каркас содержит заготовку.
- Запуск одной командой, примеры запросов `curl`, описание API (таблица или `api/openapi.yaml`).
- Ключевые решения и компромиссы: почему так, что выбрал бы при x100 нагрузке.
- Как запустить тесты и что именно они доказывают.

### Чего не нужно

Фронтенд, Kubernetes, микросервисы, Kafka, облако. Один сервис (плюс воркеры River в том же процессе) с хорошо продуманными границами сильнее пяти сервисов, склеенных наспех. Если на собесе спросят «а как бы ты разрезал», у тебя будет готовый ответ из README.

## Тема 1. Бронирование переговорок {#booking}

Офис, несколько переговорок, сотрудники бронируют их на время. Звучит просто, пока два человека не нажали «забронировать» на один слот в одну миллисекунду. Изюминка темы: корректность при конкурентных записях обеспечивает база, а не надежда на удачу, и фоновые задачи, привязанные ко времени (напоминания, автоотмена), которые нужно уметь отменять.

### User stories

- Как сотрудник, я ищу свободную переговорку на нужное время и число людей, чтобы не обходить их ногами.
- Как сотрудник, я бронирую комнату на интервал и получаю отказ с понятным сообщением, если её уже заняли.
- Как сотрудник, я вижу свои ближайшие брони и могу отменить любую из них.
- Как сотрудник, я получаю напоминание за 15 минут до начала встречи; если я отменил бронь, напоминание не приходит.
- Как офис-менеджер (роль `admin`), я завожу комнаты с вместимостью и оборудованием и вижу расписание комнаты на день.
- Как офис-менеджер, я хочу, чтобы бронь, на которую никто не пришёл, освобождалась автоматически («со звёздочкой»).

### Модель данных

```sql
CREATE EXTENSION IF NOT EXISTS btree_gist; -- уже в 00001_init.sql каркаса

CREATE TABLE users (
    id            bigserial   PRIMARY KEY,
    email         text        NOT NULL UNIQUE,
    password_hash text        NOT NULL,
    name          text        NOT NULL,
    role          text        NOT NULL DEFAULT 'user' CHECK (role IN ('user', 'admin')),
    created_at    timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE rooms (
    id         bigserial   PRIMARY KEY,
    name       text        NOT NULL UNIQUE,
    capacity   int         NOT NULL CHECK (capacity > 0),
    equipment  text[]      NOT NULL DEFAULT '{}',
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE bookings (
    id              bigserial   PRIMARY KEY,
    room_id         bigint      NOT NULL REFERENCES rooms (id),
    user_id         bigint      NOT NULL REFERENCES users (id),
    title           text        NOT NULL CHECK (length(title) BETWEEN 1 AND 200),
    during          tstzrange   NOT NULL,
    status          text        NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'cancelled')),
    reminder_job_id bigint,     -- id задачи River, чтобы отменить напоминание
    reminder_sent_at timestamptz, -- письмо отправлено: повтор задачи его не шлёт
    created_at      timestamptz NOT NULL DEFAULT now(),
    CHECK (NOT isempty(during) AND NOT lower_inf(during) AND NOT upper_inf(during)),
    CHECK (upper(during) - lower(during) <= interval '4 hours'),
    -- Две активные брони одной комнаты не пересекаются. Отменённые не мешают.
    CONSTRAINT bookings_no_overlap
        EXCLUDE USING gist (room_id WITH =, during WITH &&) WHERE (status = 'active')
);
CREATE INDEX bookings_user_start ON bookings (user_id, lower(during));
```

Что здесь важно. `tstzrange` хранит полуоткрытый интервал `[начало, конец)`, поэтому брони 10:00-11:00 и 11:00-12:00 не пересекаются, а 10:00-11:00 и 10:30-10:45 пересекаются. `EXCLUDE` проверяет, что для любых двух активных строк не выполняется одновременно «та же комната» и «интервалы пересекаются». Оператор `=` для `bigint` в GiST-индексе даёт расширение `btree_gist`. Нарушение приходит как ошибка `23P01` (`exclusion_violation`), и `repository.MapErr` в каркасе уже переводит её в `service.ErrConflict`. Из Go интервал удобно собирать в SQL: `tstzrange($3, $4, '[)')`, передавая два `time.Time`.

### API

Все маршруты под `/api/v1`, кроме регистрации и логина требуют `Authorization: Bearer <access>`. Аутентификацию (register, login, JWT с `sub` и `role`) перенеси из трекера.

| Метод и путь | Что делает | Ответы |
|---|---|---|
| `POST /auth/register`, `POST /auth/login` | как в трекере | 201, 200, 401, 409, 422 |
| `POST /rooms` (admin) | завести комнату | 201, 403, 409, 422 |
| `GET /rooms` | все комнаты | 200 |
| `GET /rooms/available?from=&to=&capacity=` | свободные на интервал | 200, 422 |
| `GET /rooms/:id/schedule?date=2026-11-10` | расписание комнаты на день | 200, 404 |
| `POST /bookings` | забронировать | 201, 404, 409, 422 |
| `GET /bookings?from=` | мои брони, начиная с даты, по возрастанию начала | 200 |
| `DELETE /bookings/:id` | отменить свою бронь | 204, 404, 409 |

Бронирование:

```http
POST /api/v1/bookings
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
Content-Type: application/json

{"room_id": 1, "title": "Планёрка", "start": "2026-11-10T10:00:00+03:00", "end": "2026-11-10T11:00:00+03:00"}
```

```http
HTTP/1.1 201 Created
Content-Type: application/json

{"id": 17, "room_id": 1, "title": "Планёрка", "start": "2026-11-10T07:00:00Z", "end": "2026-11-10T08:00:00Z", "status": "active"}
```

Слот уже занят:

```http
HTTP/1.1 409 Conflict

{"error": {"code": "conflict", "message": "комната занята в это время"}}
```

Нарушены правила времени:

```http
HTTP/1.1 422 Unprocessable Entity

{"error": {"code": "validation_failed", "message": "проверь поля запроса", "fields": {"end": "бронь длиннее 4 часов"}}}
```

Поиск свободных комнат:

```http
GET /api/v1/rooms/available?from=2026-11-10T10:30:00%2B03:00&to=2026-11-10T11:00:00%2B03:00&capacity=4
```

```json
{"items": [{"id": 2, "name": "Онега", "capacity": 4, "equipment": ["tv"]}]}
```

Плюс в query-строке кодируй как `%2B`, иначе он превратится в пробел и время не разберётся.

### Правила

- Начало в будущем, конец позже начала, оба кратны 15 минутам, длительность не больше 4 часов, не дальше 60 дней вперёд. Нарушение: 422 с полем.
- Отменить можно только свою активную бронь, которая ещё не началась. Чужая или несуществующая: 404, уже начавшаяся или отменённая: 409.
- Отмена ставит `status = 'cancelled'` (история сохраняется), освобождает слот и отменяет задачу напоминания в той же транзакции.
- Комнаты заводит только `admin`. Эндпоинта «сделай меня админом» в API нет и быть не должно: первого админа назначаешь в базе (`UPDATE users SET role = 'admin' ...`, команда в чеклисте), дальше роль приходит в JWT при логине.

### Нефункциональные требования

- **Ни одной двойной брони** при любом числе параллельных запросов. Доказывается интеграционным тестом и ручной проверкой из чеклиста.
- `POST /bookings` и `GET /rooms/available` укладываются в p95 < 50 мс на локальной машине при 10 комнатах и 10 000 броней (смотри гистограмму `http_request_duration_seconds` в `/metrics`).
- Напоминание уходит не раньше чем за 15 минут до начала и не позже начала. После отправки воркер ставит `reminder_sent_at`, и повтор задачи для уже отмеченной брони письмо не шлёт. Окно «письмо ушло, отметка не записалась» (процесс упал между ними) остаётся: это at-least-once, редкий дубль письма допустим, потеря нет.
- Бизнес-метрики: `bookings_created_total`, `booking_conflicts_total`, `reminders_sent_total`.

### Этапы по неделям

Срок темы: 4 недели.

<Hw id="booking-w1" title="Неделя 1: модель, комнаты, аутентификация" level="база">

Миграции `users`, `rooms`, `bookings` со всеми ограничениями из модели. Перенеси регистрацию, логин и middleware аутентификации из трекера, добавь проверку роли `admin`. Комнаты: создание (admin), список. Первого админа назначь через `make psql` (`UPDATE users SET role = 'admin' WHERE email = ...`). Тесты: репозиторий комнат в настоящей базе, хендлеры с fake-сервисом (401 без токена, 403 не админу, 422 на `capacity: 0`).

```bash
make migrate && make test
docker compose exec postgres psql -U diploma diploma -c '\d bookings'   # видно bookings_no_overlap
```

</Hw>

<Hw id="booking-w2" title="Неделя 2: брони без гонок" level="база">

Создание, список своих, отмена, поиск свободных комнат, расписание комнаты на день. Правила времени в сервисе с `now func() time.Time` и табличными юнит-тестами (прошлое, конец раньше начала, не кратно 15 минутам, ровно 4 часа можно, 4 часа 15 минут нельзя). Главный тест недели: 20 горутин одновременно бронируют один слот в настоящей базе, ровно одна получает бронь, 19 получают `ErrConflict`, в таблице одна активная строка. Ещё тест: соседние слоты 10:00-11:00 и 11:00-12:00 оба проходят.

```bash
go test -race -count=3 -run Concurrent ./internal/repository/
```

</Hw>

<Hw id="booking-w3" title="Неделя 3: напоминания на River" level="посложнее">

Подключи River (инструкция в README каркаса, разбор в уроке [River: основы](/m09/l02-river-basics)): таблицы River мигрируются вместе с goose в `-migrate`, клиент стартует в `main.go`, останавливается до закрытия пула. При создании брони в той же транзакции `InsertTx` ставит `ReminderArgs{BookingID}` с `ScheduledAt` за 15 минут до начала и сохраняет id задачи в `reminder_job_id`. Отмена брони в той же транзакции зовёт `JobCancelTx`. Воркер перечитывает бронь: отменена или уже есть `reminder_sent_at`, значит ничего не делает (защита от гонки «отменили, пока задача уже выполнялась» и от повтора). Письмо отправляет интерфейс `Mailer`; реализация пишет в лог, в тестах fake. Тесты через `rivertest`: задача поставлена с нужным временем, после отмены задача в состоянии `cancelled`, воркер на отменённой брони и на брони с `reminder_sent_at` не зовёт `Mailer`.

```bash
go test -race -run Reminder ./internal/...
```

</Hw>

<Hw id="booking-w4" title="Неделя 4: продакшн и упаковка" level="посложнее">

Бизнес-метрики, `request_id` в аргументах или метаданных задачи и в логах воркера, остановка River в graceful shutdown. `make up` на чистой машине поднимает всё. CI зелёный. README: схема, решения (почему `EXCLUDE`, а не `SELECT ... FOR UPDATE`), `EXPLAIN ANALYZE` поиска свободных комнат, размер образа. Отрепетируй рассказ на 3 минуты (раздел в конце страницы).

```bash
make ci && make up && curl -s localhost:8080/readyz && make down
```

</Hw>

### Чеклист приёмки

Подготовка с нуля (нужен `jq`; команды для macOS и Linux). Даты считаются от текущего момента, поэтому чеклист не протухает:

```bash
make up
API=localhost:8080/api/v1
J=(-H 'Content-Type: application/json')
# секунды Unix в RFC 3339 (на macOS date -r, на Linux date -d @)
iso() { date -u -r "$1" +%FT%TZ 2>/dev/null || date -u -d "@$1" +%FT%TZ; }

for u in admin anna; do
  curl -s "${J[@]}" -d "{\"email\":\"$u@example.com\",\"password\":\"correct-horse-42\",\"name\":\"$u\"}" $API/auth/register
done
# первый админ: только через базу
docker compose exec -T postgres psql -U diploma diploma -c "UPDATE users SET role = 'admin' WHERE email = 'admin@example.com'"

login() { curl -s "${J[@]}" -d "{\"email\":\"$1\",\"password\":\"correct-horse-42\"}" $API/auth/login | jq -r .access_token; }
A=(-H "Authorization: Bearer $(login admin@example.com)" "${J[@]}")
H=(-H "Authorization: Bearer $(login anna@example.com)" "${J[@]}")

curl -s "${A[@]}" -d '{"name":"Байкал","capacity":8,"equipment":["tv"]}' $API/rooms   # id 1
curl -s "${A[@]}" -d '{"name":"Онега","capacity":4}' $API/rooms                        # id 2

# бронь комнаты $1 с $2 до $3 (секунды Unix), печатает код ответа
book() { curl -s -o /dev/null -w '%{http_code}\n' "${H[@]}" \
  -d "{\"room_id\":$1,\"title\":\"Встреча\",\"start\":\"$(iso $2)\",\"end\":\"$(iso $3)\"}" $API/bookings; }
S=$(( ( $(date +%s) / 900 + 7 * 96 ) * 900 ))   # через неделю, на сетке 15 минут
```

- [ ] `make ci` зелёный: gofmt, vet, golangci-lint, `go test -race`.
- [ ] `docker compose up -d --build --wait` с пустым томом поднимает базу, мигрирует и стартует API; `curl -s localhost:8080/readyz` отвечает `ready`.
- [ ] Обычный пользователь не может завести комнату: `curl -s -o /dev/null -w '%{http_code}' "${H[@]}" -d '{"name":"X","capacity":2}' $API/rooms` даёт 403.
- [ ] 20 параллельных броней одного слота дают ровно один 201:

```bash
(for i in $(seq 20); do book 1 $S $((S + 3600)) & done; wait) | sort | uniq -c
#   1 201
#  19 409
```

- [ ] Соседний слот бронируется, пересекающийся нет:

```bash
book 1 $((S + 3600)) $((S + 7200))   # 201
book 1 $((S + 1800)) $((S + 2700))   # 409
```

- [ ] Неверное время даёт 422 с `fields`:

```bash
book 1 $((S - 8 * 86400)) $((S - 8 * 86400 + 3600))   # в прошлом
book 1 $((S + 86400)) $((S + 86400 + 5 * 3600))       # длиннее 4 часов
book 1 $((S + 86400 + 600)) $((S + 86400 + 4200))     # не кратно 15 минутам
```

- [ ] Отмена чужой брони даёт 404, повторная отмена своей 409; после отмены слот снова свободен в `/rooms/available`.
- [ ] Напоминание: бронь на ближайший слот, до которого 15-30 минут; не позже чем через 15 минут в логах воркера строка отправки с тем же `request_id`, что у запроса создания (`docker compose logs api | grep <request_id>`). Вторая такая бронь, отменённая сразу, напоминания не даёт:

```bash
R=$(( ( $(date +%s) / 900 + 2 ) * 900 ))   # ближайший слот, до которого 15-30 минут
book 1 $R $((R + 1800)); book 2 $R $((R + 1800))
# отмени вторую через DELETE /bookings/:id, затем:
docker compose exec -T postgres psql -U diploma diploma -c \
  "SELECT b.id, b.status, j.state FROM bookings b JOIN river_job j ON j.id = b.reminder_job_id ORDER BY b.id DESC LIMIT 2"
# активная: scheduled, затем completed; отменённая: cancelled
```

- [ ] `curl -s localhost:8080/metrics | grep -E 'bookings_created_total|booking_conflicts_total'` показывает счётчики после проверки гонки.
- [ ] Остановка: `docker compose stop api`, затем `docker compose logs api | tail -3` показывает переход в остановку и «остановлен», `docker compose ps -a api` показывает `Exited (0)`.
- [ ] В README схема, запуск, решения, `EXPLAIN ANALYZE`, бейдж CI.

### Со звёздочкой

- **Автоотмена неявки**: `POST /bookings/:id/check-in` в первые 10 минут встречи; при создании брони River ставит задачу на `начало + 10 минут`, которая отменяет бронь без отметки и освобождает остаток слота.
- **Повторяющиеся брони** («каждый вторник в 10:00 на 8 недель»): все вхождения создаются в одной транзакции, при конфликте хотя бы одного ответ 409 со списком конфликтующих дат.
- **Ежедневный дайджест**: периодическая задача River в 9:00 по Москве шлёт каждому его брони на сегодня; уникальность по пользователю и дню, чтобы рестарт не дал два письма.
- **Сравнение подходов**: ветка с `SELECT ... FOR UPDATE` по строке комнаты вместо `EXCLUDE` и короткая заметка в README, что изменилось в коде, тестах и пропускной способности.

## Тема 2. Сервис фича-флагов {#flags}

Фича-флаг это выключатель в коде: `if flags.Enabled("new-checkout", userID)`. Сервис флагов хранит, у кого какой флаг включён, а другие сервисы компании спрашивают его на каждом запросе. Отсюда изюминка: путь чтения горячий (тысячи проверок в секунду), значит нужен кеш и его правильная инвалидация; внутренние клиенты ходят по gRPC; каждый ключ ограничен по частоте; статистика вычислений копится в Redis и сливается в PostgreSQL фоновой задачей так, чтобы ретрай не посчитал дважды.

### User stories

- Как разработчик, я завожу флаг в проекте и включаю его в `dev`, не трогая `prod`.
- Как продакт, я раскатываю флаг на 10%, потом на 50% пользователей, и тот, кому флаг уже включился на 10%, не теряет его на 50%.
- Как QA, я добавляю конкретных пользователей в allow-list, чтобы они видели фичу при любом проценте.
- Как сервис-клиент, я по gRPC узнаю значения всех флагов для пользователя за один вызов меньше чем за 10 мс.
- Как владелец проекта, я вижу, кто и когда менял флаг (аудит), и сколько раз флаг вычислялся включённым и выключенным по часам.
- Как команда эксплуатации, я хочу, чтобы взбесившийся клиент упёрся в лимит запросов, а не положил сервис.

### Модель данных

```sql
CREATE TABLE projects (
    id         bigserial   PRIMARY KEY,
    key        text        NOT NULL UNIQUE CHECK (key ~ '^[a-z0-9-]{2,40}$'),
    name       text        NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

-- Ключи доступа: admin управляет флагами, sdk только вычисляет. Хранится только SHA-256.
CREATE TABLE api_keys (
    id          bigserial   PRIMARY KEY,
    project_id  bigint      NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    environment text        NOT NULL CHECK (environment IN ('dev', 'staging', 'prod')),
    kind        text        NOT NULL CHECK (kind IN ('admin', 'sdk')),
    key_hash    bytea       NOT NULL UNIQUE,
    rps_limit   int         NOT NULL DEFAULT 100 CHECK (rps_limit > 0),
    revoked_at  timestamptz,
    created_at  timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE flags (
    id          bigserial   PRIMARY KEY,
    project_id  bigint      NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    key         text        NOT NULL CHECK (key ~ '^[a-z0-9_-]{2,60}$'),
    description text        NOT NULL DEFAULT '',
    created_at  timestamptz NOT NULL DEFAULT now(),
    UNIQUE (project_id, key)
);

-- Правила флага в окружении. version для оптимистичной блокировки.
CREATE TABLE flag_rules (
    flag_id         bigint      NOT NULL REFERENCES flags (id) ON DELETE CASCADE,
    environment     text        NOT NULL CHECK (environment IN ('dev', 'staging', 'prod')),
    enabled         boolean     NOT NULL DEFAULT false,
    rollout_percent int         NOT NULL DEFAULT 0 CHECK (rollout_percent BETWEEN 0 AND 100),
    allow_users     text[]      NOT NULL DEFAULT '{}',
    version         bigint      NOT NULL DEFAULT 1,
    updated_at      timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (flag_id, environment)
);

CREATE TABLE audit_log (
    id          bigserial   PRIMARY KEY,
    flag_id     bigint      NOT NULL REFERENCES flags (id) ON DELETE CASCADE,
    environment text        NOT NULL,
    actor_key   bigint      NOT NULL REFERENCES api_keys (id),
    before      jsonb,
    after       jsonb       NOT NULL,
    at          timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE evaluation_stats (
    flag_id     bigint      NOT NULL REFERENCES flags (id) ON DELETE CASCADE,
    environment text        NOT NULL,
    hour        timestamptz NOT NULL,
    enabled     boolean     NOT NULL,
    count       bigint      NOT NULL CHECK (count >= 0),
    PRIMARY KEY (flag_id, environment, hour, enabled)
);

-- Какие пачки статистики уже слиты: защита от двойного счёта при ретрае задачи.
CREATE TABLE stats_batches (
    batch_key  text        PRIMARY KEY,
    flushed_at timestamptz NOT NULL DEFAULT now()
);
```

### Как вычисляется флаг

Правила по порядку: флаг выключен в окружении, значит `false` (`reason: "disabled"`); пользователь в `allow_users`, значит `true` (`"allow_list"`); иначе корзина `fnv32a(flagKey + ":" + userID) % 100` меньше `rollout_percent` (`"rollout"`). Корзина пользователя постоянна, поэтому рост процента с 10 до 50 только добавляет пользователей, никого не выключая. Ключ флага в хеше нужен, чтобы на 10% разных флагов попадали разные люди. Вычисление это чистая функция без базы и Redis, на неё табличные тесты.

```go
func bucket(flagKey, userID string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(flagKey + ":" + userID))
	return h.Sum32() % 100
}
// bucket("new-checkout", "user-42") == 26: при 20% выключен, при 30% включён
```

### API

REST для управления (ключ `admin`), REST и gRPC для вычисления (ключ `sdk`). Ключ передаётся в `Authorization: Bearer ff_...`; проект и окружение определяются по ключу, а не по параметрам запроса.

| Метод и путь | Что делает | Ответы |
|---|---|---|
| `POST /api/v1/flags` | создать флаг (правила во всех окружениях выключены) | 201, 401, 403, 409, 422 |
| `GET /api/v1/flags` | флаги проекта с правилами окружения ключа | 200 |
| `PUT /api/v1/flags/:key/rules` | заменить правила в окружении ключа, с `version` | 200, 404, 409, 422 |
| `GET /api/v1/flags/:key/audit` | история изменений | 200, 404 |
| `GET /api/v1/flags/:key/stats?from=&to=` | вычисления по часам | 200, 404 |
| `POST /api/v1/evaluate` | все флаги для пользователя (ключ `sdk`) | 200, 401, 429 |
| gRPC `flags.v1.FlagService/Evaluate`, `EvaluateAll` | то же для внутренних сервисов | `OK`, `Unauthenticated`, `ResourceExhausted` |

Изменение правил с оптимистичной блокировкой:

```http
PUT /api/v1/flags/new-checkout/rules
Authorization: Bearer ff_admin_prod_...

{"enabled": true, "rollout_percent": 30, "allow_users": ["qa-1"], "version": 3}
```

```http
HTTP/1.1 200 OK

{"key": "new-checkout", "environment": "prod", "enabled": true, "rollout_percent": 30, "allow_users": ["qa-1"], "version": 4}
```

Кто-то успел поменять правила раньше (в базе уже `version = 4`):

```http
HTTP/1.1 409 Conflict

{"error": {"code": "conflict", "message": "правила изменились, перечитай и повтори"}}
```

Вычисление:

```http
POST /api/v1/evaluate
Authorization: Bearer ff_sdk_prod_...

{"user_id": "user-42"}
```

```json
{"flags": {"new-checkout": {"enabled": true, "reason": "rollout"}, "dark-mode": {"enabled": false, "reason": "disabled"}}}
```

Превышен лимит ключа:

```http
HTTP/1.1 429 Too Many Requests
Retry-After: 1

{"error": {"code": "rate_limited", "message": "слишком много запросов"}}
```

Контракт gRPC (`api/proto/flags/v1/flags.proto`, генерация через buf, как в [уроке про gRPC](/m10/l03-grpc)):

```proto
syntax = "proto3";
package flags.v1;

service FlagService {
  rpc Evaluate(EvaluateRequest) returns (EvaluateResponse);
  rpc EvaluateAll(EvaluateAllRequest) returns (EvaluateAllResponse);
}

message EvaluateRequest {
  string flag_key = 1;
  string user_id = 2;
}
message EvaluateResponse {
  bool enabled = 1;
  string reason = 2;
}
message EvaluateAllRequest { string user_id = 1; }
message EvaluateAllResponse { map<string, EvaluateResponse> flags = 1; }
```

### Правила

- Ключ это строка `ff_<вид>_` плюс 64 hex-символа из `crypto/rand`, в базе только `sha256` от всей строки. Отозванный ключ: 401. Первый проект и первые ключи заводишь в базе (команда в чеклисте), выпуск ключей через API со звёздочкой.
- Изменение правил и запись в `audit_log` (до и после в `jsonb`) в одной транзакции, удаление ключа кеша после `COMMIT`. Если удалить до коммита, параллельное чтение успеет положить в кеш старые правила.
- Кеш: правила проекта и окружения в Redis одним ключом `ff:rules:<project>:<env>` с TTL 60 секунд. Промах: чтение из базы и запись в кеш (cache-aside). Даже с удалением после коммита остаётся узкая гонка: читатель на промахе прочитал правила до коммита, писатель закоммитил и удалил ключ, читатель записал в кеш старое. Такое значение живёт до TTL, поэтому TTL короткий; точное решение (версия в кеше) со звёздочкой.
- Redis недоступен: вычисление идёт в базу, сервис отвечает медленнее, но отвечает; `/readyz` показывает `redis: fail`.
- Лимит: фиксированное окно в 1 секунду на ключ, `INCR rl:<key_id>:<unix_sec>` и `EXPIRE 2` в одном pipeline; сверх `rps_limit` ответ 429 с `Retry-After: 1` или `codes.ResourceExhausted`. Лимитер получает `now func() time.Time`, чтобы тест не зависел от границы секунды.
- Статистика: каждое вычисление делает `HINCRBY ff:stats:pending <flag_id>:<env>:<час>:<enabled> 1`. Периодическая задача River раз в минуту сливает счётчики в базу так, чтобы ни повтор задачи, ни падение процесса в любой точке не потеряли и не удвоили их. Ключевая идея: у каждой забранной пачки своё уникальное имя, и это имя служит ключом идемпотентности в базе.
  1. Сначала добить незавершённые пачки: `SCAN 0 MATCH ff:stats:flush:*` и для каждого найденного ключа шаги 3-5. Так подбираются и собственные прошлые попытки этой задачи, и сироты задач, исчерпавших все попытки.
  2. `RENAME ff:stats:pending ff:stats:flush:<uuid>`, где `uuid` новый на каждое переименование (16 байт из `crypto/rand`). Ответ `no such key` значит, что новых счётчиков нет. Имя по `job_id` здесь ошибка: повтор задачи получил бы то же имя, и после падения за `DEL` свежие счётчики попали бы в пачку, которую база уже считает слитой.
  3. `HGETALL` пачки.
  4. Одна транзакция: `INSERT INTO stats_batches (batch_key) VALUES ('ff:stats:flush:<uuid>') ON CONFLICT DO NOTHING`. Вставилась строка: upsert всех полей в `evaluation_stats`. Не вставилась: эту пачку уже слили, upsert пропускается.
  5. После `COMMIT` `DEL` пачки.

  Два запуска одновременно могут взять одну и ту же сироту из `SCAN`: строка в `stats_batches` пропустит в базу только первого.

### Нефункциональные требования

- `EvaluateAll` по gRPC при прогретом кеше: p95 < 10 мс локально на проекте с 50 флагами.
- После `PUT .../rules` новые правила видны в следующем вычислении; в редкой гонке чтения и записи (см. правила) не позже TTL кеша, 60 секунд.
- Счётчики статистики сходятся с числом вычислений точно: повтор задачи слива и падение в любой точке не теряют и не удваивают их.
- Бизнес-метрики: `flag_evaluations_total{environment, reason}`, `flag_cache_requests_total{result="hit|miss"}`, `rate_limited_total`.

### Этапы по неделям

Срок темы: 5 недель. Если время кончается, в «со звёздочкой» можно перенести `GET .../stats` (итог проверяется SQL-запросом) и тесты gRPC на `bufconn` сверх одного кейса на interceptor.

<Hw id="flags-w1" title="Неделя 1: проекты, ключи, флаги, аудит" level="база">

Миграции всех таблиц. Первый проект и ключи заводишь в базе командой из чеклиста. Middleware аутентификации по ключу: `sha256` строки, поиск, проверка `revoked_at` и вида ключа, проект и окружение в контексте запроса. Создание и список флагов, `PUT .../rules` с `version` (условный `UPDATE ... WHERE version = $n`, 0 строк значит 409) и аудит в той же транзакции. Тесты: репозиторий в базе (две параллельные правки с одной `version`, успешна ровно одна), хендлеры (401, 403 для `sdk` на запись, 409, 422 на `rollout_percent: 101`).

```bash
make migrate && make test
```

</Hw>

<Hw id="flags-w2" title="Неделя 2: вычисление и кеш" level="база">

Чистая функция вычисления и табличные тесты: выключен, allow-list при 0%, граница процента (корзина 26 при 26% выключена, при 27% включена), монотонность (кто включён на 10%, включён и на 50%), распределение на 10 000 пользователей при 30% в пределах 27-33%. `POST /evaluate` через cache-aside в Redis, удаление ключа после коммита правки. Интеграционный тест с настоящим Redis: промах, попадание, новые правила после `PUT`, работа при недоступном Redis.

```bash
go test -race -run 'Evaluate|Bucket|Cache' ./internal/...
```

</Hw>

<Hw id="flags-w3" title="Неделя 3: лимиты и статистика" level="посложнее">

Rate limiter в Redis с `now func() time.Time`; тест: при лимите 5 и замороженном времени шестой запрос получает отказ, через секунду снова пропускает. Счётчики вычислений в Redis и периодическая задача River для слива по алгоритму из правил (периодические и уникальные задачи: [River: ретраи, уникальность, периодика](/m09/l03-river-advanced)). Главный тест недели: для каждой точки падения (после `RENAME`, до `COMMIT`, после `COMMIT` до `DEL`, после `DEL` до завершения задачи) первая попытка падает, между попытками приходят новые вычисления, после повтора той же задачи и следующего запуска сумма в `evaluation_stats` равна числу вычислений, ключей `ff:stats:*` не осталось. Отдельно: два запуска подряд без новых вычислений ничего не удваивают.

```bash
go test -race -run 'RateLimit|Flush' ./internal/...
```

</Hw>

<Hw id="flags-w4" title="Неделя 4: gRPC" level="посложнее">

`FlagService` на gRPC на отдельном порту (`GRPC_PORT`), unary interceptor достаёт ключ из metadata `authorization`, второй interceptor проверяет лимит. Тот же сервисный слой, что у REST. Тесты через `bufconn`: `Unauthenticated` без ключа, `ResourceExhausted` сверх лимита, совпадение с REST.

```bash
go test -race -run GRPC ./internal/...
```

</Hw>

<Hw id="flags-w5" title="Неделя 5: продакшн и упаковка" level="посложнее">

`GET .../stats` по часам, бизнес-метрики, graceful shutdown обоих серверов (`GracefulStop` у gRPC) и River. Compose публикует оба порта. CI зелёный. README: схема с путями чтения и записи, почему cache-aside и TTL, что будет при падении Redis, как устроены раскат и слив статистики, размер образа.

```bash
make ci && make up && curl -s localhost:8080/readyz && make down
```

</Hw>

### Чеклист приёмки

Подготовка с нуля: проект и ключи заводятся в базе, ключи генерирует `openssl`, в базу попадает только хеш.

```bash
make up
API=localhost:8080/api/v1
ADMIN=ff_admin_$(openssl rand -hex 32)
SDK=ff_sdk_$(openssl rand -hex 32)
docker compose exec -T postgres psql -U diploma diploma <<SQL
INSERT INTO projects (key, name) VALUES ('shop', 'Магазин');
INSERT INTO api_keys (project_id, environment, kind, key_hash) VALUES
  (1, 'prod', 'admin', sha256(convert_to('$ADMIN', 'UTF8'))),
  (1, 'prod', 'sdk',   sha256(convert_to('$SDK', 'UTF8')));
SQL
A=(-H "Authorization: Bearer $ADMIN" -H 'Content-Type: application/json')
K=(-H "Authorization: Bearer $SDK" -H 'Content-Type: application/json')
curl -s "${A[@]}" -d '{"key":"new-checkout","description":"Новый checkout"}' $API/flags
rules() { curl -s "${A[@]}" -X PUT -d "$1" $API/flags/new-checkout/rules; echo; }
evalu() { curl -s "${K[@]}" -d "{\"user_id\":\"$1\"}" $API/evaluate | jq -c '.flags["new-checkout"]'; }
```

`sha256(convert_to(...))` в PostgreSQL даёт те же байты, что `sha256.Sum256([]byte(key))` в Go.

- [ ] `make ci` зелёный, `docker compose up -d --build --wait` поднимает всё с пустым томом.
- [ ] Оптимистичная блокировка и аудит: первая правка с `version: 1` проходит (в ответе `version: 2`), повтор с `version: 1` даёт 409; в `GET $API/flags/new-checkout/audit` видна одна правка с `before` и `after`.

```bash
rules '{"enabled":true,"rollout_percent":30,"allow_users":[],"version":1}'
rules '{"enabled":true,"rollout_percent":30,"allow_users":[],"version":1}'   # 409
```

- [ ] Раскат и allow-list:

```bash
evalu user-42                                                                  # {"enabled":true,"reason":"rollout"}
rules '{"enabled":true,"rollout_percent":20,"allow_users":[],"version":2}'
evalu user-42                                                                  # {"enabled":false,"reason":"rollout"}
rules '{"enabled":true,"rollout_percent":0,"allow_users":["qa-1"],"version":3}'
evalu qa-1                                                                     # {"enabled":true,"reason":"allow_list"}
```

- [ ] Кеш: после `PUT` следующее вычисление отдаёт новые правила; в `curl -s localhost:8080/metrics | grep flag_cache_requests_total` после правки растёт `miss`, на повторных вычислениях `hit`.
- [ ] `docker compose stop redis`: `/readyz` 503 с `redis: fail`, `evalu user-42` всё ещё отвечает из базы; после `docker compose start redis` кеш снова работает.
- [ ] gRPC: `grpcurl -plaintext -H "authorization: Bearer $SDK" -d '{"user_id":"qa-1"}' localhost:9090 flags.v1.FlagService/EvaluateAll` отвечает теми же значениями, что REST (нужен reflection на сервере или `-proto`).
- [ ] Статистика сходится: запомни сумму, сделай 50 вычислений, через 2 минуты сумма выросла ровно на 50 (флаг в проекте один); `GET $API/flags/new-checkout/stats` показывает те же числа.

```bash
sum() { docker compose exec -T postgres psql -U diploma diploma -tAc "SELECT coalesce(sum(count), 0) FROM evaluation_stats"; }
sleep 70; B=$(sum); for i in $(seq 50); do evalu u$i >/dev/null; done; sleep 120; echo $(( $(sum) - B ))   # 50
```

- [ ] Лимит: ключу ставится `rps_limit = 5`, из 20 одновременных запросов часть получает 429 с `Retry-After`, успешных не больше 10 (пачка может попасть на границу двух секундных окон).

```bash
docker compose exec -T postgres psql -U diploma diploma -c "UPDATE api_keys SET rps_limit = 5 WHERE kind = 'sdk'"
(for i in $(seq 20); do curl -s -o /dev/null -w '%{http_code}\n' "${K[@]}" -d '{"user_id":"u"}' $API/evaluate & done; wait) | sort | uniq -c
curl -s -i "${K[@]}" -d '{"user_id":"u"}' $API/evaluate | grep -i -E '^HTTP|retry-after'   # сразу после пачки
```

- [ ] В базе нет открытых ключей, только `key_hash`; в логах нет заголовка `Authorization`.
- [ ] Остановка: `docker compose stop api` даёт в логах «остановлен» и `Exited (0)` в `docker compose ps -a api`.
- [ ] README со схемой, решениями, бейджем CI.

### Со звёздочкой

- **`WatchFlags`**: server streaming в gRPC, клиент получает изменения правил своего окружения сразу. Изменения разносятся между инстансами через Redis Pub/Sub, на каждый `PUT` после коммита публикуется событие.
- **Кеш без гонки**: в значении кеша лежит `version` правил, запись в Redis через Lua-скрипт «записать, только если значения нет или его версия меньше»; писатель после коммита не удаляет ключ, а кладёт новые правила тем же скриптом. Старое значение больше не может перезаписать новое.
- **Выпуск ключей через API**: `POST /api/v1/keys` (ключ `admin`), ключ показывается один раз, `DELETE` ставит `revoked_at`.
- **Локальный кеш в процессе** поверх Redis с инвалидацией через тот же Pub/Sub; бенчмарк `EvaluateAll` с ним и без.
- **Таргетинг по атрибутам**: правила вида `country in ["RU","KZ"]` с атрибутами пользователя в запросе; правила хранятся в `jsonb`.
- **Маленький Go SDK** в `pkg/client`: кеширует ответы на N секунд и не падает, если сервис флагов недоступен (значения по умолчанию).

## Тема 3. Заказы с outbox и вебхуками {#orders}

Интернет-магазин без витрины: каталог, заказ, оплата (имитация), отгрузка, отмена. На каждое изменение заказа покупатель получает письмо, а внешние системы (склад, CRM) получают вебхук. Изюминка: изменение заказа и обещание «сообщить об этом» должны быть атомарны. Если записать заказ и затем отправить сообщение отдельно, рано или поздно одно случится без другого (это называется dual write). Транзакционный outbox решает это: событие пишется в ту же базу в той же транзакции, а доставляет его фоновый процесс. River хранит задачи в PostgreSQL, поэтому `InsertTx` и есть готовый outbox. Вторая половина изюминки: идемпотентность на входе (повтор `POST` не создаёт второй заказ) и на выходе (получатель вебхука может получить событие дважды и должен это распознать).

### User stories

- Как покупатель, я оформляю заказ из нескольких товаров; если сеть оборвалась и приложение повторило запрос, второй заказ не создаётся и деньги не списываются дважды.
- Как покупатель, я не могу купить товар, которого нет на складе, даже если последний экземпляр покупают одновременно со мной.
- Как покупатель, я оплачиваю заказ, вижу его статус и историю, могу отменить неоплаченный.
- Как покупатель, я получаю письмо о создании, оплате, отгрузке и отмене заказа.
- Как интегратор, я регистрирую URL вебхука и получаю подписанные события о моих заказах; при моей недоступности события доставляются позже, а не теряются.
- Как администратор, я отмечаю отгрузку и вижу в метриках, насколько отстаёт доставка событий.

### Модель данных

```sql
CREATE TABLE products (
    id         bigserial   PRIMARY KEY,
    sku        text        NOT NULL UNIQUE,
    name       text        NOT NULL,
    price_kop  bigint      NOT NULL CHECK (price_kop > 0),
    stock      int         NOT NULL CHECK (stock >= 0),
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE orders (
    id              uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         bigint      NOT NULL REFERENCES users (id),
    status          text        NOT NULL DEFAULT 'pending'
                                CHECK (status IN ('pending', 'paid', 'shipped', 'cancelled')),
    total_kop       bigint      NOT NULL CHECK (total_kop > 0),
    idempotency_key text        NOT NULL CHECK (length(idempotency_key) BETWEEN 8 AND 64),
    request_hash    bytea       NOT NULL, -- SHA-256 тела запроса: тот же ключ с другим телом это ошибка клиента
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    UNIQUE (user_id, idempotency_key)
);
CREATE INDEX orders_user_created ON orders (user_id, created_at DESC);

CREATE TABLE order_items (
    order_id   uuid   NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    product_id bigint NOT NULL REFERENCES products (id),
    quantity   int    NOT NULL CHECK (quantity > 0),
    price_kop  bigint NOT NULL CHECK (price_kop > 0), -- цена на момент заказа
    PRIMARY KEY (order_id, product_id)
);

CREATE TABLE order_events (
    id          bigserial   PRIMARY KEY,
    order_id    uuid        NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    from_status text,
    to_status   text        NOT NULL,
    at          timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE webhooks (
    id         bigserial   PRIMARY KEY,
    user_id    bigint      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    url        text        NOT NULL CHECK (url ~ '^https?://'),
    secret     text        NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE webhook_deliveries (
    id          bigserial   PRIMARY KEY,
    webhook_id  bigint      NOT NULL REFERENCES webhooks (id) ON DELETE CASCADE,
    event_id    text        NOT NULL,
    attempt     int         NOT NULL,
    status_code int,
    error       text,
    duration_ms int         NOT NULL,
    at          timestamptz NOT NULL DEFAULT now()
);
```

Таблица `users` как в трекере. Отдельной таблицы outbox нет: её роль играет `river_job`.

### Ключевые запросы

Цены читаются в начале транзакции, до вставки заказа: из них считаются `total_kop` (у `orders` это `NOT NULL CHECK (> 0)`) и `order_items.price_kop`.

```sql
SELECT id, price_kop FROM products WHERE id = ANY($1);
-- найдено меньше строк, чем товаров в запросе: неизвестный товар, 422
```

Идемпотентная вставка заказа идёт до резерва: повтор с тем же ключом останавливается здесь и не трогает остатки.

```sql
INSERT INTO orders (user_id, total_kop, idempotency_key, request_hash)
VALUES ($1, $2, $3, $4)
ON CONFLICT (user_id, idempotency_key) DO NOTHING
RETURNING id;
-- 0 строк: заказ с этим ключом уже есть; ROLLBACK, прочитай его и сравни request_hash
```

Резерв остатка без гонок одним условным `UPDATE`: строка блокируется, условие проверяется уже на заблокированной версии, поэтому два покупателя последнего экземпляра не уйдут в минус (и `CHECK (stock >= 0)` страхует). Товары в заказе обновляй в порядке возрастания `product_id`: если два заказа берут одни и те же товары в разном порядке, они могут заблокировать друг друга и один получит deadlock.

```sql
UPDATE products SET stock = stock - $2 WHERE id = $1 AND stock >= $2;
-- 0 строк: товара не хватает, ROLLBACK всей транзакции (вместе с заказом) и 409
```

Переход статуса тоже условный: `UPDATE orders SET status = 'paid', updated_at = now() WHERE id = $1 AND user_id = $2 AND status = 'pending'`. Ноль строк значит заказ чужой, не найден (404) или в другом статусе (409): различи это повторным чтением.

### Что происходит в одной транзакции

```text
POST /orders (Idempotency-Key: k1)
BEGIN
  SELECT id, price_kop FROM products ...         цены и total_kop; неизвестный товар: 422
  INSERT orders ... ON CONFLICT DO NOTHING       повтор: ROLLBACK, вернуть первый ответ
  UPDATE products ... по возрастанию id          не хватило: ROLLBACK, 409
  INSERT order_items (цены из SELECT), order_events
  river.InsertTx(EmailArgs{order, "created"})
  river.InsertTx(WebhookArgs{order, "order.created", event_id}) на каждый вебхук пользователя
  river.InsertTx(AutoCancelArgs{order}, ScheduledAt: +30 минут)
COMMIT                                           заказ и обещания доставить появляются вместе
```

Цена могла измениться между `SELECT` и `COMMIT`: заказ фиксирует ту, что видел покупатель в момент оформления, это нормально. Если нужна строгая цена, читай её `SELECT ... FOR SHARE`.

### API

| Метод и путь | Что делает | Ответы |
|---|---|---|
| `POST /api/v1/products` (admin), `GET /api/v1/products` | каталог | 201, 200, 403, 409, 422 |
| `POST /api/v1/orders` с `Idempotency-Key` | создать заказ | 201, 400, 409, 422 |
| `GET /api/v1/orders`, `GET /api/v1/orders/:id` | мои заказы, заказ с позициями и историей | 200, 404 |
| `POST /api/v1/orders/:id/pay` | имитация оплаты | 200, 404, 409 |
| `POST /api/v1/orders/:id/cancel` | отмена неоплаченного, остатки возвращаются | 200, 404, 409 |
| `POST /api/v1/orders/:id/ship` (admin) | отгрузка оплаченного | 200, 403, 404, 409 |
| `POST /api/v1/webhooks` `{"url","secret"}`, `GET`, `DELETE /api/v1/webhooks/:id` | подписки | 201, 200, 204, 422 |

Создание заказа:

```http
POST /api/v1/orders
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
Idempotency-Key: 9b2f6c1e-4d1a-4a7e-9a41-2f0c8d5e7b13
Content-Type: application/json

{"items": [{"product_id": 1, "quantity": 2}, {"product_id": 2, "quantity": 1}]}
```

```http
HTTP/1.1 201 Created

{"id": "0c6f1d4e-5b7a-4f0e-8a2b-9d3c1e7f6a55", "status": "pending", "total_kop": 229700,
 "items": [{"product_id": 1, "quantity": 2, "price_kop": 49900}, {"product_id": 2, "quantity": 1, "price_kop": 129900}],
 "created_at": "2026-11-10T07:00:00Z"}
```

Тот же запрос повторно с тем же ключом: тот же код 201, то же тело и заголовок `Idempotent-Replayed: true`, остатки не тронуты. Тот же ключ с другим телом:

```http
HTTP/1.1 422 Unprocessable Entity

{"error": {"code": "idempotency_key_reused", "message": "ключ уже использован с другим запросом"}}
```

Без заголовка `Idempotency-Key`: 400. Товара не хватает:

```http
HTTP/1.1 409 Conflict

{"error": {"code": "out_of_stock", "message": "товара недостаточно на складе", "fields": {"items[1].quantity": "доступно 0"}}}
```

Вебхук, который получает интегратор:

```http
POST https://crm.example.com/hooks/orders
Content-Type: application/json
X-Event-Id: order.paid:0c6f1d4e-5b7a-4f0e-8a2b-9d3c1e7f6a55
X-Signature: sha256=5d41402abc4b2a76b9719d911017c592...

{"event": "order.paid", "order_id": "0c6f1d4e-5b7a-4f0e-8a2b-9d3c1e7f6a55", "status": "paid", "occurred_at": "2026-11-10T07:05:12Z"}
```

`X-Signature` это HMAC-SHA256 тела на секрете подписки в hex. Получатель считает HMAC сам и сравнивает через `hmac.Equal` (постоянное время), а `X-Event-Id` использует для дедупликации: доставка at-least-once, повтор возможен.

### Правила

- Статусы: `pending -> paid | cancelled`, `paid -> shipped`. Остальные переходы: 409. Отмена возвращает остатки в той же транзакции.
- Неоплаченный заказ отменяется сам через 30 минут: отложенная задача River проверяет статус и отменяет только `pending` (оплаченный заказ задача не трогает, повтор задачи безопасен).
- Письмо и вебхук ставятся только через `InsertTx` внутри транзакции изменения. Отправка из хендлера после коммита запрещена: это и есть dual write.
- Подписка: `POST /webhooks` с `{"url", "secret"}`, секрет 16-64 символа задаёт интегратор и получает обратно только при создании.
- Воркер вебхука: задача ставится с `InsertOpts{MaxAttempts: 10}` (по умолчанию у River 25). Таймаут клиента 5 секунд, тело закрывается всегда; ответ 2xx значит доставлено; 5xx, 429 и сетевые ошибки значит вернуть ошибку, River повторит с backoff `попытка^4` секунд (1 с, 16 с, 81 с, ...); остальные 4xx значит `river.JobCancel` (повтор не поможет). Каждая попытка пишется в `webhook_deliveries`.
- Первого админа (каталог, отгрузка) назначаешь в базе, как в теме 1.
- URL вебхука только `http` или `https`; со звёздочкой запрет адресов из приватных сетей (SSRF).

### Нефункциональные требования

- Ни одного заказа без событий и ни одного события без заказа: проверяется тестом, который роняет транзакцию после `InsertTx` и убеждается, что в `river_job` пусто.
- 50 параллельных заказов на товар с остатком 10 дают ровно 10 успешных, остаток 0, ни одного отрицательного.
- Задержка доставки события (от коммита до первой попытки) p95 < 2 секунд при работающем получателе.
- Бизнес-метрики: `orders_created_total`, `orders_status_changes_total{to}`, `webhook_deliveries_total{result}`, `webhook_delivery_duration_seconds`, и gauge `outbox_lag_seconds`: возраст самой старой задачи в состоянии `available` (раз в 15 секунд запросом к `river_job`).

### Этапы по неделям

Срок темы: 5 недель. Если время кончается, в «со звёздочкой» можно перенести `outbox_lag_seconds` и письма об отгрузке и отмене (письмо о создании и все вебхуки остаются).

<Hw id="orders-w1" title="Неделя 1: каталог и идемпотентный заказ" level="база">

Миграции, аутентификация из трекера с ролью `admin`, каталог. Создание заказа в одной транзакции в порядке из схемы выше: цены, идемпотентная вставка, резерв остатков по возрастанию `product_id`, позиции, первое событие в `order_events`. Повтор с тем же ключом отдаёт первый ответ, с другим телом 422, без ключа 400. Тесты в базе: 50 горутин покупают товар с остатком 10, успешных ровно 10; 10 горутин шлют один и тот же запрос с одним ключом, заказ один, остаток уменьшен один раз.

```bash
go test -race -count=3 -run 'Concurrent|Idempotent' ./internal/...
```

</Hw>

<Hw id="orders-w2" title="Неделя 2: статусы и outbox" level="база">

Оплата, отгрузка, отмена условными `UPDATE`, история в `order_events`, возврат остатков при отмене. River: письма (`Mailer` пишет в лог, в тестах fake; почему outbox и at-least-once, в уроке [Зачем очереди](/m09/l01-why-queues)) и автоотмена через 30 минут, всё через `InsertTx`. Тесты через `rivertest`: после создания заказа поставлены письмо и автоотмена с нужным `ScheduledAt`; транзакция, упавшая после `InsertTx`, не оставляет задач; автоотмена оплаченного заказа ничего не меняет.

```bash
go test -race -run 'Status|Outbox|AutoCancel' ./internal/...
```

</Hw>

<Hw id="orders-w3" title="Неделя 3: доставка вебхуков" level="посложнее">

Подписки, воркер доставки с подписью HMAC, таймаутом, `MaxAttempts: 10`, ретраями и `JobCancel` на 4xx, журнал попыток. Тесты воркера против `httptest.Server`: правильная подпись, 503 возвращает ошибку (River повторит), 400 отменяет задачу, медленный получатель обрывается по таймауту.

```bash
go test -race -run Webhook ./internal/...
```

</Hw>

<Hw id="orders-w4" title="Неделя 4: получатель и наблюдаемость" level="посложнее">

Маленький `cmd/webhook-echo`: на `/hook` проверяет подпись и отвечает 200, а с флагом `-fail 3` первые три запроса отвечает 503; на `/bad` всегда 400. Отдельным сервисом `webhook-echo` в compose (тот же Dockerfile с `--target`-стадией или второй `go build` в стадии сборки). Бизнес-метрики и `outbox_lag_seconds`, `request_id` из запроса в метаданных задачи и в логах воркеров, остановка River в graceful shutdown.

```bash
docker compose up -d --build --wait && docker compose logs webhook-echo
```

</Hw>

<Hw id="orders-w5" title="Неделя 5: продакшн и упаковка" level="посложнее">

`make up` на чистой машине поднимает всё, CI зелёный. README: схема с транзакцией и воркерами, почему outbox, а не отправка из хендлера, почему `UPDATE ... WHERE stock >= n`, гарантии доставки и что должен делать получатель, размер образа.

```bash
make ci && make up && curl -s localhost:8080/metrics | grep -E 'orders_created_total|outbox_lag' && make down
```

</Hw>

### Чеклист приёмки

Подготовка с нуля: пользователи и первый админ как в теме 1, товары заводит админ.

```bash
make up
API=localhost:8080/api/v1
J=(-H 'Content-Type: application/json')
for u in admin anna; do
  curl -s "${J[@]}" -d "{\"email\":\"$u@example.com\",\"password\":\"correct-horse-42\"}" $API/auth/register
done
docker compose exec -T postgres psql -U diploma diploma -c "UPDATE users SET role = 'admin' WHERE email = 'admin@example.com'"
login() { curl -s "${J[@]}" -d "{\"email\":\"$1\",\"password\":\"correct-horse-42\"}" $API/auth/login | jq -r .access_token; }
A=(-H "Authorization: Bearer $(login admin@example.com)" "${J[@]}")
H=(-H "Authorization: Bearer $(login anna@example.com)" "${J[@]}")
curl -s "${A[@]}" -d '{"sku":"MUG-01","name":"Кружка","price_kop":49900,"stock":10}' $API/products    # id 1, для гонки
curl -s "${A[@]}" -d '{"sku":"TEE-02","name":"Футболка","price_kop":129900,"stock":100}' $API/products # id 2
order() { curl -s -o /dev/null -w '%{http_code}\n' "${H[@]}" -H "Idempotency-Key: $(uuidgen)" \
  -d "{\"items\":[{\"product_id\":$1,\"quantity\":1}]}" $API/orders; }
```

- [ ] `make ci` зелёный, `docker compose up -d --build --wait` поднимает всё с пустым томом.
- [ ] Гонка за последними экземплярами: 50 параллельных заказов по одной кружке при остатке 10.

```bash
(for i in $(seq 50); do order 1 & done; wait) | sort | uniq -c
#  10 201
#  40 409
docker compose exec -T postgres psql -U diploma diploma -tAc "SELECT stock FROM products WHERE id = 1"   # 0
```

- [ ] Повтор с тем же ключом не создаёт второй заказ, тот же ключ с другим телом 422, без ключа 400:

```bash
KEY=$(uuidgen)
for i in 1 2; do curl -s -i "${H[@]}" -H "Idempotency-Key: $KEY" \
  -d '{"items":[{"product_id":2,"quantity":1}]}' $API/orders | grep -E '^HTTP|Idempotent'; done
# HTTP/1.1 201 Created
# HTTP/1.1 201 Created
# Idempotent-Replayed: true
curl -s -o /dev/null -w '%{http_code}\n' "${H[@]}" -H "Idempotency-Key: $KEY" \
  -d '{"items":[{"product_id":2,"quantity":2}]}' $API/orders                                  # 422
curl -s -o /dev/null -w '%{http_code}\n' "${H[@]}" -d '{"items":[{"product_id":2,"quantity":1}]}' $API/orders   # 400
```

- [ ] Недопустимые переходы (`ship` неоплаченного, `pay` отменённого) дают 409; чужой заказ 404.
- [ ] Вебхуки: подпишись на `/hook` и `/bad` сервиса `webhook-echo` (в compose запущен с `-fail 3`), создай заказ и подожди 2 минуты (три ретрая с backoff 1 + 16 + 81 секунда). В журнале у `/hook` четыре попытки, последняя 200, получатель в логах подтвердил подпись; у `/bad` одна попытка 400, задача отменена.

```bash
curl -s "${H[@]}" -d '{"url":"http://webhook-echo:9099/hook","secret":"dev-secret-0123456789"}' $API/webhooks
curl -s "${H[@]}" -d '{"url":"http://webhook-echo:9099/bad","secret":"dev-secret-0123456789"}' $API/webhooks
order 2; sleep 120
docker compose exec -T postgres psql -U diploma diploma -c \
  "SELECT w.url, d.attempt, d.status_code FROM webhook_deliveries d JOIN webhooks w ON w.id = d.webhook_id ORDER BY d.id"
docker compose logs webhook-echo | tail -5
```

- [ ] Задачи переживают рестарт: `docker compose stop webhook-echo`, `order 2`, через 5 секунд `docker compose stop api` (задача вебхука ждёт повтора, `SELECT kind, state FROM river_job ORDER BY id DESC LIMIT 5` показывает `retryable`), затем `docker compose start webhook-echo api`: в течение 3 минут в журнале появляется успешная попытка 200.
- [ ] `curl -s localhost:8080/metrics | grep -E 'orders_created_total|webhook_deliveries_total|outbox_lag_seconds'` показывает метрики; при остановленном `webhook-echo` `outbox_lag_seconds` растёт.
- [ ] По `request_id` из заголовка `X-Request-ID` ответа `POST /orders` в логах находятся и строка запроса, и строки воркеров письма и вебхука.
- [ ] Остановка: `docker compose stop api` даёт в логах «остановлен» и `Exited (0)` в `docker compose ps -a api`.
- [ ] README со схемой, решениями, гарантиями, бейджем CI.

### Со звёздочкой

- **SSRF-защита**: резолв хоста вебхука и запрет приватных, loopback и link-local адресов через `net.Dialer.Control`, чтобы DNS не подменил адрес между проверкой и соединением.
- **Отдельный процесс воркеров**: `cmd/worker` с теми же пакетами, API только ставит задачи; в compose два сервиса и несколько реплик воркера.
- **Повторная доставка руками**: `POST /api/v1/webhooks/:id/redeliver?event_id=...` для поддержки.
- **Kafka-вариант** на бумаге (опора: [Kafka и RabbitMQ обзорно](/m09/l04-brokers)): раздел README, как переложить события в Kafka через тот же outbox (воркер публикует, ключ партиции `order_id`) и что изменится в гарантиях.

## Как показать проект на собесе

Интервьюер обычно даёт 3-5 минут на «расскажи о своём проекте» и дальше копает туда, где ты звучал увереннее всего. Поэтому рассказ строится так, чтобы привести его к изюминке, где ты знаешь ответы.

### Рассказ на 3 минуты

1. **Что и зачем (20 секунд).** Одна фраза о задаче и одна о том, что в ней трудного. «Сервис бронирования переговорок. Трудное в нём: два человека бронируют одну комнату одновременно, и ни при какой нагрузке не должно получиться двух броней».
2. **Архитектура (40 секунд).** Схема из README: Gin, сервисный слой, PostgreSQL, Redis, River в том же процессе. Почему монолит и почему этого достаточно.
3. **Главное решение (60 секунд).** Изюминка темы: `EXCLUDE`-ограничение, outbox через `InsertTx` или кеш с инвалидацией после коммита. Назови альтернативу, которую отверг, и почему: «`SELECT FOR UPDATE` по комнате тоже работает, но проверка в базе не зависит от того, не забыл ли кто-то взять блокировку».
4. **Как доказал, что работает (40 секунд).** Конкурентный тест в настоящей базе, `-race` в CI, ручная проверка из чеклиста с цифрами: «20 параллельных запросов, один 201, девятнадцать 409».
5. **Продакшн и что дальше (20 секунд).** Docker на distroless (размер), readiness и graceful shutdown, метрики. Что сделал бы при росте нагрузки в 100 раз.

Держи открытыми README со схемой, файл с главным тестом и терминал с `make up`. Показать живой `curl` и `/metrics` за 30 секунд убедительнее любых слов.

### Чего избегать

- Перечисления технологий без «зачем». «Использовал Redis» ничего не значит, «Redis для кеша правил, потому что вычисление флага идёт на каждый запрос клиента» значит много.
- Фраз «так написано в курсе» и «так принято». Каждое решение объясняй своими словами и компромиссом.
- Скрывать недоделки. «Повторяющиеся брони не успел, вот как бы сделал» звучит лучше, чем уход от вопроса.

::: interview Вопросы по диплому
<details><summary>Почему ограничение в базе, а не проверка в коде перед вставкой?</summary>

Проверка «свободно ли» и вставка в коде это два шага, между которыми другой запрос успевает вставить свою строку: классическая гонка check-then-act. Даже `SERIALIZABLE` вернёт ошибку сериализации, которую придётся ловить и повторять. Ограничение в базе (`UNIQUE`, `EXCLUDE`, `CHECK`) проверяется атомарно при вставке, его невозможно забыть в новом коде, и оно защищает от записи из любого места, включая ручной SQL. Проверка в коде остаётся, но только ради понятного сообщения об ошибке.

</details>
<details><summary>Что такое transactional outbox и почему не отправить письмо прямо в хендлере после коммита?</summary>

Отправка после коммита это dual write: две независимые системы, и процесс может упасть между коммитом и отправкой, тогда заказ есть, а письма нет. Отправка до коммита хуже: письмо ушло, а транзакция откатилась. Outbox записывает намерение «отправить» в ту же базу в той же транзакции, что и бизнес-данные, а доставляет отдельный процесс. С River это `InsertTx`: задача появляется тогда и только тогда, когда закоммитился заказ. Цена: доставка at-least-once, получатель должен уметь дедуплицировать.

</details>
<details><summary>Как сделать POST идемпотентным?</summary>

Клиент генерирует `Idempotency-Key` на операцию и повторяет запрос с тем же ключом. Сервер хранит ключ с уникальным индексом в той же транзакции, что и результат, и на повтор возвращает сохранённый ответ, а не выполняет операцию заново. Тот же ключ с другим телом это ошибка клиента (422), поэтому храним хеш запроса. Область уникальности ключа: пользователь, иначе один клиент может угадать ключ другого.

</details>
<details><summary>Как инвалидируешь кеш и что будет, если Redis упадёт?</summary>

Cache-aside: читаем из Redis, на промахе из базы и кладём в Redis с TTL. При записи сначала коммит в базу, потом удаление ключа: если удалить до коммита, параллельный читатель успеет положить старое значение. Узкая гонка остаётся и так: читатель прочитал базу до коммита, а записал в кеш после удаления. Поэтому TTL короткий, а точное решение это версия в значении и запись «только если новее» атомарным Lua-скриптом. Если Redis упал, сервис ходит в базу напрямую и отвечает медленнее, readiness показывает проблему, но трафик не теряется: кеш это ускорение, а не источник правды.

</details>
<details><summary>Воркер River упал посреди задачи. Что будет?</summary>

Задача останется в состоянии `running`, River через некоторое время посчитает её зависшей и вернёт в очередь, она выполнится повторно. Поэтому воркеры идемпотентны: перечитывают состояние (бронь уже отменена, заказ уже оплачен), используют ключи дедупликации (`event_id`, `batch_key`) и не делают необратимых побочных эффектов без проверки. Это и есть at-least-once на практике.

</details>
<details><summary>Чем liveness отличается от readiness и что происходит при деплое?</summary>

Liveness отвечает «процесс жив», без зависимостей: если она падает, оркестратор перезапускает контейнер. Readiness отвечает «готов принимать трафик» и проверяет базу и Redis: если она падает, под убирают из балансировки без перезапуска. При деплое приходит SIGTERM, сервис сразу переводит readiness в 503, ждёт несколько секунд, пока балансировщики уберут его из ротации, затем `Shutdown` дожидается активных запросов, River дожидается текущих задач, закрывается пул.

</details>
<details><summary>Как масштабировать твой сервис в 10 раз?</summary>

Начать с измерений: метрики покажут, что упирается первым. API без состояния, поэтому реплики за балансировщиком добавляются просто; воркеры River конкурентно разбирают общую очередь в PostgreSQL через `SKIP LOCKED`. Дальше пул соединений (PgBouncer), реплики для чтения, кеш горячих данных. Очередь в базе удобна до тысяч задач в секунду; если событий станет на порядки больше или нужно много независимых потребителей, outbox-воркер начнёт публиковать в Kafka.

</details>
<details><summary>Почему distroless, а не alpine или scratch?</summary>

В distroless/static нет shell и пакетного менеджера, поэтому меньше поверхность атаки и меньше уязвимостей в сканерах, но есть CA-сертификаты, tzdata и пользователь `nonroot`, которых нет в scratch. Go-бинарник с `CGO_ENABLED=0` статический и в этом образе работает как есть. Healthcheck делает сам бинарник флагом `-healthcheck`, потому что curl в образе нет.

</details>
:::

## Что почитать

- `RU` [Диапазонные типы, документация PostgreSQL](https://postgrespro.ru/docs/postgresql/17/rangetypes): `tstzrange`, границы, ограничения-исключения для броней.
- `RU` [btree_gist](https://postgrespro.ru/docs/postgresql/17/btree-gist): зачем расширение для `EXCLUDE` с оператором `=`.
- `RU` [INSERT, документация PostgreSQL](https://postgrespro.ru/docs/postgresql/17/sql-insert): `ON CONFLICT DO NOTHING` и `DO UPDATE` для идемпотентности и upsert статистики.
- `EN` [River: transactional enqueueing](https://riverqueue.com/docs/transactional-enqueueing), [scheduled jobs](https://riverqueue.com/docs/scheduled-jobs), [unique jobs](https://riverqueue.com/docs/unique-jobs), [periodic jobs](https://riverqueue.com/docs/periodic-jobs), [cancelling jobs](https://riverqueue.com/docs/cancelling-jobs): всё про River, что нужно для любой из тем.
- `EN` [Pattern: Transactional outbox](https://microservices.io/patterns/data/transactional-outbox.html): классическое описание паттерна и проблемы dual write.
- `EN` [The Idempotency-Key HTTP Header Field](https://datatracker.ietf.org/doc/draft-ietf-httpapi-idempotency-key-header/): черновик стандарта, какие коды отдавать при повторе и конфликте.
- `EN` [Feature Toggles, Martin Fowler](https://martinfowler.com/articles/feature-toggles.html): виды флагов и как с ними жить.
- `EN` [Validating webhook deliveries, GitHub Docs](https://docs.github.com/en/webhooks/using-webhooks/validating-webhook-deliveries): как устроена подпись вебхуков у GitHub, образец для своей.
- `EN` [Redis INCR](https://redis.io/docs/latest/commands/incr/): паттерн rate limiter на счётчиках.
- `EN` [Redis RENAME](https://redis.io/docs/latest/commands/rename/): атомарное переименование, на нём держится слив статистики в теме 2; обрати внимание, что существующий ключ назначения перезаписывается.
