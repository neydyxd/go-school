# gRPC и protobuf

REST с JSON хорош для публичного API: его понимает браузер, curl и любой язык, его легко читать глазами. Но когда внутри компании пятьдесят сервисов вызывают друг друга тысячи раз в секунду, всплывают его слабости. Контракт живёт в документации (или в голове), а не в коде: поле переименовали, клиент узнал об этом в проде. JSON медленно парсится и много весит. Стриминг и дедлайны каждый делает по-своему.

**gRPC** решает это иначе: контракт описывается в файле `.proto`, из него генерируется типизированный код клиента и сервера для любого языка, данные едут в компактном бинарном формате **protobuf** поверх HTTP/2. В Go-бэкенде gRPC встречается постоянно: Go сам пришёл из Google, где gRPC родился, а большинство микросервисных проектов на Go общаются между собой именно так. На собесе спросят, чем gRPC отличается от REST, как обеспечить обратную совместимость protobuf, какие бывают виды стриминга и что такое interceptor.

В уроке опишем сервис задач в `.proto`, сгенерируем код через **buf**, реализуем unary и стриминговые методы, напишем interceptors для авторизации и восстановления после паник, разберём коды ошибок и дедлайны.

## Protobuf: контракт и формат

Файл `homework/m10/l03-grpc/proto/tasks/v1/tasks.proto` (сокращённо):

```proto
syntax = "proto3";

package tasks.v1;

option go_package = "goschool/homework/m10/l03-grpc/gen/tasks/v1;tasksv1";

service TaskService {
  rpc CreateTask(CreateTaskRequest) returns (CreateTaskResponse);
  rpc GetTask(GetTaskRequest) returns (GetTaskResponse);
  rpc ListTasks(ListTasksRequest) returns (stream ListTasksResponse);
  rpc ImportTasks(stream ImportTasksRequest) returns (ImportTasksResponse);
}

enum Priority {
  PRIORITY_UNSPECIFIED = 0;
  PRIORITY_LOW = 1;
  PRIORITY_HIGH = 2;
}

message Task {
  int64 id = 1;
  string title = 2;
  bool done = 3;
  repeated string tags = 4;
  Priority priority = 5;
}

message CreateTaskRequest {
  string title = 1;
  repeated string tags = 2;
  Priority priority = 3;
}
```

- `package tasks.v1`: пространство имён protobuf. Версия в имени пакета (`v1`) это соглашение: несовместимые изменения делают в `v2` рядом, не ломая старых клиентов.
- `option go_package`: путь импорта и имя Go-пакета для сгенерированного кода.
- `message`: структура данных. У каждого поля тип, имя и **номер**.
- `repeated`: список. `enum`: перечисление.
- `service` и `rpc`: методы. `stream` перед типом означает поток сообщений.

Обрати внимание на отдельные `CreateTaskRequest` и `CreateTaskResponse` для каждого метода, даже если ответ просто оборачивает `Task`. Это правило линтера buf: у каждого метода своя пара сообщений, и в ответ потом можно добавить поле (например, `warnings`), не трогая другие методы.

### Номера полей это и есть контракт

В бинарном формате protobuf **нет имён полей**, только номера. Сообщение `Task{id: 7, title: "молоко"}` кодируется примерно так:

```text
08 07                 поле 1, тип varint, значение 7
12 0C d0 bc ...       поле 2, тип length-delimited, 12 байт UTF-8 «молоко»
```

Первый байт каждого поля это `(номер << 3) | тип_кодирования`. Поэтому протобаф компактный (нет ключей-строк, числа в varint: маленькие занимают один байт) и быстрый в разборе.

::: hood Значения по умолчанию и отсутствующие поля
В proto3 поле со значением по умолчанию (`0`, `""`, `false`, пустой список, нулевой элемент enum) **не передаётся вообще**. Пустой `CreateTaskRequest{}` занимает 0 байт, а `{title: "a"}` три байта. На принимающей стороне отсутствующее поле читается как значение по умолчанию.

Отсюда два следствия:

1. Нельзя отличить «поле не прислали» от «прислали ноль». Если это важно (PATCH-запрос «поставь приоритет 0» против «приоритет не трогай»), поле помечают `optional`: тогда у него появляется признак присутствия и метод `Has...` в Go. Для сообщений-полей присутствие есть всегда: `nil` против пустого сообщения.
2. Нулевой элемент enum получит любой клиент, который не заполнил поле, и любой старый клиент, который этого поля не знает. Поэтому он всегда должен значить «не задано» (`PRIORITY_UNSPECIFIED = 0`), а не реальное значение вроде `PRIORITY_LOW`.

Сгенерированные геттеры (`GetTitle()`, `GetTask()`) безопасны для `nil`: на `nil`-сообщении возвращают значение по умолчанию. Цепочка `resp.GetTask().GetTitle()` не паникует, даже если `resp` равен `nil`.
:::

### Обратная совместимость

Сервер и клиенты деплоятся независимо: какое-то время в проде живут и старые, и новые версии. Правила, которые нельзя нарушать:

| Изменение | Безопасно? |
|---|---|
| Добавить поле с новым номером | да: старые клиенты его пропустят |
| Переименовать поле, оставив номер | для бинарного формата да (имя не передаётся), но ломает JSON-представление и код клиентов при перегенерации |
| Удалить поле | только с `reserved`, чтобы номер и имя никто не занял снова |
| Изменить номер поля | нет: для старых клиентов это удаление старого и появление нового |
| Изменить тип (`int64` → `string`) | нет: несовместимое кодирование |
| Добавить значение в enum | да, но старые клиенты увидят незнакомое число |

```proto
message Task {
  reserved 3;
  reserved "done";
  int64 id = 1;
  string title = 2;
  // ...
}
```

Если номер удалённого поля переиспользовать под другое поле, старый клиент, который ещё шлёт `done = true` в поле 3, внезапно запишет туда же что-то совсем другое. `reserved` запрещает это на уровне компилятора.

Самое приятное: неизвестные поля не теряются. Старый код, который читает сообщение нового формата и пересылает его дальше, сохраняет незнакомые поля в отдельном буфере и записывает их обратно при сериализации. Ниже это задача «что выведет».

## buf: генерация и линтер

Классический инструмент это компилятор `protoc` с плагинами. Его неудобно ставить и настраивать, поэтому сейчас стандарт **buf**: одна Go-утилита, которая компилирует, проверяет стиль и ловит ломающие изменения.

```bash
go install github.com/bufbuild/buf/cmd/buf@latest
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.12
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.6.2
```

Конфигурация в `homework/m10/l03-grpc/`:

```yaml
# buf.yaml
version: v2
modules:
  - path: proto
lint:
  use:
    - STANDARD
breaking:
  use:
    - FILE
```

```yaml
# buf.gen.yaml
version: v2
plugins:
  - local: protoc-gen-go
    out: gen
    opt: paths=source_relative
  - local: protoc-gen-go-grpc
    out: gen
    opt: paths=source_relative
```

```bash
cd homework/m10/l03-grpc
buf lint        # стиль: имена, версии пакетов, нулевые элементы enum
buf generate    # gen/tasks/v1/tasks.pb.go и tasks_grpc.pb.go
```

Получаются два файла. `tasks.pb.go` от `protoc-gen-go`: Go-структуры сообщений, enum, геттеры, код сериализации. `tasks_grpc.pb.go` от `protoc-gen-go-grpc`: интерфейсы `TaskServiceClient` и `TaskServiceServer`, конструктор клиента и функция регистрации сервера. Сгенерированный код коммитят в репозиторий, чтобы `go build` и `go test` не требовали buf и плагинов.

`buf breaking` сравнивает контракт с предыдущей версией (git-веткой, тегом или папкой) и находит ломающие изменения. Я удалил поле `done`, поменял тип `GetTaskRequest.id` на `string` и добавил новое поле `assignee = 6`:

```text
$ buf breaking --against ../old
proto/tasks/v1/tasks.proto:28:1:Previously present field "3" with name "done" on message "Task" was deleted.
proto/tasks/v1/tasks.proto:47:3:Field "1" with name "id" on message "GetTaskRequest" changed type from "int64" to "string".
```

Новое поле претензий не вызвало. Категория `FILE` самая строгая: она ругается на удаление поля даже с `reserved`, потому что ломается сгенерированный код. Категория `WIRE_JSON` проверяет только совместимость на проводе и в JSON, и удаление с `reserved` пропускает. В CI `buf breaking --against '.git#branch=main'` стоит рядом с тестами.

## Сервер и клиент

Сгенерированный интерфейс сервера:

```go
type TaskServiceServer interface {
	CreateTask(context.Context, *CreateTaskRequest) (*CreateTaskResponse, error)
	GetTask(context.Context, *GetTaskRequest) (*GetTaskResponse, error)
	ListTasks(*ListTasksRequest, grpc.ServerStreamingServer[ListTasksResponse]) error
	ImportTasks(grpc.ClientStreamingServer[ImportTasksRequest, ImportTasksResponse]) error
	mustEmbedUnimplementedTaskServiceServer()
}
```

Неэкспортируемый метод `mustEmbed...` заставляет встроить в свою структуру `tasksv1.UnimplementedTaskServiceServer`. Зачем: когда в `.proto` добавят новый метод, твой сервер продолжит компилироваться, а новый метод будет отвечать `codes.Unimplemented`, пока ты его не напишешь. Встраивать по значению, не по указателю: с `nil`-указателем сгенерированная `RegisterTaskServiceServer` паникует уже при регистрации.

```go
type Server struct {
	tasksv1.UnimplementedTaskServiceServer

	mu     sync.Mutex
	tasks  []*tasksv1.Task
	nextID int64
}

func (s *Server) GetTask(ctx context.Context, req *tasksv1.GetTaskRequest) (*tasksv1.GetTaskResponse, error) {
	if req.GetId() <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "id должен быть положительным, получено %d", req.GetId())
	}
	// ... поиск под мьютексом ...
	return nil, status.Errorf(codes.NotFound, "задача %d не найдена", req.GetId())
}
```

Запуск сервера и клиента:

```go
// сервер
lis, _ := net.Listen("tcp", ":9090")
srv := grpc.NewServer()
tasksv1.RegisterTaskServiceServer(srv, NewServer())
srv.Serve(lis)

// клиент
conn, err := grpc.NewClient("localhost:9090",
	grpc.WithTransportCredentials(insecure.NewCredentials()))
if err != nil { ... }
defer conn.Close()
client := tasksv1.NewTaskServiceClient(conn)
resp, err := client.GetTask(ctx, &tasksv1.GetTaskRequest{Id: 1})
```

`grpc.NewClient` не устанавливает соединение сразу: оно поднимается при первом вызове и само переподключается при обрыве. Один `*grpc.ClientConn` рассчитан на весь процесс и на конкурентное использование из многих горутин, создавать его на каждый запрос нельзя. `insecure` только для локальной разработки и тестов, в проде TLS (или mTLS между сервисами).

Для ручной проверки удобен **grpcurl** (curl для gRPC). Ему нужен либо `.proto`, либо включённая на сервере reflection (`reflection.Register(srv)`).

### Тесты через bufconn

Поднимать сервер на настоящем порту в тестах неудобно: порты конфликтуют, тесты медленнее. Пакет `google.golang.org/grpc/test/bufconn` даёт listener в памяти, но весь стек gRPC (HTTP/2, сериализация, metadata, интерцепторы, дедлайны) работает по-настоящему:

```go
lis := bufconn.Listen(1 << 20)
srv := grpc.NewServer()
tasksv1.RegisterTaskServiceServer(srv, NewServer())
go srv.Serve(lis)

conn, _ := grpc.NewClient("passthrough:///bufnet",
	grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
		return lis.DialContext(ctx)
	}),
	grpc.WithTransportCredentials(insecure.NewCredentials()),
)
```

Схема `passthrough:///` говорит клиенту не резолвить имя через DNS, а отдать его прямо в dialer. Так устроены все тесты домашки.

## Четыре вида вызовов

```text
unary              клиент ── запрос ──▶ сервер ── ответ ──▶ клиент
server streaming   клиент ── запрос ──▶ сервер ══ ответ, ответ, ответ ══▶
client streaming   клиент ══ запрос, запрос, запрос ══▶ сервер ── ответ ──▶
bidi streaming     оба направления независимо, в одном вызове
```

**Server streaming** нужен, когда ответ большой или появляется постепенно: выгрузка списка, подписка на события, прогресс долгой операции.

```go
func (s *Server) ListTasks(req *tasksv1.ListTasksRequest, stream grpc.ServerStreamingServer[tasksv1.ListTasksResponse]) error {
	for _, t := range s.snapshot(req.GetOnlyOpen()) {
		if err := stream.Send(&tasksv1.ListTasksResponse{Task: t}); err != nil {
			return err // клиент ушёл или истёк дедлайн
		}
	}
	return nil // вернуть nil значит закончить поток
}
```

Клиент читает до `io.EOF`:

```go
stream, err := client.ListTasks(ctx, &tasksv1.ListTasksRequest{OnlyOpen: true})
for {
	resp, err := stream.Recv()
	if errors.Is(err, io.EOF) {
		break
	}
	if err != nil {
		return err
	}
	fmt.Println(resp.GetTask().GetTitle())
}
```

**Client streaming** для загрузки: импорт тысяч записей, отправка файла кусками. Сервер читает `Recv()` до `io.EOF` и отвечает один раз через `SendAndClose`. Клиент шлёт `Send()` и завершает `CloseAndRecv()`.

**Bidi streaming** для чатов, синхронизации, игр: обе стороны пишут и читают независимо. Обычно чтение и запись идут в разных горутинах.

::: trap Мьютекс во время Send
```go
// плохо: Send может ждать медленного клиента (flow control HTTP/2),
// а все остальные методы сервера стоят на этом мьютексе
s.mu.Lock()
defer s.mu.Unlock()
for _, t := range s.tasks {
	stream.Send(&tasksv1.ListTasksResponse{Task: t})
}
```

```go
// хорошо: снимок под мьютексом, отправка без него
s.mu.Lock()
snapshot := slices.Clone(s.tasks) // и копии самих сообщений, если их меняют
s.mu.Unlock()
for _, t := range snapshot {
	if err := stream.Send(&tasksv1.ListTasksResponse{Task: t}); err != nil {
		return err
	}
}
```

Общее правило конкурентного кода: никакого сетевого ввода-вывода под мьютексом.
:::

## Ошибки: status и codes

В gRPC ошибка это не HTTP-код, а **status**: код из фиксированного списка плюс сообщение (и необязательные детали). Сервер возвращает `status.Error(codes.NotFound, "...")`, клиент достаёт код через `status.Code(err)`.

| Код | Когда | Аналог HTTP |
|---|---|---|
| `OK` | успех | 200 |
| `InvalidArgument` | клиент прислал невалидные данные | 400 |
| `NotFound` | сущности нет | 404 |
| `AlreadyExists` | создаём то, что уже есть | 409 |
| `PermissionDenied` | известно кто, но нельзя | 403 |
| `Unauthenticated` | неизвестно кто: нет или плохой токен | 401 |
| `FailedPrecondition` | состояние системы не позволяет (удалить непустую папку) | 400/409 |
| `ResourceExhausted` | превышена квота, rate limit | 429 |
| `DeadlineExceeded` | истёк дедлайн | 504 |
| `Unavailable` | сервис временно недоступен, **можно повторить** | 503 |
| `Internal` | баг на сервере | 500 |
| `Unknown` | ошибка без статуса | 500 |

Если обработчик вернул обычный `errors.New("...")`, клиент получит `Unknown` с текстом ошибки. Это и некрасиво (клиент не может принять решение по коду), и опасно (текст внутренней ошибки, например SQL, уходит наружу). На границе сервиса доменные ошибки переводят в статусы явно.

`Unavailable` особенный: это единственный код, который безопасно ретраить почти всегда. `DeadlineExceeded` и `Internal` повторять стоит, только если операция идемпотентна (подробнее в [уроке про system design](./l05-system-design)).

```go
fmt.Println(status.Code(nil), status.Code(errors.New("x")))
// OK Unknown
```

## Дедлайны

В gRPC у каждого вызова может быть **дедлайн**, и он берётся из `context`:

```go
ctx, cancel := context.WithTimeout(ctx, 300*time.Millisecond)
defer cancel()
resp, err := client.GetTask(ctx, req)
if status.Code(err) == codes.DeadlineExceeded { ... }
```

Дедлайн уезжает на сервер в заголовке `grpc-timeout`, и там `ctx` обработчика получает тот же дедлайн (пересчитанный от момента приёма). Если сервер внутри вызывает базу и другие сервисы с этим `ctx`, дедлайн **распространяется по всей цепочке**: когда клиент перестал ждать, вся работа под ним отменяется, а не продолжает греть процессоры впустую.

Сервис без дедлайнов опасен: одна зависшая зависимость, и горутины с соединениями копятся, пока сервис не упадёт. Правило: у каждого исходящего вызова есть таймаут, а входящий `ctx` пробрасывается дальше, а не заменяется на `context.Background()`.

::: trap Потерянный контекст
```go
// плохо: дедлайн и отмена клиента до базы не доходят
func (s *Server) GetTask(ctx context.Context, req *tasksv1.GetTaskRequest) (*tasksv1.GetTaskResponse, error) {
	row := s.db.QueryRow(context.Background(), "SELECT ...", req.GetId())
	// ...
}
```

```go
// хорошо: ctx обработчика идёт во все вызовы ниже
row := s.db.QueryRow(ctx, "SELECT ...", req.GetId())
```

То же на клиенте: функция-обёртка над вызовом должна строить таймаут от переданного `ctx` (`context.WithTimeout(ctx, d)`), а не от `Background`, иначе отмена родителя (например, входящего HTTP-запроса) потеряется.
:::

## Interceptors

**Interceptor** это middleware для gRPC: функция, которая оборачивает каждый вызов. Сквозная логика живёт в них: авторизация, логирование, метрики, трейсинг, восстановление после паник, rate limiting.

```go
type UnaryServerInterceptor func(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo, // info.FullMethod = "/tasks.v1.TaskService/GetTask"
	handler grpc.UnaryHandler,   // следующий в цепочке или сам метод
) (resp any, err error)
```

Для стриминговых методов отдельный тип `StreamServerInterceptor` с `grpc.ServerStream` вместо `req`. Есть и клиентские интерцепторы (`UnaryClientInterceptor`): добавить токен, ретраи, метрики исходящих вызовов.

Заголовки в gRPC называются **metadata**. На сервере их читают из контекста:

```go
func AuthUnary(token string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		md, _ := metadata.FromIncomingContext(ctx)
		if !slices.Contains(md.Get("authorization"), "Bearer "+token) {
			return nil, status.Error(codes.Unauthenticated, "нужен токен")
		}
		return handler(ctx, req)
	}
}
```

Клиент добавляет их в исходящий контекст:

```go
ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
```

Ключи metadata всегда в нижнем регистре (`md.Get` сам приводит регистр). Сравнивать секреты лучше через `crypto/subtle.ConstantTimeCompare`, чтобы время сравнения не подсказывало атакующему, сколько первых символов он угадал.

Подключение цепочки:

```go
srv := grpc.NewServer(
	grpc.ChainUnaryInterceptor(RecoveryUnary(log), AuthUnary(token)),
	grpc.ChainStreamInterceptor(AuthStream(token)),
)
```

Первый в списке самый внешний: `Recovery` оборачивает `Auth`, а `Auth` оборачивает обработчик. Recovery ставят первым, чтобы он ловил паники во всём, что внутри.

::: hood Почему паника роняет весь сервер
gRPC-сервер обрабатывает каждый вызов в своей горутине. В Go паника, не пойманная `recover` в **той же горутине**, завершает весь процесс: со всеми остальными запросами и соединениями. `net/http` сам ловит паники в обработчиках (и пишет их в лог), а grpc-go этого не делает. Поэтому recovery interceptor в gRPC-сервисе не опция, а обязательная часть. Ловит он только паники в горутине вызова: если обработчик запустил свою горутину и паника случилась там, interceptor не поможет.
:::

## gRPC против REST

| | REST + JSON | gRPC + protobuf |
|---|---|---|
| Контракт | OpenAPI (если не забыли обновить) | `.proto`, код генерируется |
| Формат | текст, читается глазами | бинарный, компактнее и быстрее |
| Транспорт | HTTP/1.1 или 2 | только HTTP/2 |
| Стриминг | SSE, WebSocket, отдельно | встроен, 4 вида |
| Дедлайны | каждый сам | встроены, распространяются |
| Браузер | нативно | нужен gRPC-Web или прокси |
| Отладка | curl, DevTools | grpcurl, нужен контракт |
| Кеширование HTTP | GET кешируется CDN | нет |

Типичная архитектура: **снаружи REST** (браузеры, мобильные клиенты, партнёры), **внутри gRPC** между сервисами. Для внешнего API поверх gRPC-сервиса часто ставят grpc-gateway, который генерирует REST-прокси из того же `.proto`.

Одна неочевидная особенность: HTTP/2 держит **одно долгое соединение** и мультиплексирует в нём все вызовы. Балансировщик уровня L4 (TCP) распределяет соединения, а не запросы, поэтому все вызовы клиента уйдут на один под, даже если подов десять. Решения: балансировка на клиенте (`round_robin` в grpc-go с DNS, возвращающим все адреса подов), L7-балансировщик, понимающий HTTP/2 (Envoy), или service mesh.

## Что выведет программа?

```go
var resp *tasksv1.GetTaskResponse
fmt.Println(resp.GetTask().GetTitle() == "", resp.GetTask().GetId())
```

<Predict :options="['true 0', 'panic: nil pointer dereference', 'false 0', 'ошибка компиляции']" :correct="0">

Сгенерированные геттеры проверяют получателя на `nil` и возвращают значение по умолчанию. `resp.GetTask()` на `nil` даёт `nil` `*Task`, у которого `GetTitle()` возвращает `""`, а `GetId()` ноль. Поэтому в коде с protobuf принято читать поля через геттеры, а не напрямую `resp.Task.Title`.

</Predict>

```go
wrapped := fmt.Errorf("repo: %w", status.Error(codes.NotFound, "нет"))
st, _ := status.FromError(wrapped)
fmt.Println(status.Code(wrapped), "|", st.Message())
```

<Predict :options="['Unknown | repo: rpc error: code = NotFound desc = нет', 'NotFound | нет', 'NotFound | repo: rpc error: code = NotFound desc = нет', 'Unknown | нет']" :correct="2">

`status.Code` и `status.FromError` умеют искать статус в цепочке обёрток через `%w`, поэтому код `NotFound`. Но сообщение `FromError` берёт целиком из текста обёрнутой ошибки, с префиксом `repo:` и строкой `rpc error: ...`. Если хочешь чистое сообщение клиенту, не оборачивай статус, а создавай его на границе сервиса.

</Predict>

Сообщение нового формата прочитал старый сервис, который знает только `GetTaskRequest{int64 id = 1}`, и переслал дальше:

```go
full, _ := proto.Marshal(&tasksv1.Task{Id: 7, Title: "молоко", Done: true})

var old tasksv1.GetTaskRequest
proto.Unmarshal(full, &old)
again, _ := proto.Marshal(&old)

var back tasksv1.Task
proto.Unmarshal(again, &back)
fmt.Println(old.GetId(), back.GetId(), back.GetTitle(), back.GetDone())
```

<Predict :options="['7 7 молоко true', '7 7  false', 'ошибка Unmarshal: неизвестные поля', '0 0  false']" :correct="0">

Поле 1 совпало по номеру и типу, поэтому `old.GetId()` равно 7. Поля 2 и 3 старый тип не знает, но protobuf-go не выбрасывает их, а хранит как unknown fields (здесь 16 байт) и записывает обратно при `Marshal`. Новый получатель видит сообщение целиком. Именно так старые сервисы-посредники не портят данные, добавленные новыми версиями.

</Predict>

::: interview
<details><summary>Чем gRPC отличается от REST? Когда что выбрать?</summary>

gRPC: контракт в `.proto` с генерацией кода, бинарный protobuf, HTTP/2 с мультиплексированием, встроенные стриминг (4 вида), дедлайны и коды ошибок. REST: текстовый JSON, работает везде включая браузер, легко отлаживать, кешируется HTTP-инфраструктурой. Типично gRPC между внутренними сервисами, REST для внешних клиентов и браузера (или grpc-gateway поверх gRPC).

</details>
<details><summary>Как менять .proto, не ломая клиентов?</summary>

Только добавлять поля с новыми номерами. Номера и типы существующих полей не менять. Удалённые поля помечать `reserved` (номер и имя), чтобы их не заняли повторно. Нулевой элемент enum должен означать «не задано». Несовместимые изменения делать в новой версии пакета (`v2`). Проверять автоматически через `buf breaking` в CI.

</details>
<details><summary>Почему в protobuf важны номера полей, а не имена?</summary>

В бинарном формате передаются только номер поля и тип кодирования, имён нет. Поэтому переименование поля бинарно совместимо, а смена номера равна удалению и добавлению поля. Неизвестные номера получатель пропускает, но сохраняет как unknown fields и при пересылке записывает обратно.

</details>
<details><summary>Какие виды вызовов есть в gRPC?</summary>

Unary (запрос-ответ), server streaming (один запрос, поток ответов: выгрузки, подписки), client streaming (поток запросов, один ответ: загрузка, импорт), bidirectional streaming (оба потока независимо в одном вызове: чат, синхронизация). Поток завершается: сервер возвращает из обработчика, клиент получает `io.EOF`.

</details>
<details><summary>Что такое interceptor и что в нём обычно делают?</summary>

Middleware для gRPC-вызовов, отдельно для unary и stream, отдельно для сервера и клиента. В них сквозная логика: авторизация по metadata, логирование, метрики, трейсинг, recovery после паник, rate limiting, на клиенте ретраи и добавление токена. Цепочка собирается через `ChainUnaryInterceptor`, первый самый внешний.

</details>
<details><summary>Как работают дедлайны в gRPC?</summary>

Клиент задаёт дедлайн через `context.WithTimeout`, он передаётся в заголовке `grpc-timeout`, и контекст обработчика на сервере получает тот же дедлайн. Если сервер передаёт этот `ctx` дальше в базу и другие сервисы, отмена распространяется по всей цепочке. По истечении клиент получает `codes.DeadlineExceeded`. Без дедлайнов зависшая зависимость копит горутины и соединения.

</details>
<details><summary>Какие коды ошибок можно ретраить?</summary>

`Unavailable` почти всегда: сервис временно недоступен, запрос не обработан. `DeadlineExceeded`, `ResourceExhausted`, `Aborted` с осторожностью и только для идемпотентных операций, с экспоненциальной задержкой. `InvalidArgument`, `NotFound`, `PermissionDenied`, `Unauthenticated` ретраить бессмысленно: повтор даст тот же результат.

</details>
<details><summary>Почему gRPC плохо балансируется обычным балансировщиком?</summary>

HTTP/2 держит долгое соединение и мультиплексирует в нём все вызовы. L4-балансировщик распределяет TCP-соединения, поэтому весь трафик клиента уходит на один бэкенд. Нужна балансировка на клиенте (resolver с адресами всех подов и политика `round_robin`), L7-прокси с поддержкой HTTP/2 (Envoy) или service mesh.

</details>
:::

## Источники

- `RU` [Что нужно знать о gRPC системному аналитику](https://habr.com/ru/companies/tbank/articles/780024/): Т-Банк про устройство gRPC, виды стриминга и контракт; без кода, но с хорошими схемами.
- `RU` [Сериализация данных в Golang с Protobuf](https://habr.com/ru/companies/otus/articles/784732/): protobuf в Go с установкой генераторов и примерами.
- `EN` [Basics tutorial, gRPC Go](https://grpc.io/docs/languages/go/basics/): официальный разбор всех четырёх видов вызовов на примере route_guide.
- `EN` [Language Guide (proto 3)](https://protobuf.dev/programming-guides/proto3/): типы, значения по умолчанию, `optional`, `reserved`, правила обновления сообщений.
- `EN` [Proto Best Practices](https://protobuf.dev/best-practices/dos-donts/): чего не делать с полями и enum.
- `EN` [Status Codes](https://grpc.io/docs/guides/status-codes/) и [Deadlines](https://grpc.io/docs/guides/deadlines/): какие коды когда возвращать и как распространяются дедлайны.
- `EN` [bufbuild/buf](https://github.com/bufbuild/buf): README и ссылки на документацию по lint и breaking.

## Домашка

Контракт, конфиги buf и сгенерированный код уже лежат в `homework/m10/l03-grpc/`. Все тесты ходят в сервер через bufconn и ничего не слушают в сети.

<Hw id="ex1" title="Unary-методы и коды ошибок" level="база">

Файл `homework/m10/l03-grpc/ex1_server.go`: хранилище в `Server`, `CreateTask`, `GetTask` и `SetDone`. Валидация с `InvalidArgument`, `NotFound`, длина заголовка в символах, а не байтах, ответ не должен давать доступ к внутреннему хранилищу. Запускай с `-race`: методы вызываются конкурентно.

```bash
cd homework && go test -race -v ./m10/l03-grpc/ -run 'TestCreate|TestServer|TestSetDone'
```

</Hw>

<Hw id="ex2" title="Server и client streaming" level="база">

Файл `homework/m10/l03-grpc/ex2_stream.go`: `ListTasks` (поток по возрастанию id, фильтр `only_open`, без мьютекса во время `Send`) и `ImportTasks` (читать до `io.EOF`, ошибки по строкам не обрывают импорт).

```bash
cd homework && go test -race -v ./m10/l03-grpc/ -run 'TestList|TestImport'
```

</Hw>

<Hw id="ex3" title="Interceptors: авторизация и recovery" level="посложнее">

Файл `homework/m10/l03-grpc/ex3_interceptors.go`: `AuthUnary` и `AuthStream` с публичными методами, `RecoveryUnary`, который превращает панику в `Internal` без подробностей для клиента и пишет её в лог.

```bash
cd homework && go test -race -v ./m10/l03-grpc/ -run 'TestAuth|TestRecovery'
```

</Hw>

<Hw id="ex4" title="Клиент с дедлайном" level="посложнее">

Файл `homework/m10/l03-grpc/ex4_client.go`, функция `FetchTask`: таймаут от переданного контекста, перевод `NotFound` и `DeadlineExceeded` в доменные ошибки, остальное обернуть с сохранением кода. Тест проверяет, что дедлайн действительно доехал до сервера.

```bash
cd homework && go test -race -v ./m10/l03-grpc/ -run TestFetchTask
```

</Hw>

<Hw id="ex5" title="Эволюция контракта" level="со звёздочкой">

Поставь buf и плагины (команды в разделе про buf). В `proto/tasks/v1/tasks.proto` добавь в `Task` поле `string assignee = 6`, перегенерируй код (`buf lint && buf generate` из `homework/m10/l03-grpc`) и убедись, что тесты по-прежнему зелёные. Потом скопируй папку `proto` во временную директорию, удали там поле `done` и запусти `buf breaking --against <копия исходной папки>`: найди нарушение, исправь через `reserved`, сравни поведение категорий `FILE` и `WIRE_JSON`. В конце верни `tasks.proto` и сгенерированный код к исходному виду.

</Hw>

<Hw id="ex6" title="Живой сервер и grpcurl" level="со звёздочкой">

Вне папки домашки напиши `main.go`, который поднимает твой `Server` на `:9090` с интерцепторами из ex3 и `reflection.Register(srv)`, и graceful shutdown через `srv.GracefulStop()` по SIGTERM. Поставь grpcurl (`brew install grpcurl`) и вызови методы руками:

```bash
grpcurl -plaintext localhost:9090 list
grpcurl -plaintext -H 'authorization: Bearer s3cret' -d '{"title":"молоко"}' localhost:9090 tasks.v1.TaskService/CreateTask
grpcurl -plaintext -d '{"id": 1}' localhost:9090 tasks.v1.TaskService/GetTask
```

Посмотри, как выглядит ошибка без токена и что отдаёт `ListTasks`.

</Hw>
