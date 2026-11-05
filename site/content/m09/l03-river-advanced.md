# River: ретраи, уникальность, периодика

Основы River у тебя есть: аргументы, воркер, клиент, `InsertTx`. В реальном сервисе сразу всплывают вопросы второго порядка. Внешний API лежит полчаса: сколько раз и как часто повторять? Выгрузка ещё готовится: как подождать, не тратя попытки? Пользователь трижды нажал «Сформировать отчёт»: как не поставить три одинаковые задачи? Дайджест каждое утро в девять: кто его ставит, если реплик сервиса три? И как всё это тестировать без `time.Sleep`?

Этот урок про управление жизнью задачи: ретраи и backoff, snooze, отмена, уникальность и что именно она гарантирует, периодические и отложенные задачи, очереди и приоритеты, таймауты, River UI и тестирование через `rivertest`. В конце ТЗ проекта модуля: уведомления и дайджест в «Трекере задач».

## Ретраи и backoff

Задача, чей `Work` вернул ошибку, не пропадает: River записывает ошибку в `errors` (массив jsonb с текстом, номером попытки и временем) и планирует повтор. Попытки считаются полем `attempt`, лимит `max_attempts`. Когда `attempt` дошёл до `max_attempts` и снова ошибка, задача уходит в `discarded` и больше не выполняется.

**Политика по умолчанию**: задержка `n^4` секунд, где `n` это номер ошибки по счёту, плюс случайный джиттер в пределах 10%:

| Ошибка | Задержка |
|---|---|
| 1 | ~1 с |
| 2 | ~16 с |
| 3 | ~1 мин 21 с |
| 4 | ~4 мин 16 с |
| 5 | ~10 мин 25 с |
| 24 | ~92 ч |

С `MaxAttempts` по умолчанию 25 сумма задержек 24 повторов около 490 часов, то есть **примерно 20 дней**. Для письма о сбросе пароля это бессмысленно: через час пользователь уже ушёл. Подбирай `MaxAttempts` под смысл задачи.

### Свой backoff

Два уровня настройки:

- `Config.RetryPolicy` (интерфейс `ClientRetryPolicy` с методом `NextRetry(*rivertype.JobRow) time.Time`) на весь клиент;
- метод `NextRetry(job *river.Job[T]) time.Time` у воркера, только для его задач. `WorkerDefaults` возвращает нулевое время, и тогда работает политика клиента.

```go
func (w *ReportWorker) NextRetry(job *river.Job[ReportArgs]) time.Time {
	n := len(job.Errors) + 1 // текущая ошибка ещё не дописана в Errors
	return time.Now().Add(Backoff(n, 30*time.Second, time.Hour, rand.Float64()))
}
```

::: hood Что River делает с ответом NextRetry
Код из `internal/jobexecutor` v0.51.0, упрощённо:

```text
attempt >= max_attempts           -> discarded
t := worker.NextRetry(job)
t.IsZero()                        -> t = client.RetryPolicy.NextRetry(job)
t раньше now                      -> предупреждение в лог, t = политика по умолчанию
t - now <= 5 с (интервал scheduler) -> state = available, scheduled_at = t
иначе                             -> state = retryable,  scheduled_at = t
```

Задачи в `retryable` переводит в `available` планировщик лидера (раз в 5 секунд), поэтому для коротких задержек River ставит `available` сразу, иначе повтор через 1 секунду фактически ждал бы до 5. Номер ошибки считается как `len(job.Errors) + 1`, а не по `attempt`: так расписание не сбивается, если задача раньше засыпала (snooze).
:::

::: trap Ретраи для постоянной ошибки
```go
// плохо: 25 попыток за 20 дней для пользователя, которого больше нет
user, err := w.users.Get(ctx, job.Args.UserID)
if err != nil {
	return err
}
```

```go
// хорошо: постоянную ошибку не повторяем
user, err := w.users.Get(ctx, job.Args.UserID)
if errors.Is(err, ErrNotFound) {
	return river.JobCancel(fmt.Errorf("пользователь %d удалён: %w", job.Args.UserID, err))
}
if err != nil {
	return err // временная: база, сеть
}
```

Каждая лишняя попытка это нагрузка, шум в логах и алертах. Разделяй ошибки на временные (вернуть как есть) и постоянные (`JobCancel`).
:::

### Хук на ошибки

`Config.ErrorHandler` вызывается на каждую ошибку и панику задачи: туда подключают Sentry, логи с контекстом, метрики. Он тоже может отменить задачу через `ErrorHandlerResult{SetCancelled: true}`, это удобно, когда решение о постоянстве ошибки принимается централизованно.

## Snooze: подождать без траты попытки

Иногда задача не упала, а **рано**: выгрузка у партнёра ещё готовится, сработал rate limit, ждём, пока завершится соседняя операция. Ошибка тут неправильный сигнал: потратится попытка, запишется ошибка, сработает алерт.

```go
switch status {
case "pending":
	return river.JobSnooze(30 * time.Second) // через 30 с ещё раз, attempt не растёт
case "failed":
	return river.JobCancel(fmt.Errorf("выгрузка %s провалилась", id))
}
```

Snooze переводит задачу в `scheduled` (или сразу в `available`, если задержка не больше 5 секунд), увеличивает счётчик `snoozes` в метаданных и **не тратит попытку**: River уменьшает `attempt` обратно. `JobSnooze(0)` делает задачу доступной немедленно. Прерывание остановкой клиента так обрабатывать не нужно: с v0.44 River сам возвращает в `available` без траты попытки задачу, которая вернула `ctx.Err()` из-за `Stop` или `StopAndCancel` (урок 2). `JobSnooze(0)` нужен, когда ту же семантику хочется для своих причин: например, воркер сам заметил, что зависимость переключается, и хочет уступить задачу без ошибки. Отрицательная длительность вызывает панику.

Snooze можно повторять бесконечно, поэтому думай о верхней границе сам: считай ожидание через `job.CreatedAt` и отменяй задачу, если ждёшь слишком долго.

## Отмена

Изнутри задачи: `return river.JobCancel(err)`. Задача сразу `cancelled`, ошибка записана, повторов не будет.

Снаружи: `client.JobCancel(ctx, jobID)`. Если задача ещё в очереди (`available`, `scheduled`, `retryable`), она сразу становится `cancelled`. Если она **выполняется**, River помечает её на отмену и через уведомление отменяет её `ctx` на том клиенте, который её выполняет. Горутину в Go прервать нельзя, поэтому задача доработает до своего `return`: вернёт ошибку, и задача окончательно отменится; вернёт `nil`, и она будет `completed`. Ещё одна причина слушать `ctx`.

## Уникальные задачи

Пользователь три раза нажал «Сформировать отчёт», вебхук от партнёра пришёл дважды, периодическая задача поставилась двумя лидерами при смене. Хочется: «если такая задача уже есть, не ставь вторую».

```go
type DigestArgs struct {
	UserID      int64  `json:"user_id" river:"unique"` // только это поле в ключе
	RequestedBy string `json:"requested_by"`
}

func (DigestArgs) Kind() string { return "digest" }

func (DigestArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{UniqueOpts: river.UniqueOpts{
		ByArgs:   true,
		ByPeriod: 24 * time.Hour,
	}}
}
```

Измерения уникальности, все складываются (AND):

- **Kind** входит всегда (если не `ExcludeKind`);
- `ByArgs`: аргументы целиком или только поля с тегом `river:"unique"`;
- `ByPeriod`: время запуска (`ScheduledAt` или «сейчас») округляется вниз до кратного периоду **в UTC**, и в одном окне может быть одна задача;
- `ByQueue`: уникальность в пределах очереди;
- `ByState`: в каких состояниях существующая задача блокирует новую. По умолчанию все, кроме `cancelled` и `discarded`.

На дубле `Insert` **не возвращает ошибку**: в `JobInsertResult` будет `UniqueSkippedAsDuplicate: true`, а в `Job` уже существующая задача.

```go
opts := &river.InsertOpts{UniqueOpts: river.UniqueOpts{ByArgs: true}}
r1, _ := client.Insert(ctx, ReportArgs{ReportID: 9}, opts)
r2, _ := client.Insert(ctx, ReportArgs{ReportID: 9}, opts)
fmt.Println(r1.UniqueSkippedAsDuplicate, r2.UniqueSkippedAsDuplicate, r1.Job.ID == r2.Job.ID)
```

<Predict :options="['false false false', 'false true true', 'false true false', 'ошибка unique_violation на второй вставке']" :correct="1">

Первая вставка новая, вторая распознана как дубль: флаг `true`, и в `r2.Job` лежит та же задача, что в `r1`. Ошибки нет, это нормальный путь. Если тебе важно отличать «поставил» от «уже было» (например, ответить пользователю «отчёт уже готовится»), смотри на флаг.

</Predict>

::: hood Как устроена уникальность
River считает `unique_key`: хеш от Kind и выбранных измерений (аргументы с отсортированными ключами верхнего уровня, округлённое время, очередь) и кладёт его в колонку `river_job.unique_key` вместе с битовой маской `unique_states`. На таблице частичный уникальный индекс `river_job_unique_idx ON river_job (unique_key) WHERE unique_key IS NOT NULL AND ...` с условием по состояниям. Дубль упирается в индекс, River ловит конфликт и возвращает существующую строку. Отсюда и требование: `available`, `pending`, `running` и `scheduled` обязаны быть в `ByState`, иначе задаче было бы некуда переходить без конфликта.
:::

**Что уникальность гарантирует, а что нет.** Она гарантирует, что **вставка** второй такой задачи не пройдёт, пока первая в одном из состояний `ByState`. Она **не** гарантирует, что задача выполнится один раз: River остаётся at-least-once, rescuer может вернуть зависшую задачу, а воркер мог успеть сделать эффект. Ещё два тонких момента: завершённые задачи удаляет чистильщик (по умолчанию через 24 часа), после этого та же задача снова вставится; а отменённая или отброшенная задача не блокирует новую.

::: trap Уникальность по всем аргументам
```go
// плохо: RequestID в каждой вставке разный, значит, и ключ разный
type ReportArgs struct {
	UserID    int64  `json:"user_id"`
	RequestID string `json:"request_id"` // для логов
}
// UniqueOpts{ByArgs: true}: три клика дают три задачи
```

```go
// хорошо: в ключ попадает только то, что определяет работу
type ReportArgs struct {
	UserID    int64  `json:"user_id" river:"unique"`
	RequestID string `json:"request_id"`
}
```

Любое поле «для диагностики» (trace id, время запроса, инициатор) в ключе превращает уникальность в фикцию. Помечай тегом только поля, которые определяют, что делать.
:::

## Отложенные задачи

`InsertOpts.ScheduledAt` ставит задачу на будущее: она будет `scheduled`, планировщик лидера переведёт её в `available`, когда подойдёт время. Гарантия «не раньше», а не «ровно в»: задача может стартовать немного позже, в зависимости от загрузки очереди.

```go
// напомнить о задаче за час до срока
_, err := client.InsertTx(ctx, tx, RemindArgs{TaskID: t.ID}, &river.InsertOpts{
	ScheduledAt: t.DueAt.Add(-time.Hour),
})
```

Если срок задачи потом поменяют, старое напоминание останется. Либо отменяй его (`JobCancel` по сохранённому id задачи), либо делай воркер, который перечитывает актуальный срок и молча выходит, если напоминание устарело. Второе надёжнее.

## Периодические задачи

```go
periodic := river.NewPeriodicJob(
	river.PeriodicInterval(15*time.Minute),
	func() (river.JobArgs, *river.InsertOpts) {
		return CleanupArgs{}, nil
	},
	&river.PeriodicJobOpts{ID: "cleanup", RunOnStart: true},
)

client, err := river.NewClient(riverpgxv5.New(pool), &river.Config{
	PeriodicJobs: []*river.PeriodicJob{periodic},
	// ... Queues, Workers
})
```

- Расписание это любой тип с методом `Next(time.Time) time.Time` (`river.PeriodicSchedule`). Для cron-выражений подходит `github.com/robfig/cron/v3`: его расписание реализует тот же метод. В домашке ты напишешь своё «каждый день в 09:00 по Москве».
- Ставит периодические задачи только **лидер**. Поэтому одна и та же конфигурация `PeriodicJobs` должна быть у всех клиентов, которые могут стать лидером, иначе при смене лидера расписание пропадёт.
- Планировщик **без состояния**: время следующего запуска хранится в памяти лидера. Перезапуск или смена лидера начинает отсчёт заново, и запуск может пропуститься. `RunOnStart: true` ставит задачу сразу при старте лидера, а чтобы при этом не получить два запуска за день, добавь уникальность `ByPeriod`.
- Конструктор вызывается при каждом срабатывании: в нём можно вычислить аргументы (например, «за какой день дайджест»).
- Задачи можно добавлять и убирать на ходу: `client.PeriodicJobs().Add(...)`, `RemoveByID("cleanup")`.

::: trap Периодическая задача через time.Ticker в каждой реплике
```go
// плохо: три реплики, три дайджеста каждое утро
go func() {
	for range time.Tick(24 * time.Hour) {
		sendDigest(ctx)
	}
}()
```

Каждая реплика сервиса запустит свою горутину. River решает это лидером, а страховку от двойного запуска даёт `UniqueOpts{ByPeriod: 24 * time.Hour}`: даже если два клиента поставят задачу за одни сутки, вставится одна.
:::

## Очереди, приоритеты, таймауты

**Очереди** изолируют разные типы работы: у каждой свой пул горутин (`MaxWorkers`) в каждом процессе. Медленные отчёты не мешают письмам. Очередь можно поставить на паузу (`client.QueuePause(ctx, "email", nil)`) на время инцидента у провайдера и потом возобновить (`QueueResume`), задачи при этом копятся, а не падают.

**Приоритет** 1-4 (1 самый высокий) работает внутри одной очереди: producer выбирает `ORDER BY priority, scheduled_at, id`. Если высокоприоритетных задач всегда больше, чем воркеров, низкие не выполнятся никогда. Поэтому для действительно разной важности надёжнее отдельные очереди.

**Таймаут** задачи по умолчанию 1 минута (`Config.JobTimeout`), для отдельного воркера метод `Timeout(job) time.Duration` (`-1` значит без таймаута). По истечении River отменяет `ctx` задачи. Не путай с `RescueStuckJobsAfter` (по умолчанию час): это срок, после которого лидер считает задачу в `running` зависшей (процесс умер) и возвращает её. Он обязан быть больше таймаута, иначе живые задачи будут возвращаться в очередь.

```text
Задача с MaxAttempts 1 трижды вернула river.JobSnooze(time.Minute),
а на четвёртом запуске вернула nil.
```

<Predict :options="['discarded после первого snooze', 'completed, attempt = 1', 'completed, attempt = 4', 'cancelled']" :correct="1">

Snooze не тратит попытки: после каждого River возвращает `attempt` к прежнему значению, поэтому даже с `MaxAttempts: 1` задача не отбрасывается. Четвёртый запуск засчитан как первая настоящая попытка, она успешна. Тест `TestExportWorkerSnoozeDoesNotDiscard` в домашке проверяет именно это.

</Predict>

## River UI

Веб-интерфейс для просмотра задач по состояниям, очередей, ошибок с трейсами; позволяет повторить, отменить или удалить задачу вручную и поставить очередь на паузу. Открытый (`riverui`) и платная версия River Pro.

```bash
docker run -p 8080:8080 --env DATABASE_URL="postgres://goschool:goschool@host.docker.internal:55432/tasktracker" \
  ghcr.io/riverqueue/riverui:latest
```

Есть и бинарник (`go install riverqueue.com/riverui/cmd/riverui@latest`), и встраивание в свой сервис как `http.Handler` (`riverui.NewHandler`). Без переменных `RIVER_BASIC_AUTH_USER` и `RIVER_BASIC_AUTH_PASS` UI открыт всем, кто достучится до порта, а в аргументах задач лежат персональные данные. В проде его прячут за аутентификацию.

## Тестирование: rivertest

Три уровня тестов задач.

**1. Вставка.** Проверяем, что бизнес-код поставил нужную задачу с нужными опциями:

```go
// в транзакции, до коммита
job := rivertest.RequireInsertedTx[*riverpgxv5.Driver](ctx, t, tx, &WelcomeEmailArgs{},
	&rivertest.RequireInsertedOpts{Queue: "email", MaxAttempts: 5})
fmt.Println(job.Args.UserID)

// через пул, после коммита или отката
rivertest.RequireInserted(ctx, t, riverpgxv5.New(pool), &WelcomeEmailArgs{}, nil)
rivertest.RequireNotInserted(ctx, t, riverpgxv5.New(pool), &WelcomeEmailArgs{}, nil)
```

`RequireInserted` падает, если задач такого Kind больше одной (для нескольких есть `RequireManyInserted` с порядком). Возвращает задачу с типизированными аргументами.

**2. Воркер.** `rivertest.Worker` вставляет задачу в переданную транзакцию и выполняет её **синхронно**, со всем окружением River (контекст, middleware, `ClientFromContext`):

```go
tw := rivertest.NewWorker(t, riverpgxv5.New(nil), &river.Config{}, &ExportWorker{...})
res, err := tw.Work(ctx, t, tx, ExportArgs{ExportID: "exp-1"}, nil)
// res.EventKind: job_completed, job_failed, job_snoozed, job_cancelled
// res.Job: строка ПОСЛЕ выполнения: State, Attempt, ScheduledAt, Errors
```

`err` содержит только настоящие ошибки `Work`, snooze и cancel ошибкой не считаются. Драйвер без пула (`riverpgxv5.New(nil)`) нормален: всё идёт через транзакцию, которую тест в конце откатывает. Уникальность в таком воркере отключена, чтобы параллельные тесты не конфликтовали. Повторно выполнить ту же задачу (второй ретрай) можно через `tw.WorkJob(ctx, t, tx, res.Job)`.

**3. Сквозной.** Настоящий клиент с очередями, вставка, ожидание события через `client.Subscribe` с таймаутом (`select` с `time.After`), `Stop` в конце. Медленнее, но проверяет конфиг: зарегистрирован ли воркер, есть ли очередь. `Config.TestOnly: true` убирает случайные задержки старта служебных процессов.

::: trap time.Sleep в тесте очереди
```go
// плохо: на медленном CI не успеет, на быстром ждёт зря
client.Insert(ctx, args, nil)
time.Sleep(2 * time.Second)
assertEmailSent(t)
```

```go
// хорошо: ждём событие, таймаут только как предохранитель
events, cancel := client.Subscribe(river.EventKindJobCompleted)
defer cancel()
client.Insert(ctx, args, nil)
select {
case <-events:
case <-time.After(10 * time.Second):
	t.Fatal("задача не выполнилась за 10 секунд")
}
```

А логику воркера вообще проверяй через `rivertest.Worker`, без клиента и без ожиданий.
:::

## Что выведет программа?

```go
// DefaultClientRetryPolicy, задача падает каждый раз
// через сколько примерно River повторит её после 4-й ошибки подряд?
```

<Predict :options="['16 секунд', '1 минута 21 секунда', '4 минуты 16 секунд', '8 секунд']" :correct="2">

Задержка `n^4` секунд, где `n` номер ошибки: 4^4 = 256 секунд, это 4 мин 16 с (плюс-минус 10% джиттера). Политика считает `n` по `len(job.Errors) + 1`, а не по `attempt`, чтобы snooze не сбивал расписание.

</Predict>

```go
// DigestArgs из урока: UniqueOpts{ByArgs: true, ByPeriod: 24 * time.Hour}
day := time.Date(2030, 3, 10, 0, 0, 0, 0, time.UTC)
_, dup1, _ := EnqueueDigest(ctx, client, 1, "кнопка", day.Add(23*time.Hour+59*time.Minute))
_, dup2, _ := EnqueueDigest(ctx, client, 1, "кнопка", day.Add(24*time.Hour+time.Minute))
fmt.Println(dup1, dup2)
```

<Predict :options="['false true', 'false false', 'true true', 'true false']" :correct="1">

`ByPeriod` округляет время запуска вниз до кратного 24 часам **в UTC**: 23:59 10 марта и 00:01 11 марта попадают в разные сутки, поэтому вторая задача не дубль, хотя между ними две минуты. Уникальность «раз в сутки» это «раз в календарные сутки UTC», а не «не чаще раза в 24 часа». Для Москвы окно сдвинуто на 3 часа, это надо помнить, когда считаешь «дайджест за сегодня».

</Predict>

::: interview
<details><summary>Как River повторяет упавшие задачи и как это настроить?</summary>

Ошибка из `Work` записывается в `errors`, задача планируется на повтор: по умолчанию через `n^4` секунд с джиттером 10%, где `n` номер ошибки. Когда `attempt` достиг `MaxAttempts` (по умолчанию 25, это около 20 дней повторов), задача уходит в `discarded`. Настраивается `MaxAttempts` (в `InsertOpts` у аргументов или при вставке), политика клиента `Config.RetryPolicy` и метод `NextRetry` у воркера. Постоянные ошибки отменяют через `JobCancel`.

</details>
<details><summary>Чем JobSnooze отличается от возврата ошибки?</summary>

Ошибка тратит попытку, пишется в `errors`, планирует повтор по backoff и в итоге может отбросить задачу. Snooze означает «рано, попробуй через d»: задача уходит в `scheduled`, попытка не засчитывается, ошибок нет. Используется при rate limit и ожидании внешнего процесса. Верхнюю границу ожидания приходится контролировать самому.

</details>
<details><summary>Что гарантируют уникальные задачи River и чего не гарантируют?</summary>

Гарантируют, что не вставится вторая задача с тем же ключом (Kind плюс выбранные измерения: аргументы или помеченные поля, период в UTC, очередь), пока первая в одном из состояний `ByState` (по умолчанию все, кроме cancelled и discarded). Реализовано частичным уникальным индексом по `unique_key`. Не гарантируют однократного выполнения: River at-least-once, поэтому воркер всё равно должен быть идемпотентным. После очистки завершённых задач (24 часа) и после отмены та же задача вставится снова.

</details>
<details><summary>Как в River устроены периодические задачи и почему они могут пропуститься?</summary>

Их ставит лидер по расписанию из `Config.PeriodicJobs`. Состояние расписания в памяти лидера, поэтому рестарт или смена лидера сбрасывает отсчёт, и запуск может пропуститься или сдвинуться. Страховки: `RunOnStart` (ставить при старте лидера), уникальность `ByPeriod` (не больше одной за период), одинаковая конфигурация на всех клиентах. Для гарантированного расписания есть платный вариант с сохранением состояния.

</details>
<details><summary>Как протестировать воркер River?</summary>

`rivertest.NewWorker(t, driver, config, worker)` и `Work(ctx, t, tx, args, opts)`: задача вставляется в транзакцию теста и выполняется синхронно, возвращается `EventKind` и строка задачи после выполнения (состояние, попытки, время повтора). Зависимости воркера подменяются fake-ами. Вставку проверяют `RequireInserted`/`RequireInsertedTx`/`RequireNotInserted`. Сквозной тест с клиентом ждёт события через `Subscribe` с таймаутом, без `time.Sleep`.

</details>
<details><summary>Как сделать идемпотентный воркер, который списывает деньги?</summary>

Три слоя. Проверка «уже сделано» в своей базе перед вызовом (экономит вызов при повторной доставке). Ключ идемпотентности у платёжного шлюза, зависящий только от платежа (защищает от падения между списанием и записью и от потерянного ответа). Запись результата с `ON CONFLICT DO NOTHING` по уникальному ключу (защищает от двух копий задачи одновременно). Плюс уникальность задачи при вставке, чтобы не плодить копии.

</details>
<details><summary>Зачем разные очереди, если есть приоритеты?</summary>

Приоритет работает внутри очереди и при постоянном потоке высокоприоритетных задач может навсегда задержать низкие. Очереди дают изоляцию: свой пул воркеров, своё `MaxWorkers`, паузу при инциденте. Медленные отчёты не съедают воркеры у писем о сбросе пароля. Внешние лимиты (почтовый провайдер, партнёрский API) естественно задаются через `MaxWorkers` отдельной очереди.

</details>
<details><summary>Чем JobTimeout отличается от RescueStuckJobsAfter?</summary>

`JobTimeout` (по умолчанию 1 минута, у воркера метод `Timeout`) отменяет `ctx` выполняющейся задачи: живой процесс сам прерывает работу. `RescueStuckJobsAfter` (по умолчанию час) про мёртвые процессы: лидер находит задачи, которые висят в `running` дольше этого срока, и возвращает их в работу или отбрасывает. Второе должно быть больше первого.

</details>
:::

## Источники

- `RU` [Очередь задач на Postgres: SKIP LOCKED + lease/heartbeat + backpressure](https://habr.com/ru/articles/984102/): как те же задачи (аренда, возврат, давление) решают в самописной очереди; хорошо видно, что River делает за тебя.
- `EN` [River: Job retries](https://riverqueue.com/docs/job-retries): политика повторов, `NextRetry`, `MaxAttempts`.
- `EN` [River: Snoozing jobs](https://riverqueue.com/docs/snoozing-jobs) и [Cancelling jobs](https://riverqueue.com/docs/cancelling-jobs): snooze и отмена изнутри и снаружи.
- `EN` [River: Unique jobs](https://riverqueue.com/docs/unique-jobs): измерения уникальности, `ByState`, что гарантируется.
- `EN` [River: Periodic and cron jobs](https://riverqueue.com/docs/periodic-jobs): расписания, `RunOnStart`, связь с уникальностью.
- `EN` [River: Testing](https://riverqueue.com/docs/testing) и [пакет rivertest](https://pkg.go.dev/github.com/riverqueue/river/rivertest): `RequireInserted`, `rivertest.Worker`.
- `EN` [Running the River web UI](https://riverqueue.com/docs/river-ui): запуск, встраивание, basic auth.

## Домашка

Пакет `homework/m09/l03-river-advanced/`. Перед запуском `cd homework && docker compose up -d`.

<Hw id="ex1" title="Свой backoff и NextRetry" level="база">

`ex1_backoff.go`: чистая функция `Backoff` (экспонента, потолок, равный джиттер, без переполнения на 1000 ошибках) и `ReportWorker.NextRetry`. Тест через `rivertest.Worker` роняет задачу трижды и проверяет время повторов и `discarded` при `MaxAttempts 3`.

```bash
cd homework && go test -v ./m09/l03-river-advanced/ -run 'TestBackoff|TestReportWorker'
```

</Hw>

<Hw id="ex2" title="Snooze, отмена, таймаут" level="база">

`ex2_snooze.go`, `ExportWorker.Work` и `Timeout`: «готовится» засыпает на 30 секунд, «провалилась» и неизвестный статус отменяют, ошибка API идёт в обычный повтор. Отдельный тест проверяет, что snooze не отбрасывает задачу даже при `MaxAttempts 1`.

```bash
cd homework && go test -v ./m09/l03-river-advanced/ -run TestExportWorker
```

</Hw>

<Hw id="ex3" title="Один дайджест в сутки" level="посложнее">

`ex3_unique.go`: тег `river:"unique"`, `InsertOpts` с уникальностью и `EnqueueDigest`. Тест: другой инициатор в те же сутки это дубль, другой пользователь и следующие сутки UTC нет, отменённая задача не блокирует новую.

```bash
cd homework && go test -v ./m09/l03-river-advanced/ -run TestEnqueueDigest
```

</Hw>

<Hw id="ex4" title="Расписание «каждый день в 09:00»" level="посложнее">

`ex4_schedule.go`, `DailyAt.Next`: ближайший момент строго после текущего по часам заданной зоны, правильно через переход на летнее и зимнее время (сутки 23 и 25 часов) и когда `current` пришёл в UTC.

```bash
cd homework && go test -v ./m09/l03-river-advanced/ -run TestDailyAt
```

</Hw>

<Hw id="ex5" title="Списание денег без двойного списания" level="со звёздочкой">

`ex5_charge.go`, `ChargeWorker.Work`: проверка в своей базе, стабильный ключ идемпотентности для шлюза, запись с `ON CONFLICT`. Тесты: повторная доставка, потерянный ответ шлюза, две копии задачи одновременно (синхронизация через каналы fake-шлюза).

```bash
cd homework && go test -race -v ./m09/l03-river-advanced/ -run TestChargeWorker
```

</Hw>

## Проект модуля: уведомления и дайджест

Продолжаем «Трекер задач» из модуля 08 (`projects/tasktracker/`). У задачи появляются исполнитель и срок. Когда задачу назначают на человека, ему уходит письмо. Каждый день исполнители получают одно письмо со списком своих просроченных задач. Всё через River, письмо ставится в той же транзакции, что и изменение задачи.

### Что уже готово в каркасе

- River добавлен в модуль (`go get github.com/riverqueue/river@v0.51.0 github.com/riverqueue/river/riverdriver/riverpgxv5@v0.51.0`).
- Миграция `00002_assignee_due.sql`: колонки `tasks.assignee_id` и `tasks.due_at`, индекс для дайджеста, таблица `sent_emails` (журнал отправленных писем по ключу идемпотентности). `migrations.RiverUp` создаёт таблицы River на старте.
- `internal/jobs`: аргументы `TaskAssignedArgs` (очередь `email`, 10 попыток) и `OverdueDigestArgs`, периодическая задача `DigestPeriodicJob` (раз в сутки, `RunOnStart`, уникальность на сутки), интерфейс `Mailer` с `LogMailer` (пишет письма в лог) и `FakeMailer` (для тестов, дедуплицирует по ключу), `NewClient` с очередями `default` и `email`.
- `repository.Jobs`: реализация `service.JobQueue` через `InsertTx` в транзакции `TxManager.InTx`; `TxManager.WithRiver` подключает клиент. Образец теста `TestJobsInTx`: коммит оставляет задачу, откат не оставляет.
- `cmd/api/main.go`: миграции River, запуск клиента, остановка River после HTTP-сервера и до закрытия пула.
- Заготовки с подробными TODO: `Tasks.Assign` (репозиторий), `TaskService.Assign` (сервис), `TaskAssignedWorker.Work`, `OverdueDigestWorker.Work`.

Сейчас воркеры возвращают `jobs.ErrNotImplemented`: запусти сервис, и в `river_job` появится дайджест в состоянии `retryable` с этой ошибкой. Дайджест (25 попыток, это около 20 дней повторов) дождётся твоей реализации. А вот у `task_assigned` 10 попыток: сумма задержек 1^4..9^4 секунд около 4.3 часа, после чего задача уйдёт в `discarded`. Такие задачи можно вернуть в работу вручную: `client.JobRetry(ctx, id)` или кнопкой в River UI.

### Этапы

0. **Перенос каркаса.** Каркас m09 меняет файлы, которые ты уже заполнял в модуле 08. Что изменилось:

   | Файл | Изменение |
   |---|---|
   | `go.mod`, `go.sum` | зависимости River v0.51.0 |
   | `migrations/00002_assignee_due.sql`, `migrations/river.go` | новые файлы |
   | `internal/jobs/*` | новый пакет |
   | `internal/service/service.go` | поля `Task.AssigneeID`, `Task.DueAt`; поле `Store.Jobs`; интерфейс `JobQueue` |
   | `internal/service/tasks.go` | в `TaskRepo` новый метод `Assign(ctx, ownerID, id int64, assigneeID *int64, dueAt *time.Time) (Task, error)`; заготовка `TaskService.Assign`; импорт `time` |
   | `internal/repository/repository.go` | поле `TxManager.river`, метод `WithRiver`, `Jobs` в `InTx`, тип `Jobs`, заготовка `Tasks.Assign`, импорты `river` и `jobs` |
   | `internal/repository/jobs_test.go` | новый тест-образец |
   | `cmd/api/main.go` | `migrations.RiverUp`, `jobs.NewClient`, `Start`, остановка River, `NewTxManager(pool).WithRiver(riverClient)` |

   Интерфейсы HTTP-слоя (`internal/http/router.go`) не менялись. Если каркас уже содержал эти изменения, когда ты делал модуль 08, переносить нечего, проверь только fake-и (ниже).

   Как перенести поверх своей работы:

   ```bash
   cd ~/Desktop/globus/go-school
   git add -A && git commit -m "трекер: модуль 08"          # сначала зафиксируй свою работу
   git log --oneline -- projects/tasktracker                  # найди коммит каркаса m09 (в сообщении m09)
   git show <хеш-m09> -- projects/tasktracker > /tmp/tt-m09.patch
   git apply --3way /tmp/tt-m09.patch                         # конфликты только в файлах из таблицы
   ```

   В конфликтах оставляй свою реализацию из модуля 08 и добавляй блоки, помеченные «модуль 09». Новых TODO всего три: `Tasks.Assign`, `TaskService.Assign` и два воркера.

   **Fake-репозитории.** Твой fake `TaskRepo` из тестов сервиса (m08/l07) перестанет компилироваться: в интерфейсе появился `Assign`. До этапа 2 хватит заглушки:

   ```go
   func (f *fakeTasks) Assign(ctx context.Context, ownerID, id int64, assigneeID *int64, dueAt *time.Time) (service.Task, error) {
   	return service.Task{}, service.ErrNotImplemented
   }
   ```

   Или встрой интерфейс в fake (`type fakeTasks struct { service.TaskRepo; ... }`, как `fakeAuth` в `router_test.go`): невызванные методы реализовывать не нужно, вызов нереализованного упадёт с nil pointer, и тест сразу покажет, что fake неполон. Проверка этапа: `go build ./... && go vet ./... && go test ./...` зелёные, как в конце модуля 08.

1. **Репозиторий.** `Tasks.Assign`: один `UPDATE ... FROM projects` с проверкой владельца, `RETURNING` с новыми полями. Добавь `assignee_id` и `due_at` в остальные запросы задач (`Get`, `List`, `Update`). В `mapErr` переведи 23503 (нет такого пользователя) в ошибку валидации поля `assignee_id`. Тесты в базе: чужая задача даёт `ErrNotFound`, несуществующий исполнитель даёт ошибку валидации, `nil` снимает исполнителя.
2. **Сервис.** `TaskService.Assign` по TODO: всё в `s.tx.InTx`, письмо ставится только если исполнитель появился или сменился. Юнит-тесты с fake `TxRunner` и fake `JobQueue`: назначение ставит одну задачу с правильными id, повторное назначение того же человека и снятие не ставят, ошибка `Jobs` возвращается наружу.
3. **HTTP.** `PUT /v1/tasks/:id/assignee` с телом `{"assignee_id": 2, "due_at": "2026-11-20T18:00:00Z"}`, оба поля можно передать `null`. Ответ 200 с задачей. Чужая задача 404, несуществующий исполнитель и плохая дата 422 `validation_failed` с `fields`. Добавь метод в интерфейс `TaskService` в `internal/http/router.go` и тесты через `httptest`.
4. **Письмо о назначении.** `TaskAssignedWorker.Work` по TODO в `internal/jobs/workers.go`. Тесты через `rivertest.NewWorker` и `FakeMailer`: письмо уходит один раз при повторной доставке, задача отменяется, если задачи или пользователя нет, письмо не уходит, если задачу переназначили, ошибка почты ведёт к повтору без записи в `sent_emails`.

   Повторная доставка это **та же** задача, второй раз. `tw.Work` каждый раз вставляет новую задачу с новым id, а ключ идемпотентности строится по `job.ID`, поэтому два `Work` подряд дадут два письма, и это правильно: это две разные задачи (два назначения). Повтор той же задачи делается через `WorkJob`:

   ```go
   res, err := tw.Work(ctx, t, tx, jobs.TaskAssignedArgs{TaskID: taskID, AssigneeID: bobID}, nil)
   // ... проверки первого запуска: job_completed, одно письмо
   res, err = tw.WorkJob(ctx, t, tx, res.Job) // та же строка river_job, тот же job.ID
   // want: job_completed и по-прежнему одно письмо в FakeMailer
   ```

   `WorkJob` переводит строку в `running` из любого состояния и выполняет её снова, как rescuer или повтор после сбоя.
5. **Дайджест.** `OverdueDigestWorker.Work`: одно письмо на исполнителя со списком его просроченных незакрытых задач, ключ на человека и день, ошибка одному получателю не мешает остальным. Тест: три исполнителя, у одного почта падает; повторный запуск задачи досылает только ему.
6. **Атомарность целиком.** Интеграционный тест в базе (пакет `repository`, хелпер `newTestPool` и `migrations.RiverUp`). Назначение на несуществующего пользователя тут не годится: внешний ключ падает на шаге `Tasks.Assign`, ДО постановки письма, и тест прошёл бы даже с `Insert` вместо `InsertTx`. Ошибка должна случиться ПОСЛЕ `st.Jobs.TaskAssigned`. Для этого оберни настоящий `TxManager`:

   ```go
   var errBoom = errors.New("упали после постановки письма")

   // failAfter выполняет fn в настоящей транзакции и потом возвращает ошибку:
   // всё, что fn успела сделать (Tasks.Assign, Jobs.TaskAssigned), обязано откатиться.
   type failAfter struct{ service.TxRunner }

   func (f failAfter) InTx(ctx context.Context, fn func(service.Store) error) error {
   	return f.TxRunner.InTx(ctx, func(st service.Store) error {
   		if err := fn(st); err != nil {
   			return err
   		}
   		return errBoom
   	})
   }
   ```

   Тест: пользователи alice и bob, проект и задача alice. `TaskService` с `failAfter{NewTxManager(pool).WithRiver(client)}`: `Assign` на bob возвращает ошибку, для которой `errors.Is(err, errBoom)`; `rivertest.RequireNotInserted(..., &jobs.TaskAssignedArgs{}, nil)`; `assignee_id` задачи в базе по-прежнему `NULL`. Контрольный случай с тем же сервисом без `failAfter`: задача River есть (`RequireInserted`), иначе тест проверял бы пустоту. Проверь сам, что тест ловит ошибку: временно замени в `repository.Jobs` вызов `InsertTx(ctx, j.tx, ...)` на `Insert(ctx, ...)`, тест должен упасть.

**Со звёздочкой**

- Расписание дайджеста «каждый день в 09:00 по Москве» через `DailyAt` из домашки вместо `PeriodicInterval`.
- Fan-out: при тысячах исполнителей `OverdueDigestWorker` не шлёт письма сам, а ставит по задаче на человека через `InsertManyTx` с уникальностью по исполнителю и дню.
- Отдельный процесс `cmd/worker`: тот же `jobs.NewClient`, а `cmd/api` использует клиент только для вставки (без `Queues`).
- Метрики: подписка на события клиента (`Subscribe`) и счётчики выполненных и упавших задач по Kind в логах.
- River UI в docker рядом с сервисом с basic auth.

### Чеклист приёмки

- [ ] `go build ./...`, `go vet ./...`, `go test -race ./...` в `projects/tasktracker` зелёные.
- [ ] Назначение задачи ставит ровно одну задачу `task_assigned` в очередь `email`; повторное назначение того же человека и снятие исполнителя не ставят.
- [ ] Ошибка после постановки письма откатывает и изменение задачи, и задачу River (тест с `failAfter`, падает при `Insert` вместо `InsertTx`).
- [ ] Письмо о назначении не дублируется при повторной доставке той же задачи (`tw.Work`, затем `tw.WorkJob` с тем же `res.Job`: одно письмо).
- [ ] Удалённая задача или пользователь отменяют задачу River (`cancelled`), а не гоняют её 10 попыток.
- [ ] Дайджест приходит одним письмом на человека, не чаще раза в сутки, повтор после частичной ошибки досылает только недошедшие.
- [ ] Тесты не используют `time.Sleep` для ожидания задач.
- [ ] Ctrl+C: в логах «HTTP остановлен», затем `River client stopped` и «остановлен», процесс выходит.
- [ ] В аргументах задач только id, без email и текстов писем.

### Проверка руками

Запуск как в модуле 08 (`make db-create`, `make run`). Пользователи, токены, проект и задача создаются командами из раздела «Проверка руками» [урока m08/l07](/m08/l07-testing-service), ниже продолжение с переменными `$API`, `$A` и `$T` оттуда.

```bash
# psql без TTY, чтобы подставлять результат в переменные
PSQL="docker compose -f ../../homework/docker-compose.yml exec -T postgres psql -U goschool tasktracker"

# id второго пользователя (bob)
BOB=$($PSQL -tAc "SELECT id FROM users WHERE email = 'bob@example.com'")

# назначить задачу alice на bob: 200, в логе сервиса запись «письмо» с to=bob@example.com
curl -s -X PUT $API/tasks/$T/assignee -H "Authorization: Bearer $A" \
  -H 'Content-Type: application/json' -d "{\"assignee_id\": $BOB, \"due_at\": \"2020-01-01T00:00:00Z\"}" | jq

# повторно на того же: 200, нового письма нет
curl -s -o /dev/null -w '%{http_code}\n' -X PUT $API/tasks/$T/assignee -H "Authorization: Bearer $A" \
  -H 'Content-Type: application/json' -d "{\"assignee_id\": $BOB, \"due_at\": \"2020-01-01T00:00:00Z\"}"

# несуществующий исполнитель: 422 validation_failed с fields.assignee_id
curl -s -X PUT $API/tasks/$T/assignee -H "Authorization: Bearer $A" \
  -H 'Content-Type: application/json' -d '{"assignee_id": 999999, "due_at": null}' | jq

# что в очереди и что отправлено
$PSQL -c "SELECT id, kind, state, queue, attempt, args FROM river_job ORDER BY id"
$PSQL -c "SELECT * FROM sent_emails"

# дайджест: срок задачи в прошлом. Перезапусти сервис (RunOnStart ставит дайджест при старте лидера)
# и найди в логе письмо дайджеста для bob. Второй перезапуск в те же сутки UTC новой задачи не ставит:
$PSQL -c "SELECT id, state, created_at FROM river_job WHERE kind = 'overdue_digest'"
```
