# syntax=docker/dockerfile:1

# ---- стадия сборки ----
FROM golang:1.27-alpine AS build
WORKDIR /src

# Сначала только go.mod и go.sum: слой с зависимостями
# пересоберётся, только когда они изменятся.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/app ./cmd/api

# ---- финальный образ: без shell, пользователь nonroot ----
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/app /app
USER nonroot:nonroot
EXPOSE 8080
HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
  CMD ["/app", "-healthcheck"]
ENTRYPOINT ["/app"]
