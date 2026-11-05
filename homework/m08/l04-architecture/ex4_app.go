// Package l04architecture собирает сервис библиотеки из слоёв.
package l04architecture

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NewRouter связывает слои вручную: postgres.NewBookRepo -> service.New -> httpapi.New
// и регистрирует маршруты на группе /api (POST /api/books и так далее) роутера gin.New().
// Никаких глобальных переменных: всё, что нужно слоям, передаётся аргументами.
func NewRouter(pool *pgxpool.Pool) *gin.Engine {
	panic("TODO: реализуй NewRouter")
}
