# Трекер задач (проект модуля 08)

REST API на Gin + PostgreSQL + JWT. Полное ТЗ, чеклист приёмки и curl-команды:
урок «Тестирование сервиса» (m08/l07), раздел «Проект модуля».

```bash
cp .env.example .env   # и впиши JWT_SECRET: openssl rand -base64 48
make db-create   # база tasktracker в контейнере из homework
make run         # http://localhost:8080/healthz
make test
```

Раскладка:

```
cmd/api/              main: конфиг, логгер, пул, миграции, DI, graceful shutdown (готово)
internal/config/      конфиг из окружения с валидацией (готово)
internal/auth/        bcrypt (готово), JWT (TODO), refresh-токены (готово)
internal/service/     бизнес-логика; Register готов, остальное TODO
internal/repository/  pgx через интерфейс DBTX (пул или транзакция), TxManager.InTx (готово);
                      Users.Create готов, остальное TODO
internal/http/        Gin: роутер, единый формат ошибок из m08/l02, логи запросов (готово),
                      register и /healthz (готово), authenticate и остальные хендлеры (TODO)
migrations/           goose SQL, встроены в бинарник через embed
```

Зависимости для TODO добавляются по ходу: `go get github.com/golang-jwt/jwt/v5@v5.3.1`
(JWT), `go get golang.org/x/time@v0.16.0` (rate limit со звёздочкой).

Готовые образцы помечены «ГОТОВЫЙ ОБРАЗЕЦ», заглушки отвечают 501 `not_implemented`.
