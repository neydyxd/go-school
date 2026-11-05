# syntax=docker/dockerfile:1
# Сборка из папки homework:
#   docker build -f m10/l01-docker/Dockerfile -t goschool-l01:dev .

# ---- стадия сборки ----
FROM golang:1.27-alpine AS build
WORKDIR /src

# Зависимости отдельным слоем: пересобираются, только когда меняется go.mod/go.sum.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Кеш компилятора переживает пересборки (BuildKit cache mount).
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./m10/l01-docker/cmd/server

# ---- финальный образ ----
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/server /server
USER nonroot:nonroot
EXPOSE 8080
HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
  CMD ["/server", "-healthcheck"]
ENTRYPOINT ["/server"]
