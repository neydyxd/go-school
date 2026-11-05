package http

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"

	"goschool/tasktracker/internal/auth"
)

type loggerKey struct{}

var validRequestID = regexp.MustCompile(`^[A-Za-z0-9-]{1,64}$`)

// requestLogger: request_id, логгер запроса в контексте и итоговая строка лога. Готово.
func requestLogger(base *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		id := c.GetHeader("X-Request-ID")
		if !validRequestID.MatchString(id) {
			b := make([]byte, 16)
			_, _ = rand.Read(b)
			id = hex.EncodeToString(b)
		}
		c.Header("X-Request-ID", id)
		l := base.With("request_id", id)
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), loggerKey{}, l))
		c.Next()

		level := slog.LevelInfo
		if c.Writer.Status() >= 500 {
			level = slog.LevelError
		}
		l.Log(c.Request.Context(), level, "http request",
			"method", c.Request.Method, "path", c.Request.URL.Path, // без query: там бывают токены
			"status", c.Writer.Status(), "duration_ms", time.Since(start).Milliseconds())
	}
}

func loggerFrom(c *gin.Context) *slog.Logger {
	if l, ok := c.Request.Context().Value(loggerKey{}).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}

const claimsKey = "auth.claims"

// authenticate проверяет "Authorization: Bearer <access>" через tokens.ParseAccess
// и кладёт auth.Claims под claimsKey. Нет токена или он невалиден: 401
// fail(c, 401, "unauthorized", ...) и заголовок WWW-Authenticate: Bearer.
//
// TODO: реализуй (образец: домашка m08/l05, ex3). Пока все защищённые маршруты отвечают 501.
func authenticate(tokens TokenParser) gin.HandlerFunc {
	return func(c *gin.Context) {
		fail(c, http.StatusNotImplemented, "not_implemented", "authenticate ещё не реализован", nil)
	}
}

// currentUser достаёт claims, которые положил authenticate. Готово.
func currentUser(c *gin.Context) (auth.Claims, bool) {
	v, ok := c.Get(claimsKey)
	if !ok {
		return auth.Claims{}, false
	}
	cl, ok := v.(auth.Claims)
	return cl, ok
}
