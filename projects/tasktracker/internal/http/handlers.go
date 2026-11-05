package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"goschool/tasktracker/internal/service"
)

type handlers struct{ d Deps }

type credentials struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// register: ГОТОВЫЙ ОБРАЗЕЦ хендлера. Разобрать тело (ошибку отдать bindError), позвать сервис,
// ошибку сервиса отдать writeError, успех вернуть с нужным кодом. Пароль и хеш наружу не уходят.
func (h *handlers) register(c *gin.Context) {
	var in credentials
	if err := c.ShouldBindJSON(&in); err != nil {
		bindError(c, err) // битый JSON: 400 bad_request, нет поля: 422 validation_failed
		return
	}
	u, err := h.d.Auth.Register(c.Request.Context(), in.Email, in.Password)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": u.ID, "email": u.Email, "created_at": u.CreatedAt})
}

// Остальные хендлеры: TODO. Пока все отвечают 501 в едином формате ошибок.
// Идентификатор пользователя бери только из currentUser(c), никогда из тела или query.

func (h *handlers) login(c *gin.Context)         { writeError(c, service.ErrNotImplemented) }
func (h *handlers) refresh(c *gin.Context)       { writeError(c, service.ErrNotImplemented) }
func (h *handlers) logout(c *gin.Context)        { writeError(c, service.ErrNotImplemented) }
func (h *handlers) listProjects(c *gin.Context)  { writeError(c, service.ErrNotImplemented) }
func (h *handlers) createProject(c *gin.Context) { writeError(c, service.ErrNotImplemented) }
func (h *handlers) getProject(c *gin.Context)    { writeError(c, service.ErrNotImplemented) }
func (h *handlers) deleteProject(c *gin.Context) { writeError(c, service.ErrNotImplemented) }
func (h *handlers) listTasks(c *gin.Context)     { writeError(c, service.ErrNotImplemented) }
func (h *handlers) createTask(c *gin.Context)    { writeError(c, service.ErrNotImplemented) }
func (h *handlers) getTask(c *gin.Context)       { writeError(c, service.ErrNotImplemented) }
func (h *handlers) updateTask(c *gin.Context)    { writeError(c, service.ErrNotImplemented) }
func (h *handlers) deleteTask(c *gin.Context)    { writeError(c, service.ErrNotImplemented) }
