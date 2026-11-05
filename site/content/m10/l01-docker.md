# Docker для Go

Сервис на ноутбуке и сервис в проде отличаются всем: версией ОС, библиотеками, переменными окружения, правами. Docker упаковывает программу вместе с окружением в **образ**, и этот образ одинаково запускается у тебя, в CI, на staging и в Kubernetes. Сегодня почти любой бэкенд-сервис доставляется в прод контейнером, поэтому «напиши Dockerfile для этого сервиса» вполне реальная задача на испытательном сроке, а «чем multi-stage лучше обычной сборки» и «почему образ весит гигабайт» звучат на собесах.

Для Go Docker особенно приятен. Go собирает **статический бинарник**: один файл, которому не нужны ни интерпретатор, ни рантайм, ни системные библиотеки. Такой файл можно положить в почти пустой образ на 10-20 МБ. У Node.js или Python так не выйдет: им нужен интерпретатор и `node_modules`/`site-packages` рядом.

В уроке соберём образ сервиса тремя способами и сравним размеры, разберём слои и кеш, distroless и scratch, non-root, healthcheck, корректную остановку по SIGTERM и docker compose для сервиса с PostgreSQL. Все цифры ниже получены на Go 1.27.3, Docker 28.5 и macOS на arm64; у тебя они будут немного другими, но порядок тот же.

## Образ, контейнер, слой

Три слова, которые путают чаще всего:

- **Образ** (image): неизменяемый шаблон. Файловая система плюс метаданные: какую команду запускать, от какого пользователя, какие порты открыты.
- **Контейнер**: запущенный экземпляр образа. Это обычный процесс Linux, которого ядро изолировало через **namespaces** (свои PID, сеть, точки монтирования, hostname) и ограничило через **cgroups** (CPU, память). Никакой виртуальной машины внутри нет: ядро общее с хостом.
- **Слой** (layer): одна неизменяемая «дельта» файловой системы. Образ это стопка слоёв.

Аналогия с git точная: слой похож на коммит (набор изменений с хешем), образ на ветку (ссылку на верхний коммит со всей историей под ним), а контейнер на рабочую копию, в которой можно менять файлы, не трогая коммиты.

::: hood Как устроены слои
Каждая инструкция Dockerfile, которая меняет файлы (`RUN`, `COPY`, `ADD`), создаёт новый слой. Инструкции вроде `ENV`, `USER`, `EXPOSE`, `ENTRYPOINT` меняют только метаданные образа.

При запуске контейнера слои образа складываются друг на друга **union-файловой системой** (обычно overlayfs): нижние слои только для чтения, сверху добавляется тонкий слой для записи. Если контейнер меняет файл из нижнего слоя, файл сначала копируется наверх (copy-on-write). Поэтому десять контейнеров одного образа не занимают в десять раз больше места: слои образа у них общие.

```text
 контейнер:  [ слой записи контейнера ]   ← исчезнет вместе с контейнером
 образ:      [ COPY /out/app /app      ]   8,9 МБ
             [ слои distroless          ]   <1 МБ в сжатом виде
```

Отсюда важное следствие: **удалить файл в следующем слое не значит уменьшить образ**. `RUN rm -rf /src` после `COPY . /src` только помечает файлы удалёнными сверху, байты остаются в нижнем слое. Уменьшает образ только то, что в него не попало.

Посмотреть слои: `docker history <образ>`.
:::

## Наивный Dockerfile и почему он плохой

Возьмём небольшой сервис: HTTP-сервер с `/healthz` и `/readyz`, ходит в PostgreSQL через pgx, логирует через `slog`. Структура:

```text
notes/
├── go.mod
├── go.sum
└── cmd/app/main.go
```

Первая попытка, которую пишет каждый:

```dockerfile
FROM golang:1.27
WORKDIR /src
COPY . .
RUN go build -o /app ./cmd/app
CMD ["/app"]
```

Работает. Но `docker images` показывает:

```text
naive        1.58GB
```

Внутри компилятор Go, стандартная библиотека в исходниках, кеш модулей, gcc, git, весь Debian. Для запуска нужен только бинарник на 13 МБ. Остальное:

- **размер**: каждый деплой тянет по сети лишний гигабайт, ноды кластера забиваются;
- **безопасность**: сканер уязвимостей (Trivy, Grype) найдёт сотни CVE в пакетах Debian, которые программа не использует; атакующий, получивший выполнение кода, найдёт в контейнере shell, компилятор и пакетный менеджер;
- **кеш**: `COPY . .` первой строкой означает, что любая правка кода инвалидирует всё, включая скачивание зависимостей.

## Multi-stage сборка

Идея: собирать в одном образе, а запускать в другом. В Dockerfile несколько `FROM`, и из предыдущей стадии в следующую копируется только нужный файл.

```dockerfile
# syntax=docker/dockerfile:1

# ---- стадия сборки ----
FROM golang:1.27-alpine AS build
WORKDIR /src

# Сначала только go.mod и go.sum: слой с зависимостями
# пересоберётся, только когда они изменятся.
COPY go.mod go.sum ./
RUN go mod download

# Теперь исходники. Правка кода инвалидирует кеш начиная отсюда.
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/app ./cmd/app

# ---- финальный образ ----
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/app /app
USER nonroot:nonroot
EXPOSE 8080
HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
  CMD ["/app", "-healthcheck"]
ENTRYPOINT ["/app"]
```

Результат на том же сервисе:

```text
naive        1.58GB
distroless   18.6MB
scratch      12.7MB
```

Стадия `build` в итоговый образ не попадает вообще: Docker хранит её только в кеше сборки. В финальном образе лежат слои distroless и один слой с нашим бинарником.

Разберём каждую строку.

### Порядок инструкций и кеш

Docker кеширует каждый шаг. Шаг берётся из кеша, если не изменились ни сама инструкция, ни все шаги до неё, ни (для `COPY`) содержимое копируемых файлов. Как только один шаг пересобран, все шаги ниже пересобираются тоже.

Поэтому редко меняющееся кладём наверх, часто меняющееся вниз. `go.mod` и `go.sum` меняются раз в неделю, код каждые пять минут. Копируем их отдельно и скачиваем модули: этот слой переживает любые правки кода. Вот что показывает `docker build --progress=plain` после правки `main.go`:

```text
#11 [build 3/6] COPY go.mod go.sum ./
#11 CACHED
#12 [build 4/6] RUN go mod download
#12 CACHED
#13 [build 5/6] COPY . .
#13 DONE 0.2s
#14 [build 6/6] RUN CGO_ENABLED=0 go build ...
#14 DONE 22.9s
```

Зависимости не перекачивались, но компиляция заняла 23 секунды: кеш компилятора Go (`~/.cache/go-build`) живёт внутри шага `RUN` и выбрасывается вместе с ним. BuildKit умеет монтировать в шаг постоянный кеш:

```dockerfile
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/app ./cmd/app
```

У меня пересборка после правки упала с 31 до 5,6 секунды. Содержимое cache mount не попадает ни в какой слой образа, это просто папка, которую BuildKit хранит между сборками. Так же можно монтировать `/go/pkg/mod` для модулей.

### .dockerignore

`COPY . .` копирует весь **контекст сборки**: папку, которую ты передал `docker build`. Без `.dockerignore` туда уедут `.git`, локальные `.env` с паролями, собранные бинарники, и любое их изменение сломает кеш шага `COPY`.

```text
.git
.env
bin/
*.md
Dockerfile*
compose.yaml
```

BuildKit сначала ищет рядом с Dockerfile файл `<имя Dockerfile>.dockerignore` (например `Dockerfile.dockerignore`), и если он есть, то главнее общего `.dockerignore` в корне контекста. Это удобно, когда в одном контексте несколько Dockerfile с разными нуждами. Пригодится в домашке.

### CGO_ENABLED=0 и статический бинарник

Go умеет вызывать C-код через **cgo**. Если cgo включён, некоторые пакеты стандартной библиотеки (`net` для DNS, `os/user`) используют системную libc, и бинарник становится **динамически слинкованным**: при старте ему нужны `libc.so.6` и динамический загрузчик `ld-linux`.

Когда cgo включён? Если в окружении есть C-компилятор. В образе `golang:1.27` (Debian) gcc есть, и `go env CGO_ENABLED` даёт `1`. В `golang:1.27-alpine` gcc нет, и там уже `0`. Полагаться на это не стоит: пиши `CGO_ENABLED=0` явно, тогда сборка не зависит от базового образа.

`ldd` бинарника, собранного в `golang:1.27` без флага:

```text
linux-vdso.so.1
libc.so.6 => /lib/aarch64-linux-gnu/libc.so.6
/lib/ld-linux-aarch64.so.1
```

С `CGO_ENABLED=0` Go использует свои реализации на чистом Go (DNS-резолвер, разбор `/etc/passwd`), и бинарнику не нужно ничего, кроме ядра Linux.

Остальные флаги:

- `-trimpath` убирает из бинарника абсолютные пути твоей машины (`/Users/stepan/...`): сборка воспроизводима и не светит структуру диска.
- `-ldflags="-s -w"` выбрасывает таблицу символов и отладочную информацию DWARF. Бинарник становится на 30% меньше (13,0 МБ против 8,9 МБ у меня), но `dlv` к нему уже не подключить. Стек-трейсы паник при этом остаются читаемыми: имена функций рантайм берёт из своей таблицы `pclntab`, а не из DWARF.

### distroless и scratch

**scratch** это пустой образ: ноль файлов. Подходит для статического бинарника, но всё, что обычно даёт ОС, придётся принести самому:

- **CA-сертификаты** для HTTPS-запросов наружу (`/etc/ssl/certs/ca-certificates.crt`), иначе любой запрос к внешнему API упадёт с ошибкой `x509: certificate signed by unknown authority`;
- **часовые пояса**: `time.LoadLocation("Europe/Moscow")` ищет базу в `/usr/share/zoneinfo`, которой нет;
- **пользователя**: в scratch нет `/etc/passwd`, поэтому `USER` пишут числом (`USER 65532:65532`);
- **`/tmp`**: если программа пишет временные файлы, его тоже нет.

```dockerfile
FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /out/app /app
USER 65532:65532
ENTRYPOINT ["/app"]
```

Часовые пояса проще вшить в бинарник: пустой импорт `_ "time/tzdata"` добавляет около 400 КБ и делает `LoadLocation` независимым от ОС.

**distroless** от Google это «почти scratch»: никакого shell и пакетного менеджера, но уже есть CA-сертификаты, tzdata, `/etc/passwd` с пользователем `nonroot` (uid 65532) и `/tmp`. Вариант `static-debian12` занимает меньше мегабайта в сжатом виде (`docker image inspect` показал у меня 722 КБ) и рассчитан как раз на статические бинарники. Тег `:nonroot` сразу выставляет непривилегированного пользователя. Есть тег `:debug` с busybox-шеллом, если очень нужно залезть внутрь.

| База | Размер у меня | Shell | CA, tzdata | Когда брать |
|---|---|---|---|---|
| `golang:1.27` | 1,58 ГБ | да | да | никогда для запуска, только для сборки |
| `alpine` | +8 МБ к бинарнику | да (busybox) | ставятся `apk add` | нужен shell для отладки или cgo с musl |
| `distroless/static` | 18,6 МБ | нет | да | разумный дефолт для Go |
| `scratch` | 12,7 МБ | нет | нет, копируешь сам | максимальный минимализм |

Отсутствие shell это плюс для безопасности и минус для отладки: `docker exec app sh` не сработает (`exec: "sh": executable file not found`). Для отладки есть `docker debug` и эфемерные контейнеры в Kubernetes (`kubectl debug`), которые подключают инструменты снаружи, не раздувая образ.

```go
package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	_, err := net.LookupHost("localhost")
	fmt.Println("lookup:", err)
	_, err = time.LoadLocation("Europe/Moscow")
	fmt.Println("tz:", err)
}
```

Программа собрана с `CGO_ENABLED=0` и запущена в образе `FROM scratch`, куда скопирован только бинарник.

<Predict :options="['lookup: <nil> и tz: <nil>', 'lookup: <nil> и tz: unknown time zone Europe/Moscow', 'обе строки с ошибкой', 'exec /app: no such file or directory']" :correct="1">

Пустой образ scratch не значит пустой `/etc` в контейнере: Docker при запуске монтирует в любой контейнер `/etc/hosts`, `/etc/resolv.conf` и `/etc/hostname`. Чистый Go-резолвер читает `/etc/hosts` и находит там `localhost`, поэтому первая строка без ошибки. Если подменить `/etc/hosts` пустым файлом и отключить сеть, lookup пойдёт в DNS из `resolv.conf` и упадёт: `dial udp ...: connect: network is unreachable`. А базу часовых поясов Docker не приносит, и `LoadLocation` честно говорит `unknown time zone Europe/Moscow`. Лечится `import _ "time/tzdata"` или переходом на distroless.

</Predict>

### Не root

По умолчанию процесс в контейнере работает от root (uid 0). Это тот же uid 0, что и на хосте: изоляция namespaces снижает риск, но при уязвимости в рантайме контейнеров или неудачном монтировании root внутри превращается в root снаружи. Плюс Kubernetes с политикой `runAsNonRoot: true` просто откажется запускать такой под.

В distroless `:nonroot` пользователь уже есть, в alpine создают сами:

```dockerfile
FROM alpine:3.22
RUN adduser -D -u 10001 app
USER app
```

Порты ниже 1024 непривилегированный процесс открыть не может, поэтому Go-сервисы слушают 8080, а 80/443 публикуются снаружи (`-p 80:8080`) или на балансировщике.

## Сигналы и PID 1

`docker stop` отправляет главному процессу контейнера **SIGTERM**, ждёт 10 секунд (настраивается `docker stop -t` или `stop_grace_period` в compose) и добивает **SIGKILL**. Kubernetes делает то же самое с `terminationGracePeriodSeconds` (по умолчанию 30 секунд). За это время сервис должен перестать принимать новые запросы и дообслужить текущие: это **graceful shutdown**. Ты уже делал его в [уроке «Сервер к продакшну»](/m06/l05-production-server), а в [уроке про наблюдаемость](./l02-observability) добавим паузу drain перед остановкой.

```go
ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
defer stop()
// ... srv.ListenAndServe в горутине ...
<-ctx.Done()
sctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
srv.Shutdown(sctx) // перестаёт принимать, ждёт активные запросы
```

Чтобы SIGTERM вообще дошёл до программы, она должна быть **PID 1** в контейнере. Это зависит от формы `CMD`/`ENTRYPOINT`. Shell-форма запускает `/bin/sh -c "..."`. Одиночную простую команду busybox (alpine), bash и dash сами заменяют программой через exec, и тогда PID 1 всё-таки она. Но стоит появиться `&&`, `;` или пайпу, PID 1 остаётся shell, и сигнал до программы не доходит:

```dockerfile
ENTRYPOINT ["/app"]           # exec-форма: /app это PID 1, сигнал доходит
ENTRYPOINT /migrate && /app   # shell-форма: PID 1 это /bin/sh -c "/migrate && /app"
```

::: hood Почему PID 1 особенный
В Linux процесс с PID 1 в своём PID-namespace это init. Ядро не применяет к нему действия сигналов **по умолчанию**: если у init нет своего обработчика SIGTERM, сигнал просто игнорируется. У Go-программы с `signal.NotifyContext` обработчик есть, поэтому exec-форма работает.

В shell-форме PID 1 это `sh`. Сигнал получает он и не пересылает его дочернему процессу. Вот `ps` внутри контейнера с `CMD /app && echo done` на alpine:

```text
PID   USER     TIME  COMMAND
    1 root      0:00 /bin/sh -c /app && echo done
    7 root      0:00 /app
```

Проверка оптимизации на alpine: `sh -c 'ps'` показывает PID 1 у самого `ps`, а `sh -c 'ps && echo done'` показывает PID 1 у `/bin/sh -c ps && echo done`. Полагаться на такое поведение шелла нельзя, пиши exec-форму. Ещё два минуса shell-формы: в distroless и scratch она вообще не запустится (нет `/bin/sh`), а при shell-форме `ENTRYPOINT` аргументы из `CMD` и `docker run` игнорируются.

Если программе всё же нужен shell-скрипт перед стартом, в конце скрипта пиши `exec /app "$@"`: `exec` заменяет процесс shell программой, и она становится PID 1. Если программа сама не умеет обрабатывать сигналы и собирать зомби-процессы, ставят крошечный init: `docker run --init` или `tini`.
:::

### HEALTHCHECK без curl

`HEALTHCHECK` говорит Docker, как проверить, что контейнер жив. Статус виден в `docker ps` (`Up 5 seconds (healthy)`), compose может ждать `service_healthy`, Docker Swarm перезапускает нездоровые контейнеры.

В distroless нет ни curl, ни wget, поэтому проверку делает сам бинарник: флаг `-healthcheck` превращает сервер в маленький HTTP-клиент, который стучится в свой же `/healthz` и выходит с кодом 0 или 1.

```go
health := flag.Bool("healthcheck", false, "проверить /healthz и выйти")
flag.Parse()
if *health {
	os.Exit(probe("http://127.0.0.1:8080/healthz"))
}
```

Kubernetes инструкцию `HEALTHCHECK` игнорирует: там свои `livenessProbe` и `readinessProbe`, которые kubelet выполняет снаружи контейнера (HTTP-запрос на порт пода). Поэтому HTTP-эндпоинты `/healthz` и `/readyz` нужны в любом случае, а флаг `-healthcheck` полезен для Docker и compose.

### Go в контейнере: CPU и память

С Go 1.25 рантайм на Linux читает лимит CPU из cgroup и уменьшает `GOMAXPROCS` под него: лимит округляется вверх, но не ниже 2 (`limit = max(ceil(limit), 2)` в `runtime/cgroup_linux.go`), и не выше числа ядер. Проверка на машине с 10 ядрами:

```text
docker run --rm app                  → NumCPU: 10 GOMAXPROCS: 10
docker run --rm --cpus=0.5 app       → NumCPU: 10 GOMAXPROCS: 2
docker run --rm --cpus=1 app         → NumCPU: 10 GOMAXPROCS: 2
docker run --rm --cpus=1.5 app       → NumCPU: 10 GOMAXPROCS: 2
docker run --rm --cpus=3 app         → NumCPU: 10 GOMAXPROCS: 3
```

Обрати внимание на `--cpus=1`: частый `limits.cpu: 1` в Kubernetes даёт GOMAXPROCS 2, то есть больше лимита. Нижняя граница 2 выбрана, чтобы не терять параллелизм (GC, сетевые горутины) на маленьких лимитах. До 1.25 Go видел все ядра хоста, создавал лишние P и упирался в троттлинг CFS, поэтому в старых проектах встретишь библиотеку `go.uber.org/automaxprocs`.

С памятью автоматики нет: `--memory=256m` рантайм не учитывает, и `debug.SetMemoryLimit(-1)` в таком контейнере вернёт `9223372036854775807`, то есть «без лимита». Сборщик мусора не знает, что рядом стенка, и контейнер может получить OOM kill. Выставляй переменную `GOMEMLIMIT` чуть ниже лимита контейнера (например `GOMEMLIMIT=230MiB` при лимите 256 МБ): GC начнёт работать агрессивнее, приближаясь к ней.

## docker compose: сервис и PostgreSQL

Compose описывает несколько контейнеров, сеть между ними и тома в одном YAML. Для локальной разработки и учебных стендов это стандарт (ты уже пользуешься им: `homework/docker-compose.yml` поднимает базу курса, см. [урок про PostgreSQL в Docker](/m07/l01-postgres-intro)).

```yaml
services:
  app:
    build: .
    ports: ["18080:8080"]
    environment:
      DATABASE_URL: postgres://notes:notes@db:5432/notes?sslmode=disable
    depends_on:
      db:
        condition: service_healthy
    restart: unless-stopped

  db:
    image: postgres:17-alpine
    environment:
      POSTGRES_USER: notes
      POSTGRES_PASSWORD: notes
      POSTGRES_DB: notes
    volumes: ["dbdata:/var/lib/postgresql/data"]
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U notes -d notes"]
      interval: 2s
      timeout: 3s
      retries: 15

volumes:
  dbdata:
```

Что здесь важно:

- **Сеть и DNS.** Compose создаёт сеть проекта, и контейнеры видят друг друга по имени сервиса. Внутри `app` адрес базы `db:5432`, а не `localhost`: `localhost` внутри контейнера это сам контейнер.
- **`ports`** публикует порт на хост (`хост:контейнер`). Базе порт наружу не нужен, `app` достучится до неё по внутренней сети.
- **`depends_on` с `condition: service_healthy`** ждёт, пока healthcheck базы станет зелёным. Без условия compose ждёт только *старта* контейнера, а PostgreSQL в этот момент ещё инициализируется, и первый запрос сервиса упадёт.
- **Именованный том** `dbdata` переживает `docker compose down`. Удаляет его только `down -v`.
- **Конфиг через `environment`**, а не файлом внутри образа. Один и тот же образ едет на staging и в прод с разными переменными. Это принцип 12-factor app: конфиг живёт в окружении (подробнее в уроках [«Сервер к продакшну»](/m06/l05-production-server) и [«Конфиг, логи, OpenAPI и версионирование»](/m08/l06-config-logs-docs)).

Запуск и проверка:

```bash
docker compose up -d --build
docker compose ps                  # app ... Up 5 seconds (healthy)
curl localhost:18080/readyz        # ready
docker compose logs -f app
docker compose down -v             # снести всё вместе с томом
```

::: trap localhost внутри контейнера
```yaml
# плохо: внутри контейнера localhost это сам контейнер, базы там нет
DATABASE_URL: postgres://notes:notes@localhost:5432/notes
```

```yaml
# хорошо: имя сервиса из compose резолвится во внутренней сети
DATABASE_URL: postgres://notes:notes@db:5432/notes
```

С хоста (из `go test` на ноутбуке) к той же базе ходят через опубликованный порт: `localhost:55432` в базе курса. Два разных адреса для одной базы это нормально: изнутри сети compose и снаружи.
:::

::: trap Секреты в образе
```dockerfile
# плохо: пароль навсегда в слое образа, его видит любой, кто скачал образ
ENV DATABASE_URL=postgres://prod:SuperSecret@db/prod
COPY .env /app/.env
```

```dockerfile
# хорошо: в образе нет конфигурации, она приходит при запуске
ENTRYPOINT ["/app"]
```

```bash
docker run -e DATABASE_URL="$DATABASE_URL" app
```

`docker history` и `docker inspect` показывают все `ENV` и команды сборки. Даже `RUN rm .env` в следующем слое не поможет: файл остался в слое ниже. В проде секреты приходят из Kubernetes Secrets, Vault или переменных CI, а для секретов, нужных во время сборки (токен приватного Git), есть `RUN --mount=type=secret`.
:::

::: trap Пакеты ОС, которые не нужны
```dockerfile
# плохо: shell, curl и apk в рантайм-образе ради healthcheck
FROM alpine:3.22
RUN apk add --no-cache curl
HEALTHCHECK CMD curl -f http://localhost:8080/healthz || exit 1
```

```dockerfile
# хорошо: проверку делает сам бинарник
FROM gcr.io/distroless/static-debian12:nonroot
HEALTHCHECK CMD ["/app", "-healthcheck"]
```

Каждый лишний пакет это лишние CVE в отчёте сканера и лишний инструмент для атакующего.
:::

## Что выведет программа?

Бинарник собран так (обрати внимание на базовый образ стадии сборки и отсутствие `CGO_ENABLED`), программа импортирует `net` и делает DNS-запрос:

```dockerfile
FROM golang:1.27 AS build
WORKDIR /src
COPY . .
RUN go build -o /app .

FROM scratch
COPY --from=build /app /app
ENTRYPOINT ["/app"]
```

```bash
docker run --rm app
```

<Predict :options="['программа отработает нормально', 'exec /app: no such file or directory', 'permission denied', 'panic: cgo not available']" :correct="1">

В `golang:1.27` есть gcc, поэтому cgo включён, а пакет `net` слинковался с libc. Бинарник динамический: при запуске ядро ищет интерпретатор `/lib/ld-linux-aarch64.so.1`, указанный в ELF-заголовке, и не находит его в пустом scratch. Ошибка сбивает с толку: «нет такого файла», хотя `/app` на месте. Не хватает загрузчика, а не программы. Лечится `CGO_ENABLED=0`.

</Predict>

Dockerfile из раздела про multi-stage, образ уже собирался. Ты поправил комментарий в `cmd/app/main.go` и запустил `docker build` снова.

```dockerfile
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/app ./cmd/app
```

<Predict :options="['всё возьмётся из кеша: код не менялся по смыслу', 'пересоберутся только COPY . . и go build', 'пересоберутся go mod download и всё ниже', 'пересоберётся всё, начиная с FROM']" :correct="1">

Кеш `COPY` проверяет содержимое файлов по контрольным суммам, смысл кода его не интересует: изменился байт, шаг пересобирается, а за ним и все нижние. `go.mod` и `go.sum` не менялись, поэтому `go mod download` остаётся `CACHED`. Это и есть смысл разделения на два `COPY`.

</Predict>

Сервис корректно ловит SIGTERM и пишет в лог `stopped`. Его запустили в образе с `CMD /app && echo done` и сделали `docker stop`.

<Predict :options="['в логе stopped, код выхода 0', 'в логе stopped и done', 'в логе нет stopped, код выхода 137', 'контейнер не остановится никогда']" :correct="2">

Shell-форма запускает `/bin/sh -c "..."`, а из-за `&&` shell не может заменить себя программой через exec, и PID 1 остаётся он. SIGTERM уходит ему, он не пересылает сигнал программе, и Docker после таймаута отправляет SIGKILL. Код 137 это 128 + 9 (номер SIGKILL). Graceful shutdown не случился, активные запросы оборвались. Exec-форма `CMD ["/app"]` в том же эксперименте дала `stopped` и код 0.

</Predict>

::: interview
<details><summary>Чем образ отличается от контейнера? Что такое слой?</summary>

Образ это неизменяемый шаблон: стопка слоёв файловой системы плюс метаданные (команда, пользователь, порты). Контейнер это запущенный процесс, изолированный namespaces и ограниченный cgroups, с файловой системой из слоёв образа и тонким слоем записи сверху (overlayfs, copy-on-write). Слой создаёт каждая инструкция, меняющая файлы (`RUN`, `COPY`, `ADD`), слои кешируются и переиспользуются между образами.

</details>
<details><summary>Что такое multi-stage сборка и зачем она Go-сервису?</summary>

Несколько `FROM` в одном Dockerfile: в стадии сборки есть компилятор и все зависимости, в финальный образ через `COPY --from=build` попадает только бинарник. Образ уменьшается с 1-1,5 ГБ до 10-20 МБ, в нём нет компилятора, shell и сотен пакетов с CVE. Go здесь особенно выигрывает, потому что статическому бинарнику не нужна даже libc.

</details>
<details><summary>Зачем CGO_ENABLED=0?</summary>

С включённым cgo пакеты `net` и `os/user` используют libc, и бинарник становится динамически слинкованным: ему нужен загрузчик `ld-linux` и `libc.so`. В scratch или distroless/static их нет, и запуск падает с `exec /app: no such file or directory`. С `CGO_ENABLED=0` Go берёт реализации на чистом Go, бинарник полностью статический. Флаг пишут явно, потому что значение по умолчанию зависит от наличия C-компилятора в образе сборки.

</details>
<details><summary>scratch или distroless? Что нужно добавить в scratch?</summary>

scratch пустой: нужно принести CA-сертификаты для исходящего HTTPS, базу часовых поясов (или `import _ "time/tzdata"`), задать пользователя числом, при необходимости создать `/tmp`. distroless/static это всё уже содержит, весит меньше мегабайта и даёт пользователя `nonroot`. Для Go distroless/static разумный дефолт, scratch выбирают ради последних мегабайт или из принципа.

</details>
<details><summary>Как ускорить пересборку образа?</summary>

Порядок слоёв: сначала `COPY go.mod go.sum` и `go mod download`, потом `COPY . .` и сборка. Тогда правка кода не перекачивает модули. `.dockerignore`, чтобы `.git` и мусор не ломали кеш `COPY`. Cache mounts BuildKit (`RUN --mount=type=cache,target=/root/.cache/go-build`) сохраняют кеш компилятора между сборками. В CI дополнительно кеш слоёв в registry или в кеше GitHub Actions.

</details>
<details><summary>Почему важна exec-форма ENTRYPOINT?</summary>

В exec-форме (`["/app"]`) программа становится PID 1 и получает SIGTERM от `docker stop` или Kubernetes. В shell-форме программу запускает `sh -c`. Одиночную команду шеллы часто заменяют через exec, но с `&&`, `;` или пайпом PID 1 остаётся `sh`, он не пересылает сигнал, программа не делает graceful shutdown, и через таймаут её убивает SIGKILL (код 137). Полагаться на оптимизацию шелла нельзя, а в distroless shell-форма вообще не запустится: нет `/bin/sh`. Если нужен скрипт-обёртка, он заканчивается `exec /app`.

</details>
<details><summary>Как Go-программа узнаёт о лимитах контейнера?</summary>

С Go 1.25 рантайм читает лимит CPU из cgroup и уменьшает `GOMAXPROCS` под него: округляет вверх, но не ниже 2, поэтому `--cpus=1` даёт 2. При изменении лимита значение обновляется. Раньше для этого подключали `automaxprocs`. Лимит памяти рантайм сам не читает: нужно задать `GOMEMLIMIT` немного ниже лимита контейнера, иначе GC не знает о стенке и процесс может получить OOM kill.

</details>
<details><summary>Как в compose дождаться готовности базы?</summary>

У сервиса базы описать `healthcheck` (`pg_isready`), а у приложения `depends_on: db: condition: service_healthy`. Простой `depends_on` ждёт только старта контейнера, не готовности. И всё равно приложение должно уметь переживать недоступность базы (ретраи при старте, readiness 503), потому что в проде база может отвалиться в любой момент, а Kubernetes порядок запуска не гарантирует.

</details>
:::

## Источники

- `RU` [Как я уменьшил Docker-образ Go-приложения с 1.92 GB до 9 MB](https://habr.com/ru/articles/962318/): та же дорога, что в уроке, шаг за шагом с цифрами; про UPX читай критично, в проде его почти не используют.
- `EN` [Multi-stage builds](https://docs.docker.com/build/building/multi-stage/): официальная документация, разделы про именованные стадии и `--target`.
- `EN` [Optimize cache usage in builds](https://docs.docker.com/build/cache/optimize/): порядок слоёв, `.dockerignore`, cache mounts.
- `EN` [Dockerfile reference](https://docs.docker.com/reference/dockerfile/): точная семантика `ENTRYPOINT`/`CMD` (exec и shell форма), `HEALTHCHECK`, `USER`.
- `EN` [GoogleContainerTools/distroless](https://github.com/GoogleContainerTools/distroless): что внутри образов, теги `nonroot` и `debug`.
- `EN` [Control startup and shutdown order in Compose](https://docs.docker.com/compose/how-tos/startup-order/): `depends_on` с условиями и healthcheck.

## Домашка

Сервис для упражнений уже лежит в `homework/m10/l01-docker/cmd/server/main.go`: конфиг из окружения, `/healthz`, `/readyz` с пингом базы, graceful shutdown и флаг `-healthcheck`. Он заработает, когда ты решишь первые две задачи.

<Hw id="ex1" title="Probe для HEALTHCHECK" level="база">

Файл `homework/m10/l01-docker/ex1_probe.go`, функция `Probe(ctx, url, timeout)`: GET с таймаутом, `nil` на любой 2xx, ошибка с кодом ответа в тексте на остальные. Тесты проверяют, что таймаут работает на зависшем сервере, отменённый `ctx` уважается, а пять проверок подряд идут по одному keep-alive соединению (закрывай тело ответа, передавай `ctx` в запрос через `http.NewRequestWithContext` и не создавай свой `Transport` на каждый вызов). Тест узнаёт о возврате соединения в пул через `httptrace`, а не по таймингам.

```bash
cd homework && go test -v ./m10/l01-docker/ -run TestProbe
```

</Hw>

<Hw id="ex2" title="Конфиг из окружения" level="база">

Файл `homework/m10/l01-docker/ex2_config.go`, функция `LoadConfig(getenv)`. Значения по умолчанию, проверка диапазонов, и главное: все ошибки разом через `errors.Join`, чтобы тот, кто запускает контейнер, увидел все проблемы с первого раза.

```bash
cd homework && go test -v ./m10/l01-docker/ -run TestLoadConfig
```

</Hw>

<Hw id="ex3" title="Dockerfile сервиса" level="посложнее">

Заполни `homework/m10/l01-docker/Dockerfile` по требованиям из комментария в файле: multi-stage, отдельный слой зависимостей, `CGO_ENABLED=0`, distroless nonroot, HEALTHCHECK через `-healthcheck`, exec-форма. Контекст сборки это вся папка `homework`, лишнее отсекает уже готовый `Dockerfile.dockerignore` рядом.

```bash
cd homework
docker build -f m10/l01-docker/Dockerfile -t goschool-l01:dev .
docker images goschool-l01          # меньше 30 МБ
docker run --rm -e PORT=abc goschool-l01:dev; echo $?   # ошибки конфига и код 2
```

Чеклист: поправь комментарий в `main.go` и пересобери с `--progress=plain`, шаг `go mod download` должен быть `CACHED`; `docker history goschool-l01:dev` показывает один слой с бинарником поверх distroless; `docker inspect goschool-l01:dev | grep '"User"'` показывает `nonroot:nonroot`, а не пустую строку.

</Hw>

<Hw id="ex4" title="compose: сервис и PostgreSQL" level="посложнее">

Заполни `homework/m10/l01-docker/compose.yaml` по требованиям в комментарии: сервис собирается твоим Dockerfile, ждёт healthcheck базы, данные в именованном томе.

```bash
cd homework/m10/l01-docker
docker compose up -d --build
docker compose ps                     # app (healthy), db (healthy)
curl localhost:18080/readyz           # ready
docker compose stop db
curl -i localhost:18080/readyz        # 503: сервис жив, но не готов
curl localhost:18080/healthz          # ok
docker compose down -v
```

Чеклист: `docker compose stop app` завершается быстро, а в `docker compose logs app` есть строка «остановлен» (graceful shutdown сработал); после `down` без `-v` и нового `up` данные базы на месте.

</Hw>

<Hw id="ex5" title="scratch своими руками" level="со звёздочкой">

Сделай второй Dockerfile на `FROM scratch` для того же сервиса. Добавь в программу на время эксперимента HTTPS-запрос к `https://go.dev` и `time.LoadLocation("Europe/Moscow")`. Добейся, чтобы оба работали: скопируй CA-сертификаты из стадии сборки и подключи `time/tzdata`. Сравни размер с distroless-версией и сформулируй, что ты выиграл и что потерял. Не забудь удалить свои образы: `docker rmi <тег>`.

</Hw>
