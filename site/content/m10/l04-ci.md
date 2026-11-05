# CI на GitHub Actions

Тесты, которые запускаются только когда разработчик вспомнил, не защищают ни от чего. **CI** (continuous integration) запускает проверки автоматически на каждый пуш и пул-реквест: форматирование, линтер, тесты с детектором гонок, сборку образа. Пул-реквест с красным CI не мержат, и в `main` всегда лежит код, который хотя бы компилируется и проходит тесты. **CD** (continuous delivery/deployment) продолжает цепочку: собранный образ автоматически едет на staging или в прод.

Для джуна это важно по двум причинам. В первый же рабочий день ты откроешь пул-реквест, и CI либо пропустит его, либо покажет, что сломалось: надо понимать, где смотреть и как воспроизвести локально. А pet-проект с зелёным бейджем CI, линтером и тестами в README выглядит на собесе как работа человека, который знает, как устроена разработка в команде.

В уроке соберём весь путь: Makefile как единая точка входа для команд, конфиг golangci-lint, workflow GitHub Actions с линтером, тестами против настоящего PostgreSQL и сборкой Docker-образа, кеширование, и разберём, почему CI бывает красным, когда у тебя всё зелёное.

## Makefile: одна точка входа

У проекта набирается десяток команд: `go test -race -count=1 ./...`, `golangci-lint run`, `go test -coverprofile=...`, `docker build ...` с флагами. Помнить их наизусть не нужно никому. **Makefile** даёт им короткие имена: `make test`, `make lint`, `make ci`. Новый человек в команде открывает Makefile и видит все операции над проектом. CI вызывает те же цели, что и ты локально, поэтому «у меня проходит, а в CI нет» случается реже.

Make старше Go на 30 лет и устроен так: **цель**, двоеточие, **зависимости**, а на следующих строках **рецепт**, команды через **таб** (не пробелы, это частая ошибка новичков: `missing separator`).

```make
GO        ?= go
PKG       ?= ./...
COVER_MIN ?= 80
# main-пакеты это обвязка без тестов: меряем покрытие только пакетов с логикой
COVER_PKG ?= $(shell $(GO) list $(PKG) | grep -v /cmd/)

.DEFAULT_GOAL := help
.PHONY: help fmt-check vet lint test cover build ci

help: ## показать цели
	@grep -E '^[a-z-]+:.*## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*## "} {printf "  %-10s %s\n", $$1, $$2}'

fmt-check: ## упасть, если есть неотформатированные файлы
	@out=$$(gofmt -l .); if [ -n "$$out" ]; then echo "gofmt не прогнан:"; echo "$$out"; exit 1; fi

vet: ## go vet
	$(GO) vet $(PKG)

lint: ## golangci-lint
	golangci-lint run $(PKG)

test: ## тесты с детектором гонок
	$(GO) test -race -count=1 $(PKG)

cover: ## покрытие и шлагбаум
	$(GO) test -count=1 -coverprofile=coverage.out $(COVER_PKG)
	$(GO) run ./cmd/covergate -min $(COVER_MIN) coverage.out

build: ## бинарник
	CGO_ENABLED=0 $(GO) build -trimpath -o bin/app ./cmd/app

ci: fmt-check vet lint test cover ## всё, что делает CI
```

Что здесь стоит знать:

- `?=` задаёт значение, только если переменная ещё не задана. Поэтому `make cover COVER_MIN=90` переопределяет порог из командной строки.
- `.PHONY` говорит, что цель не файл. Без этого, если в папке случайно появится файл `test`, `make test` решит, что цель уже «собрана», и ничего не сделает.
- `$$` в рецепте превращается в `$` для shell. Одиночный `$` make раскрывает сам.
- Каждая строка рецепта выполняется в **отдельном shell**: `cd dir` на одной строке не действует на следующую. Многострочную логику пиши через `;` или `&&` в одной строке.
- `@` перед командой не печатает её саму перед выполнением.
- `ci: fmt-check vet lint test cover` выполняет зависимости по порядку и останавливается на первой упавшей.
- Цель `help` с комментариями `##` это популярный приём самодокументирования: `make` без аргументов показывает список.

Если в Makefile нет `.DEFAULT_GOAL`, `make` без аргументов выполняет **первую цель в файле**. Поэтому `help` ставят первой или явно делают целью по умолчанию.

Альтернативы Makefile: `Taskfile.yml` (утилита task на Go, YAML вместо синтаксиса make) и `just`. Идея та же, но make уже стоит на любом Linux и macOS, и его знает каждый.

## golangci-lint

`go vet` ловит явные ошибки, но этого мало. **golangci-lint** это агрегатор: десятки линтеров в одном бинарнике, общий кеш, один конфиг. В Go-командах он стоит почти везде.

Конфиг `.golangci.yml` в корне проекта (формат второй версии):

```yaml
version: "2"

run:
  timeout: 5m

linters:
  # standard: errcheck, govet, ineffassign, staticcheck, unused
  default: standard
  enable:
    - bodyclose   # не закрытое тело HTTP-ответа
    - errorlint   # сравнение ошибок через == вместо errors.Is
    - gosec       # типовые проблемы безопасности
    - noctx       # HTTP-запросы и SQL без context
    - revive      # стиль: имена, комментарии
    - gocritic    # подозрительные конструкции
    - unconvert   # лишние преобразования типов
  exclusions:
    presets:
      - std-error-handling
    rules:
      - path: _test\.go
        linters: [gosec, noctx, bodyclose, errcheck]

formatters:
  enable:
    - gofmt
    - goimports
  settings:
    goimports:
      local-prefixes:
        - github.com/stepan/notes
```

Что ловят линтеры из базового набора и почему они полезны:

- **errcheck**: необработанная ошибка, `json.NewEncoder(w).Encode(v)` без проверки. Если ошибку осознанно игнорируешь, пиши `_ =`, тогда это видно на ревью.
- **govet**: то же, что `go vet` (форматные строки `Printf`, копирование мьютекса, `defer` в цикле и многое другое).
- **staticcheck**: большой набор проверок от устаревших API до бессмысленных условий.
- **ineffassign**: присваивание, результат которого никто не читает (часто признак бага).
- **unused**: неиспользуемые функции, типы, поля.

Когда я прогнал этот конфиг по решениям домашки модуля, он нашёл пять мест: необработанную ошибку `Encode`, `os.Exit` после `defer` (отложенные функции не выполнятся), неиспользуемое присваивание в тесте, необработанную ошибку `srv.Close()` и чтение файла по пути из переменной (gosec G304). Четыре были настоящими недочётами, пятое (CLI-утилита читает файл, который передал пользователь) подавлено с объяснением:

```go
f, err := os.Open(path) //nolint:gosec // путь к профилю передаёт сам пользователь CLI
```

`//nolint` без указания линтера и без причины на ревью не пропускают.

Полезные команды:

```bash
golangci-lint run ./...            # прогон
golangci-lint run --fix ./...      # починить то, что умеет (форматирование, импорты)
golangci-lint config verify        # проверить конфиг по схеме
golangci-lint linters              # какие линтеры включены
```

::: trap Линтер новой версии в CI
```yaml
# плохо: завтра выйдет новая версия с новыми проверками,
# и CI покраснеет на коде, который никто не трогал
- uses: golangci/golangci-lint-action@v9
  with:
    version: latest
```

```yaml
# хорошо: версия закреплена, обновление это отдельный осознанный PR
- uses: golangci/golangci-lint-action@v9
  with:
    version: v2.15.0
```

Это общее правило CI: всё, что влияет на результат (версия Go, линтера, базового образа, action), закрепляется. Иначе прогоны перестают быть воспроизводимыми.
:::

## Тесты в CI: -race, -count=1, покрытие

Три флага, которые в CI ставят почти всегда.

**`-race`** включает детектор гонок. Без него тест с гонкой может стабильно проходить годами:

```go
func TestCounter(t *testing.T) {
	n := 0
	done := make(chan bool)
	go func() {
		n++
		done <- true
	}()
	n++
	<-done
	if n == 0 {
		t.Fatal("n == 0")
	}
}
```

<Predict :options="['ok и с -race, и без', 'ok без -race, FAIL с -race', 'FAIL и с -race, и без', 'тест зависнет']" :correct="1">

Без детектора тест проходит: `n` почти всегда получается 2, а проверка и вовсе слабая. С `-race` в выводе `WARNING: DATA RACE` с двумя стеками (запись в горутине и запись в тесте), а тест падает с `race detected during execution of test`. Две горутины пишут `n` без синхронизации, и это гонка независимо от того, какое значение получилось.

</Predict>

::: hood Как работает детектор гонок
`-race` компилирует программу с инструментированием: каждое чтение и запись памяти сопровождается вызовом в рантайм ThreadSanitizer (библиотека из LLVM). Тот для каждой ячейки памяти помнит недавние обращения и для каждой горутины ведёт **векторные часы**: что она «видела» от других горутин. Операции синхронизации (каналы, мьютексы, `WaitGroup`, атомики) связывают часы горутин отношением **happens-before**. Если два обращения к одной ячейке, хотя бы одно из них запись, не упорядочены через happens-before, это гонка, даже если в этом прогоне они случились в «удачном» порядке.

Цена: по документации Go память растёт в 5-10 раз, время в 2-20 раз. Поэтому `-race` включают в тестах и CI, а не в продакшне. И главное ограничение: детектор находит гонки только в коде, который **реально выполнился**. Нет теста на конкурентный путь, нет и гонки в отчёте.
:::

**`-count=1`** отключает кеш тестов. `go test` кеширует успешные результаты пакета и при повторном запуске пишет `(cached)`, не запуская тесты. Кеш учитывает исходники, флаги, а также переменные окружения и файлы, которые тест читал. Но он не знает про внешний мир: базу, сеть, время. Для интеграционных тестов в CI кеш вреден.

```go
func TestEnv(t *testing.T) {
	t.Log(os.Getenv("APP_MODE"))
}
```

```bash
go test .                    # первый запуск
go test .                    # второй
APP_MODE=prod go test .      # третий
```

<Predict :options="['ok, ok (cached), ok (cached)', 'ok, ok (cached), ok', 'ok, ok, ok', 'ok, ok (cached), FAIL']" :correct="1">

Второй запуск ничего не поменял и взят из кеша. В третьем изменилась переменная окружения, которую тест читал через `os.Getenv`, и Go это отслеживает: тест запущен заново. Четвёртый запуск с тем же `APP_MODE=prod` снова был бы `(cached)`. Кеш учитывает переменные и файлы, прочитанные тестом, но не состояние базы или сети, поэтому для интеграционных тестов пишут `-count=1`.

</Predict>

**Покрытие**: `-coverprofile=coverage.out` пишет профиль, `go tool cover -func=coverage.out` печатает покрытие по функциям и итог, `go tool cover -html=coverage.out` открывает раскрашенный код в браузере. Порог покрытия в CI («ниже 80% красный») спорная, но распространённая практика: он не гарантирует качества тестов, зато ловит пул-реквесты, где новый код вообще не покрыт.

Профиль это текстовый файл:

```text
mode: set
goschool/homework/m10/l01-docker/ex1_probe.go:22.2,25.16 4 1
goschool/homework/m10/l01-docker/ex1_probe.go:26.3,27.1 1 0
```

Блок кода (файл, начало и конец), число операторов в нём, сколько раз выполнился. Разобрать его и посчитать покрытие самому это первая задача домашки. Там же подвох: с `-coverpkg` один блок встречается в профиле несколько раз, и наивный подсчёт на реальном профиле даёт 48,5% вместо 97%.

::: trap Пакеты без тестов и порог покрытия
С Go 1.22 `go test -cover ./...` считает пакеты без тестов как 0%, а раньше писал `[no test files]` и пропускал их. На решениях домашки это выглядит так:

```text
ok  	goschool/homework/m10/l04-ci	coverage: 98.4% of statements
	goschool/homework/m10/l04-ci/cmd/covergate		coverage: 0.0% of statements
покрытие 68.2%, порог 80.0%
```

```make
# плохо: main-пакет без тестов роняет общий процент
cover:
	go test -coverprofile=coverage.out ./...
```

```make
# хорошо: меряем пакеты с логикой, обвязку main проверяет интеграционный тест
cover:
	go test -coverprofile=coverage.out $$(go list ./... | grep -v /cmd/)
```

Второй вариант честнее показывает, что покрыто в бизнес-логике. Сам `main` держи тонким: разбор флагов и сборка зависимостей, вся логика в пакетах с тестами.
:::

## GitHub Actions

### Словарь

- **Workflow**: YAML-файл в `.github/workflows/`. Описывает, когда запускаться и что делать.
- **Триггер** (`on`): `push`, `pull_request`, расписание (`schedule` с cron), ручной запуск (`workflow_dispatch`), релиз.
- **Job**: набор шагов на одной виртуальной машине (**runner**). Разные jobs по умолчанию идут **параллельно** на разных машинах; `needs` задаёт порядок.
- **Step**: команда (`run`) или готовое действие (`uses`).
- **Action**: переиспользуемый шаг из маркетплейса, например `actions/checkout` (склонировать репозиторий) или `actions/setup-go` (поставить Go).
- **Service container**: Docker-контейнер, который живёт рядом с job (база, Redis).
- **Secrets**: зашифрованные переменные репозитория, в логах маскируются звёздочками.
- **Artifact**: файл, сохранённый после прогона (отчёт покрытия, бинарник).

::: hood Что происходит при пуше
1. GitHub находит workflow, у которых триггер совпал с событием, и ставит их jobs в очередь.
2. Для каждого job выделяется **свежая виртуальная машина** (`ubuntu-latest`): чистая, без следов прошлых прогонов. Всё, что надо сохранить между jobs, передаётся через artifacts, а между прогонами через cache.
3. Если у job есть `services`, на этой машине запускаются Docker-контейнеры. Job в нашем workflow выполняется прямо на VM, а не в контейнере, поэтому сервис доступен через опубликованный порт на `localhost`. Если job запущен в контейнере (`container:`), к сервису ходят по его имени (`postgres:5432`), как в compose.
4. Шаги выполняются по порядку. Первый упавший шаг (ненулевой код выхода) останавливает job, а `needs` не даст запуститься зависимым jobs.
5. Статус каждого job отображается в пул-реквесте. В настройках репозитория (branch protection, rulesets) нужные проверки помечают обязательными, и кнопка Merge блокируется, пока они не зелёные.
:::

### Workflow целиком

Файл `.github/workflows/ci.yml` для сервиса из [урока про Docker](./l01-docker): `go.mod` и `Dockerfile` в корне, тесты ходят в PostgreSQL по `DATABASE_URL`.

```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:

# Новый пуш в ту же ветку отменяет ещё идущий прогон.
concurrency:
  group: ci-${{ github.ref }}
  cancel-in-progress: true

# Минимальные права токена по умолчанию; кому нужно больше, просит в job.
permissions:
  contents: read

jobs:
  lint:
    runs-on: ubuntu-latest
    timeout-minutes: 10
    steps:
      - uses: actions/checkout@v7
      - uses: actions/setup-go@v7
        with:
          go-version-file: go.mod
      - name: gofmt
        run: |
          out=$(gofmt -l .)
          if [ -n "$out" ]; then echo "gofmt не прогнан:"; echo "$out"; exit 1; fi
      - run: go vet ./...
      - uses: golangci/golangci-lint-action@v9
        with:
          version: v2.15.0

  test:
    runs-on: ubuntu-latest
    timeout-minutes: 15
    services:
      postgres:
        image: postgres:17-alpine
        env:
          POSTGRES_USER: app
          POSTGRES_PASSWORD: app
          POSTGRES_DB: app
        ports:
          - 5432:5432
        options: >-
          --health-cmd "pg_isready -U app -d app"
          --health-interval 2s
          --health-timeout 3s
          --health-retries 15
    env:
      DATABASE_URL: postgres://app:app@localhost:5432/app?sslmode=disable
    steps:
      - uses: actions/checkout@v7
      - uses: actions/setup-go@v7
        with:
          go-version-file: go.mod
      - run: go mod download
      - name: Тесты с гонками и покрытием
        run: go test -race -count=1 -coverprofile=coverage.out ./...
      - name: Итог покрытия
        run: go tool cover -func=coverage.out | tail -1
      - uses: actions/upload-artifact@v7
        with:
          name: coverage
          path: coverage.out

  build:
    needs: [lint, test]
    runs-on: ubuntu-latest
    timeout-minutes: 15
    permissions:
      contents: read
      packages: write
    steps:
      - uses: actions/checkout@v7
      - uses: docker/setup-buildx-action@v4
      - name: Имя образа в нижнем регистре
        run: echo "IMAGE=ghcr.io/${GITHUB_REPOSITORY,,}" >> "$GITHUB_ENV"
      - uses: docker/login-action@v4
        if: github.event_name == 'push'
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      - uses: docker/build-push-action@v7
        with:
          context: .
          push: ${{ github.event_name == 'push' }}
          tags: |
            ${{ env.IMAGE }}:${{ github.sha }}
            ${{ env.IMAGE }}:latest
          cache-from: type=gha
          cache-to: type=gha,mode=max
```

Разбор по частям.

**Триггеры.** `push` только в `main`, иначе каждый пуш в ветку с открытым PR запускал бы два одинаковых прогона (по `push` и по `pull_request`). `concurrency` с `cancel-in-progress` отменяет устаревший прогон, когда ты быстро пушишь исправления подряд: минуты раннеров не тратятся на код, который уже никому не нужен.

**Права.** `GITHUB_TOKEN` выдаётся каждому прогону автоматически. `permissions: contents: read` на уровне файла урезает его до чтения кода, а `packages: write` получает только job сборки, которому нужно запушить образ. Принцип минимальных привилегий: если в зависимость из `uses` проберётся вредоносный код, он получит как можно меньше.

**lint.** `go-version-file: go.mod` берёт версию Go из `go.mod`: одна правда для разработчиков и CI. Проверка `gofmt -l` печатает неотформатированные файлы и падает. `golangci-lint-action` ставит закреплённую версию линтера, кеширует его анализ между прогонами и показывает замечания прямо в диффе пул-реквеста.

**test.** Сервис `postgres` поднимается до шагов, а `--health-cmd` заставляет раннер ждать, пока база станет здоровой: тот же приём, что `service_healthy` в compose. `DATABASE_URL` в `env` уровня job виден всем шагам. Отчёт покрытия сохраняется артефактом, его можно скачать со страницы прогона.

**build.** Запускается только если `lint` и `test` зелёные. На пул-реквесте образ собирается (проверяем, что Dockerfile не сломан), но не пушится; на пуше в `main` уходит в GitHub Container Registry с тегом SHA коммита. Имя образа приводится к нижнему регистру (`${GITHUB_REPOSITORY,,}` это синтаксис bash): registry не принимает заглавные буквы, а в имени владельца репозитория они бывают.

**Таймауты.** `timeout-minutes` у каждого job: по умолчанию job может висеть 6 часов, и зависший тест сожжёт минуты раннеров.

### Кеширование

На чистой машине каждый прогон заново скачивал бы все модули и компилировал всё с нуля. Три уровня кеша:

1. **`actions/setup-go`** с `cache: true` (это значение по умолчанию) сохраняет кеш модулей (`GOMODCACHE`) и кеш сборки (`GOCACHE`). Ключ строится из хеша `go.sum`: поменялись зависимости, кеш собирается заново.
2. **`golangci-lint-action`** кеширует результаты анализа линтера.
3. **Docker-слои** через `cache-from`/`cache-to: type=gha` хранятся в кеше GitHub Actions. Это тот же кеш слоёв из [урока про Docker](./l01-docker): если `go.mod` не менялся, слой с `go mod download` берётся из кеша.

Общий механизм за ними это `actions/cache`: сохранить папку под ключом, восстановить по точному ключу или по префиксу (`restore-keys`). Кеш доступен веткам по правилам: PR видит кеш своей ветки и базовой (`main`), но не соседних PR.

### Матрица

`strategy.matrix` запускает один job в нескольких вариантах параллельно:

```yaml
test:
  strategy:
    matrix:
      go: ['1.26', '1.27']
      os: [ubuntu-latest, macos-latest]
  runs-on: ${{ matrix.os }}
  steps:
    - uses: actions/checkout@v7
    - uses: actions/setup-go@v7
      with:
        go-version: ${{ matrix.go }}
    - run: go test -race ./...
```

Для сервиса, который деплоится одним образом, матрица обычно не нужна. Для библиотеки, которую используют с разными версиями Go и на разных ОС, нужна.

::: trap Секреты в логах и форки
```yaml
# плохо: секрет в командной строке может попасть в лог (set -x, ошибка утилиты),
# а pull_request_target с checkout кода из форка отдаёт секреты чужому коду
on: pull_request_target
steps:
  - uses: actions/checkout@v7
    with:
      ref: ${{ github.event.pull_request.head.sha }}
  - run: ./deploy.sh --token ${{ secrets.DEPLOY_TOKEN }}
```

```yaml
# хорошо: обычный pull_request (у форков секретов нет), секрет через env
on: pull_request
steps:
  - uses: actions/checkout@v7
  - run: ./deploy.sh
    env:
      DEPLOY_TOKEN: ${{ secrets.DEPLOY_TOKEN }}
```

Для `pull_request` из форка GitHub не передаёт секреты, и это защита: код из чужого PR не может их прочитать. `pull_request_target` выполняется в контексте базового репозитория с секретами, и запуск в нём кода из PR это классическая уязвимость CI.
:::

### Когда CI красный, а у тебя зелёный

Типичные причины:

- **Тест зависит от окружения**: часовой пояс (раннеры в UTC), локаль, порядок файлов в директории, наличие файла у тебя в домашней папке.
- **Гонка**: у тебя тест проходит без `-race`, CI запускает с `-race`.
- **Кеш тестов**: у тебя `(cached)` от прошлого зелёного прогона, а тест уже сломан. `-count=1` локально покажет правду.
- **Флаки**: тест зависит от времени (`time.Sleep(100 * time.Millisecond)` и надежда, что горутина успела), а раннер CI медленнее твоего ноутбука. Лечится синхронизацией и `testing/synctest`, а не увеличением sleep.
- **Не закоммичен файл**: сгенерированный код, `go.sum`, тестовые данные.
- **Неотформатированный код или забытый `go mod tidy`**.

Правило отладки: сначала воспроизвести локально **тем же набором команд**, что в CI. Для этого и нужен `make ci`. Утилита `act` умеет запускать workflow GitHub Actions локально в Docker, но для отладки обычно хватает Makefile.

## Что выведет программа?

```make
build:
	@echo build

help:
	@echo help
```

```bash
make
```

<Predict :options="['build', 'help', 'ошибка: не указана цель', 'build и help']" :correct="0">

Без `.DEFAULT_GOAL` make выполняет первую цель в файле. Поэтому в Makefile с самодокументированием цель `help` либо ставят первой, либо пишут `.DEFAULT_GOAL := help`. Иначе `make` без аргументов неожиданно запустит сборку, а то и деплой.

</Predict>

Папка `demo` с подпапкой `sub` и таким Makefile:

```make
show:
	cd sub
	pwd | xargs basename
	cd sub && pwd | xargs basename
```

```bash
make -s show
```

<Predict :options="['sub и sub', 'demo и sub', 'demo и demo', 'ошибка: cd в рецепте запрещён']" :correct="1">

Каждая строка рецепта запускается в отдельном shell. `cd sub` на первой строке поменял каталог только своему shell, который сразу завершился, поэтому вторая строка печатает исходную папку. Третья строка выполняет `cd` и `pwd` в одном shell через `&&`. Флаг `-s` просто не печатает сами команды. Если в рецепте нужен `cd`, его пишут в той же строке, или используют `$(MAKE) -C sub`.

</Predict>

::: interview
<details><summary>Что такое CI и CD? Что обычно проверяют в CI Go-сервиса?</summary>

CI автоматически проверяет каждое изменение: форматирование (`gofmt`), `go vet`, golangci-lint, тесты с `-race` (часто против настоящей базы в service container), покрытие, сборку бинарника и Docker-образа. CD автоматически доставляет собранный артефакт: пушит образ в registry и деплоит на staging или в прод. Пул-реквест с красным CI не мержат, обязательные проверки настраиваются в branch protection.

</details>
<details><summary>Зачем запускать тесты с -race? Почему не в проде?</summary>

Детектор находит гонки данных: обращения к одной памяти из разных горутин без happens-before, хотя бы одно из которых запись. Без него тест с гонкой может стабильно проходить. В проде не включают из-за цены: память в 5-10 раз, время в 2-20 раз. И ограничение: находит гонки только в выполненном коде, поэтому нужны тесты на конкурентные пути.

</details>
<details><summary>Что такое кеш тестов и когда он мешает?</summary>

`go test` кеширует успешный результат пакета по исходникам, флагам, переменным окружения и файлам, которые тест читал, и выводит `(cached)`. Он не знает о внешнем состоянии (база, сеть, время), поэтому для интеграционных тестов и в CI используют `-count=1`. Локально кеш полезен: не перезапускает тесты неизменённых пакетов.

</details>
<details><summary>Что такое golangci-lint и какие линтеры ты бы включил?</summary>

Агрегатор линтеров с общим конфигом и кешем. Базовый набор: errcheck (необработанные ошибки), govet, staticcheck, ineffassign, unused. Полезно добавить errorlint (сравнение ошибок без `errors.Is`), bodyclose, noctx, gosec, gocritic, revive. Версию линтера закрепляют в CI, а подавления `//nolint` пишут с указанием линтера и причины.

</details>
<details><summary>Как ускорить CI?</summary>

Кеш модулей и сборки (`setup-go` делает это по хешу `go.sum`), кеш линтера, кеш Docker-слоёв (`type=gha` или registry). Параллельные jobs вместо одного длинного. `concurrency` с отменой устаревших прогонов. Правильный порядок слоёв в Dockerfile. Не гонять тяжёлые проверки там, где изменений нет (`paths`-фильтры триггера).

</details>
<details><summary>Как в GitHub Actions поднять базу для тестов?</summary>

Через `services`: контейнер `postgres` с `env`, опубликованным портом и `options` с `--health-cmd pg_isready`, чтобы раннер дождался готовности. Если job выполняется на VM, база доступна на `localhost:<порт>`; если job сам в контейнере, то по имени сервиса. Строка подключения передаётся тестам через `env`.

</details>
<details><summary>Зачем Makefile, если есть go test и go build?</summary>

Единая точка входа: короткие имена для длинных команд с флагами, самодокументирование (`make help`), одни и те же команды локально и в CI (`make ci`), зависимости между шагами. Новый человек сразу видит, какие операции есть в проекте. Альтернативы с тем же смыслом: Taskfile, just.

</details>
:::

## Источники

- `RU` [Continuous integration в Go с GitHub Actions](https://habr.com/ru/articles/650435/): перевод статьи Alex Edwards, пошаговый workflow с тестами и статическим анализом. Версии actions там старые, сверяй с уроком.
- `RU` [Пример создания Makefile для Go-приложений](https://habr.com/ru/articles/461467/): цели, переменные, `.PHONY` на примере Go-проекта.
- `EN` [Workflow syntax for GitHub Actions](https://docs.github.com/en/actions/writing-workflows/workflow-syntax-for-github-actions): полный справочник по ключам workflow, `services`, `concurrency`, `permissions`.
- `EN` [actions/setup-go](https://github.com/actions/setup-go): README с `go-version-file` и кешированием.
- `EN` [golangci-lint: Configuration File](https://golangci-lint.run/docs/configuration/file/) и [CI Installation](https://golangci-lint.run/docs/welcome/install/ci/): формат v2 и официальный action.
- `EN` [Data Race Detector](https://go.dev/doc/articles/race_detector): как пользоваться `-race`, цена и ограничения.
- `EN` [Cache management with GitHub Actions](https://docs.docker.com/build/ci/github-actions/cache/): кеш Docker-слоёв `type=gha`.

## Домашка

Всё лежит в `homework/m10/l04-ci/`. Утилита `cmd/covergate` уже написана и заработает, когда решишь первые две задачи: это шлагбаум покрытия для CI.

<Hw id="ex1" title="Покрытие по профилю" level="база">

Файл `homework/m10/l04-ci/ex1_coverage.go`, функция `Coverage(r)`: разобрать профиль `go test -coverprofile` и посчитать процент так же, как `go tool cover`. Повторы одного блока сливаются, битые строки дают ошибку с номером строки. Тест прогоняет настоящий профиль из `testdata/real.out`, где наивный подсчёт даёт 48,5% вместо 97%.

```bash
cd homework && go test -v ./m10/l04-ci/ -run TestCoverage
```

</Hw>

<Hw id="ex2" title="Худшие файлы" level="база">

Файл `homework/m10/l04-ci/ex2_files.go`, функция `WorstFiles(r, threshold)`: файлы с покрытием ниже порога, от худшего к лучшему, при равенстве по имени. Это то, что covergate печатает, когда покрытие ниже нормы.

```bash
cd homework && go test -v ./m10/l04-ci/ -run TestWorstFiles
```

</Hw>

<Hw id="ex3" title="Makefile" level="посложнее">

Напиши `homework/m10/l04-ci/Makefile` по требованиям в комментарии файла: цели `help`, `fmt`, `fmt-check`, `vet`, `lint`, `test`, `cover`, `build`, `ci`, `clean`, переопределяемые переменные.

```bash
cd homework
make -C m10/l04-ci                   # список целей
make -C m10/l04-ci ci                # всё зелёное
make -C m10/l04-ci cover COVER_MIN=99   # шлагбаум срабатывает, печатает худший файл
```

Чеклист: поставь в любом `.go` файле лишний пробел в отступе, `make -C m10/l04-ci fmt-check` должен упасть и назвать файл; `make` без аргументов ничего не собирает, а печатает помощь.

</Hw>

<Hw id="ex4" title="Конфиг golangci-lint" level="посложнее">

Заполни `homework/m10/l04-ci/.golangci.yml` по требованиям в файле. Поставь линтер (`brew install golangci-lint`), проверь конфиг и прогон:

```bash
cd homework/m10/l04-ci
golangci-lint config verify
golangci-lint run ./...            # 0 issues на решённых ex1 и ex2
```

Потом намеренно напиши в `ex1_coverage.go` сравнение `if err == io.EOF` и неиспользуемое присваивание и посмотри, какие линтеры это поймали. Верни код обратно.

</Hw>

<Hw id="ex5" title="Workflow на GitHub" level="со звёздочкой">

Возьми сервис из [урока про Docker](./l01-docker) (или любой свой Go-проект с тестами), заведи для него репозиторий на GitHub и заполни workflow по требованиям из `homework/m10/l04-ci/ci.yml`, взяв за основу пример из урока. Проверь синтаксис до пуша (`go install github.com/rhysd/actionlint/cmd/actionlint@latest`, затем `actionlint ci.yml`). Добейся зелёного прогона, затем по очереди сломай: неотформатированный файл, тест с гонкой, ошибку линтера, синтаксическую ошибку в Dockerfile. Для каждой поломки запиши, какой job упал и где в логе это видно. Включи в настройках репозитория обязательные проверки для `main` и убедись, что PR с красным CI нельзя смержить.

</Hw>
