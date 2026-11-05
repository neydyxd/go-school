//go:build deps

// Package deps держит в go.mod библиотеки, которые нужны урокам модулей 06-10,
// даже пока ни один пакет домашки их не импортирует. В обычную сборку не попадает.
package deps

import (
	_ "github.com/gin-gonic/gin"
	_ "github.com/go-playground/validator/v10"
	_ "github.com/golang-jwt/jwt/v5"
	_ "github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/pressly/goose/v3"
	_ "github.com/prometheus/client_golang/prometheus"
	_ "github.com/prometheus/client_golang/prometheus/promhttp"
	_ "github.com/redis/go-redis/v9"
	_ "github.com/riverqueue/river"
	_ "github.com/riverqueue/river/riverdriver/riverpgxv5"
	_ "github.com/riverqueue/river/rivermigrate"
	_ "github.com/riverqueue/river/rivertest"
	_ "github.com/riverqueue/river/rivertype"
	_ "golang.org/x/crypto/bcrypt"
	_ "golang.org/x/sync/errgroup"
	_ "golang.org/x/time/rate"
	_ "google.golang.org/grpc"
	_ "google.golang.org/grpc/codes"
	_ "google.golang.org/grpc/status"
	_ "google.golang.org/grpc/test/bufconn"
	_ "google.golang.org/protobuf/proto"
)
