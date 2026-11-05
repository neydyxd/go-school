package l02riverbasics

import (
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
)

// NewClient собирает клиент River для процесса, который выполняет задачи:
//   - river.NewWorkers() и river.AddWorker с &WelcomeEmailWorker{Pool: pool, Mailer: mailer};
//   - очереди (Queues): river.QueueDefault с MaxWorkers 10 и "email" с MaxWorkers 2
//     (почтовый провайдер не любит много параллельных соединений);
//   - SoftStopTimeout 10 секунд: при остановке клиент даёт текущим задачам
//     доработать, а потом отменяет их контекст.
//
// Драйвер: riverpgxv5.New(pool). Клиент не запускай: Start делает вызывающий.
//
// Тест проверяет то, что видно снаружи: воркер зарегистрирован и очередь
// "email" обрабатывается. MaxWorkers и SoftStopTimeout у клиента не прочитать
// (публичного доступа к конфигу нет), их проверь сам: запусти клиент с задачей,
// которая ждёт ctx, вызови Stop и посмотри, через сколько отменится её ctx.
func NewClient(pool *pgxpool.Pool, mailer Mailer) (*river.Client[pgx.Tx], error) {
	workers := river.NewWorkers()
	river.AddWorker(workers, &WelcomeEmailWorker{Pool: pool, Mailer: mailer})
	return river.NewClient(riverpgxv5.New(pool), &river.Config{
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: 10},
			"email":            {MaxWorkers: 2},
		},
		Workers:         workers,
		SoftStopTimeout: 10 * time.Second,
	})
}
