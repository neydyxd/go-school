package l03middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/gin-gonic/gin"
)

// Тесты таймаута идут внутри synctest.Test: время фейковое, ожидание мгновенное и точное.

func TestTimeout(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var ctx context.Context
		r := gin.New()
		r.Use(Timeout(2 * time.Second))
		r.GET("/slow", func(c *gin.Context) {
			ctx = c.Request.Context()
			select {
			case <-ctx.Done(): // обработчик честно бросает работу и ничего не пишет
			case <-time.After(time.Minute):
				c.JSON(http.StatusOK, gin.H{"late": true})
			}
		})
		r.GET("/stream", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"part": 1}) // ответ уже ушёл клиенту
			<-c.Request.Context().Done()
		})
		r.GET("/fast", func(c *gin.Context) {
			ctx = c.Request.Context()
			c.JSON(http.StatusOK, gin.H{"ok": true})
		})

		start := time.Now()
		w := serve(r, http.MethodGet, "/slow", nil)
		if took := time.Since(start); took != 2*time.Second {
			t.Fatalf("медленный запрос длился %v, want 2s", took)
		}
		if w.Code != http.StatusGatewayTimeout {
			t.Fatalf("медленный запрос: код %d, want 504", w.Code)
		}
		if got, want := decodeOne(t, w.Body.Bytes()), map[string]any{"error": "timeout"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("медленный запрос: тело %v, want %v", got, want)
		}

		w = serve(r, http.MethodGet, "/stream", nil)
		if w.Code != http.StatusOK {
			t.Fatalf("ответ записан до дедлайна: код %d, want 200", w.Code)
		}
		decodeOne(t, w.Body.Bytes()) // 504 поверх записанного ответа допишет второй JSON

		w = serve(r, http.MethodGet, "/fast", nil)
		if w.Code != http.StatusOK {
			t.Fatalf("быстрый запрос: код %d, want 200 (тело %q)", w.Code, w.Body)
		}
		if _, ok := ctx.Deadline(); !ok {
			t.Fatal("у контекста запроса нет дедлайна")
		}
		if !errors.Is(ctx.Err(), context.Canceled) {
			t.Fatalf("после ответа ctx.Err() = %v, want context.Canceled: забыт cancel()", ctx.Err())
		}
	})
}

func TestBodyLimit(t *testing.T) {
	calls := 0
	r := gin.New()
	r.Use(BodyLimit(16))
	r.POST("/echo", func(c *gin.Context) {
		calls++
		var in struct {
			S string `json:"s"`
		}
		if err := c.ShouldBindJSON(&in); err != nil {
			if mbe := (*http.MaxBytesError)(nil); errors.As(err, &mbe) {
				c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "body too large"})
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"s": in.S})
	})

	tests := []struct {
		name      string
		body      string
		chunked   bool
		status    int
		wantCalls int
	}{
		{"ровно 16 байт", `{"s":"12345678"}`, false, 200, 1},
		{"17 байт с Content-Length", `{"s":"123456789"}`, false, 413, 0},
		{"17 байт без длины (chunked)", `{"s":"123456789"}`, true, 413, 1},
		{"большое без длины", `{"s":"` + strings.Repeat("x", 1<<20) + `"}`, true, 413, 1},
		{"маленькое без длины", `{"s":"ok"}`, true, 200, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls = 0
			req := httptest.NewRequest(http.MethodPost, "/echo", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			if tt.chunked {
				req.ContentLength = -1
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != tt.status {
				t.Fatalf("код %d, want %d (тело %q)", w.Code, tt.status, w.Body)
			}
			decodeOne(t, w.Body.Bytes())
			if calls != tt.wantCalls {
				t.Fatalf("обработчик вызван %d раз, want %d", calls, tt.wantCalls)
			}
		})
	}
}
