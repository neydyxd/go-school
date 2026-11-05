// Package http: транспортный слой. Разбирает запрос, зовёт сервис, переводит
// результат и ошибки в HTTP. Бизнес-правил здесь нет.
package http

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"goschool/tasktracker/internal/auth"
	"goschool/tasktracker/internal/service"
)

// Интерфейсы того, что нужно хендлерам от сервисов. Объявлены здесь, у потребителя:
// в тестах хендлеров вместо настоящих сервисов подставляются fake.

type AuthService interface {
	Register(ctx context.Context, email, password string) (service.User, error)
	Login(ctx context.Context, email, password string) (service.TokenPair, error)
	Refresh(ctx context.Context, refreshToken string) (service.TokenPair, error)
	Logout(ctx context.Context, userID int64) error
}

type ProjectService interface {
	Create(ctx context.Context, ownerID int64, name string) (service.Project, error)
	Get(ctx context.Context, ownerID, id int64) (service.Project, error)
	List(ctx context.Context, ownerID int64) ([]service.Project, error)
	Delete(ctx context.Context, ownerID, id int64) error
}

type TaskService interface {
	Create(ctx context.Context, ownerID, projectID int64, title string) (service.Task, error)
	List(ctx context.Context, ownerID, projectID int64, f service.TaskFilter) ([]service.Task, error)
	Get(ctx context.Context, ownerID, id int64) (service.Task, error)
	Update(ctx context.Context, ownerID, id int64, title, status *string) (service.Task, error)
	Delete(ctx context.Context, ownerID, id int64) error
}

type TokenParser interface {
	ParseAccess(token string) (auth.Claims, error)
}

// Pinger: *pgxpool.Pool подходит без обёрток.
type Pinger interface {
	Ping(ctx context.Context) error
}

type Deps struct {
	Log      *slog.Logger
	DB       Pinger
	Tokens   TokenParser
	Auth     AuthService
	Projects ProjectService
	Tasks    TaskService
}

func NewRouter(d Deps) *gin.Engine {
	useJSONNames()
	r := gin.New()
	_ = r.SetTrustedProxies(nil) // ClientIP не верит X-Forwarded-For от кого попало
	r.Use(requestLogger(d.Log), gin.CustomRecovery(func(c *gin.Context, _ any) {
		writeError(c, errPanic)
	}))

	r.GET("/healthz", health(d.DB))

	h := &handlers{d: d}
	v1 := r.Group("/v1")
	v1.POST("/auth/register", h.register)
	v1.POST("/auth/login", h.login)
	v1.POST("/auth/refresh", h.refresh)

	authed := v1.Group("", authenticate(d.Tokens))
	authed.POST("/auth/logout", h.logout)

	authed.GET("/projects", h.listProjects)
	authed.POST("/projects", h.createProject)
	authed.GET("/projects/:id", h.getProject)
	authed.DELETE("/projects/:id", h.deleteProject)

	authed.GET("/projects/:id/tasks", h.listTasks)
	authed.POST("/projects/:id/tasks", h.createTask)
	authed.GET("/tasks/:id", h.getTask)
	authed.PATCH("/tasks/:id", h.updateTask)
	authed.DELETE("/tasks/:id", h.deleteTask)
	return r
}

// health: ГОТОВЫЙ ОБРАЗЕЦ. 200, если база отвечает за секунду, иначе 503.
func health(db Pinger) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), time.Second)
		defer cancel()
		if err := db.Ping(ctx); err != nil {
			loggerFrom(c).Warn("health: база недоступна", "err", err)
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}
