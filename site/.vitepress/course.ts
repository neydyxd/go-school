// Единый манифест курса: из него строятся сайдбар, схема маршрута и табло урока.
// Урок доступен, если существует файл content/<module.id>/<lesson.slug>.md.

export type Lesson = { id: string; slug: string; title: string }
export type Module = { id: string; title: string; short: string; color: string; lessons: Lesson[] }

type Raw = [id: string, title: string, short: string, color: string, lessons: [slug: string, title: string][]]

// ponytail: цвета линий в OKLCH с близкой светлостью, чтобы ни одна линия не перетягивала внимание
const raw: Raw[] = [
  ['m00', 'Старт', 'Старт', 'oklch(0.62 0.02 250)', [
    ['l01-how-to-learn', 'Как устроен курс и как учиться'],
    ['l02-tools', 'Установка и инструменты Go'],
  ]],
  ['m01', 'Основы языка', 'Основы', 'oklch(0.60 0.20 27)', [
    ['l01-types', 'Переменные, типы, константы'],
    ['l02-control-flow', 'Управляющие конструкции и range'],
    ['l03-functions', 'Функции, замыкания, defer, panic'],
    ['l04-pointers', 'Указатели'],
    ['l05-slices', 'Массивы и слайсы'],
    ['l06-strings', 'Строки, байты, руны'],
    ['l07-maps', 'Мапы'],
    ['l08-structs', 'Структуры и методы'],
  ]],
  ['m02', 'Интерфейсы, ошибки, дженерики', 'Интерфейсы', 'oklch(0.60 0.15 150)', [
    ['l01-interfaces', 'Интерфейсы'],
    ['l02-interfaces-internals', 'Интерфейсы под капотом'],
    ['l03-errors', 'Ошибки'],
    ['l04-generics', 'Дженерики'],
    ['l05-packages', 'Пакеты и модули'],
    ['l06-stdlib', 'Стандартная библиотека, которую надо знать'],
  ]],
  ['m03', 'Тестирование и инструменты', 'Тесты', 'oklch(0.55 0.17 262)', [
    ['l01-testing', 'Пакет testing и табличные тесты'],
    ['l02-bench-fuzz', 'Бенчмарки, фаззинг, покрытие'],
    ['l03-testable-design', 'Тестируемый дизайн, fake и mock'],
    ['l04-quality', 'Линтеры, race detector, pprof'],
  ]],
  ['m04', 'Конкурентность', 'Конкурентность', 'oklch(0.66 0.13 225)', [
    ['l01-goroutines', 'Горутины и WaitGroup'],
    ['l02-channels', 'Каналы'],
    ['l03-select', 'select и таймауты'],
    ['l04-sync', 'Пакет sync и atomic'],
    ['l05-context', 'context'],
    ['l06-patterns', 'Паттерны конкурентности'],
    ['l07-problems', 'Гонки, дедлоки, утечки, модель памяти'],
    ['l08-scheduler', 'Планировщик GMP под капотом'],
  ]],
  ['m05', 'Рантайм под капотом', 'Рантайм', 'oklch(0.52 0.08 55)', [
    ['l01-memory', 'Стек, куча, escape analysis'],
    ['l02-gc', 'Сборщик мусора'],
    ['l03-internals-perf', 'Каналы и мапы изнутри, оптимизация'],
  ]],
  ['m06', 'HTTP на стандартной библиотеке', 'HTTP', 'oklch(0.70 0.17 55)', [
    ['l01-web', 'Как работает веб'],
    ['l02-net-http', 'Сервер на net/http'],
    ['l03-json-api', 'JSON API и валидация'],
    ['l04-http-client', 'http.Client без сюрпризов'],
    ['l05-production-server', 'Сервер к продакшну'],
  ]],
  ['m07', 'Базы данных', 'Базы данных', 'oklch(0.55 0.17 310)', [
    ['l01-postgres-intro', 'PostgreSQL: старт'],
    ['l02-sql-basics', 'SQL: основы'],
    ['l03-sql-advanced', 'SQL: JOIN, CTE, оконные функции'],
    ['l04-schema-design', 'Проектирование схем'],
    ['l05-indexes', 'Индексы и EXPLAIN'],
    ['l06-transactions', 'Транзакции, изоляция, MVCC'],
    ['l07-go-db', 'Go и БД: database/sql и pgx'],
    ['l08-sqlc-orm-migrations', 'sqlc, ORM, миграции'],
    ['l09-redis', 'Redis и кеширование'],
  ]],
  ['m08', 'Gin и архитектура сервиса', 'Gin', 'oklch(0.78 0.15 90)', [
    ['l01-gin-intro', 'Gin: роутинг и первые шаги'],
    ['l02-binding', 'Binding, валидация, ответы'],
    ['l03-middleware', 'Middleware'],
    ['l04-architecture', 'Архитектура сервиса'],
    ['l05-auth', 'Аутентификация и безопасность'],
    ['l06-config-logs-docs', 'Конфиг, логи, OpenAPI'],
    ['l07-testing-service', 'Тестирование сервиса'],
  ]],
  ['m09', 'Фоновые задачи и River', 'River', 'oklch(0.72 0.17 130)', [
    ['l01-why-queues', 'Зачем очереди'],
    ['l02-river-basics', 'River: основы'],
    ['l03-river-advanced', 'River: ретраи, уникальность, периодика'],
    ['l04-brokers', 'Kafka и RabbitMQ обзорно'],
  ]],
  ['m10', 'Продакшн', 'Продакшн', 'oklch(0.58 0.10 195)', [
    ['l01-docker', 'Docker для Go'],
    ['l02-observability', 'Наблюдаемость'],
    ['l03-grpc', 'gRPC и protobuf'],
    ['l04-ci', 'CI на GitHub Actions'],
    ['l05-system-design', 'System design для джуна'],
  ]],
  ['m11', 'Алгоритмы и структуры данных', 'Алгоритмы', 'oklch(0.62 0.20 350)', [
    ['l01-big-o', 'Big-O и как решать задачу вслух'],
    ['l02-two-pointers', 'Два указателя и префиксные суммы'],
    ['l03-hashing', 'Хеш-таблицы'],
    ['l04-sliding-window', 'Скользящее окно'],
    ['l05-stack-queue', 'Стек, очередь, монотонный стек'],
    ['l06-binary-search', 'Бинарный поиск'],
    ['l07-linked-lists', 'Связные списки'],
    ['l08-sorting', 'Сортировки'],
    ['l09-trees', 'Деревья и рекурсия'],
    ['l10-graphs', 'Графы'],
    ['l11-heap', 'Куча и top-K'],
    ['l12-dp-greedy', 'Динамика и жадные алгоритмы'],
  ]],
  ['m12', 'Подготовка к собеседованию', 'Собес', 'oklch(0.40 0.03 260)', [
    ['l01-go-questions', 'Банк вопросов по Go'],
    ['l02-live-coding', 'Что выведет код и live-coding'],
    ['l03-db-net-arch', 'БД, сети, архитектура'],
    ['l04-resume-interview', 'Резюме и поведение на собесе'],
  ]],
]

export const modules: Module[] = raw.map(([id, title, short, color, ls]) => ({
  id, title, short, color,
  lessons: ls.map(([slug, t]) => ({ id: `${id}-${slug.slice(0, 3)}`, slug, title: t })),
}))

export const lessonPath = (m: Module, l: Lesson) => `/${m.id}/${l.slug}`

export const allLessons = modules.flatMap((m) => m.lessons.map((l) => ({ m, l })))

// Цвет текста поверх цвета линии: на светлых линиях (жёлтая, салатовая) тёмный.
export const inkOn = (color: string) => (Number(color.match(/oklch\(([\d.]+)/)?.[1] ?? 0) >= 0.68 ? '#10161d' : '#ffffff')
