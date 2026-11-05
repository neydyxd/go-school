// Package l04architecture собирает сервис библиотеки из слоёв.
package l04architecture

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"goschool/homework/m08/l04-architecture/httpapi"
	"goschool/homework/m08/l04-architecture/postgres"
	"goschool/homework/m08/l04-architecture/service"
)

// NewRouter связывает слои вручную: postgres.NewBookRepo -> service.New -> httpapi.New
// и регистрирует маршруты на группе /api (POST /api/books и так далее) роутера gin.New().
// Никаких глобальных переменных: всё, что нужно слоям, передаётся аргументами.
func NewRouter(pool *pgxpool.Pool) *gin.Engine {
	repo := postgres.NewBookRepo(pool)
	svc := service.New(repo)
	h := httpapi.New(svc)

	r := gin.New()
	h.Register(r.Group("/api"))
	return r
}
