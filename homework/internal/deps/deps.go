//go:build deps

// Package deps держит в go.mod библиотеки, которые нужны урокам модулей 06-09,
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
	_ "github.com/redis/go-redis/v9"
	_ "golang.org/x/crypto/bcrypt"
	_ "golang.org/x/sync/errgroup"
	_ "golang.org/x/time/rate"
)
