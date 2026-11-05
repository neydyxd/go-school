# syntax=docker/dockerfile:1
# Задание ex3 урока «Docker для Go». Собери образ сервиса m10/l01-docker/cmd/server.
# Сборка из папки homework:
#   docker build -f m10/l01-docker/Dockerfile -t goschool-l01:dev .
#
# Требования:
#   1. Multi-stage: стадия сборки на golang:1.27-alpine, финальный образ
#      gcr.io/distroless/static-debian12:nonroot.
#   2. go.mod и go.sum копируются и скачиваются (go mod download) ДО копирования
#      исходников, чтобы правка кода не перекачивала зависимости.
#   3. Статический бинарник: CGO_ENABLED=0, флаги -trimpath -ldflags="-s -w".
#   4. Процесс работает не от root.
#   5. HEALTHCHECK вызывает сам бинарник с флагом -healthcheck (curl в distroless нет).
#   6. ENTRYPOINT в exec-форме (JSON-массив), чтобы SIGTERM доходил до процесса.
#   7. Итоговый образ меньше 30 МБ (docker images goschool-l01).

# TODO
