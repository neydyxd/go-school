package l03middleware

import (
	"bufio"
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
)

// capture возвращает логгер в JSON и функцию, которая разбирает записанные строки.
func capture(t *testing.T) (*slog.Logger, func() []map[string]any) {
	var buf bytes.Buffer
	log := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	return log, func() []map[string]any {
		t.Helper()
		var out []map[string]any
		sc := bufio.NewScanner(bytes.NewReader(buf.Bytes()))
		for sc.Scan() {
			var rec map[string]any
			if err := json.Unmarshal(sc.Bytes(), &rec); err != nil {
				t.Fatalf("строка лога не JSON: %q", sc.Text())
			}
			delete(rec, "time")
			out = append(out, rec)
		}
		buf.Reset()
		return out
	}
}

// requestRecord находит единственную запись "request" и проверяет, что duration есть.
func requestRecord(t *testing.T, recs []map[string]any) map[string]any {
	t.Helper()
	var found []map[string]any
	for _, r := range recs {
		if r["msg"] == "request" {
			found = append(found, r)
		}
	}
	if len(found) != 1 {
		t.Fatalf("записей \"request\" %d, want 1; весь лог: %v", len(found), recs)
	}
	rec := found[0]
	if d, ok := rec["duration"].(float64); !ok || d < 0 {
		t.Fatalf("duration = %v, want неотрицательную длительность (slog.Duration)", rec["duration"])
	}
	delete(rec, "duration")
	return rec
}

func TestLogger(t *testing.T) {
	log, read := capture(t)
	r := gin.New()
	r.Use(Logger(log))
	r.GET("/users/:id", func(c *gin.Context) {
		if c.Param("id") == "0" {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"id": c.Param("id")})
	})
	r.GET("/boom", func(c *gin.Context) { c.Status(http.StatusBadGateway) })

	tests := []struct {
		url  string
		want map[string]any
	}{
		{"/users/42", map[string]any{"level": "INFO", "msg": "request", "method": "GET", "route": "/users/:id", "path": "/users/42", "status": 200.0, "request_id": ""}},
		{"/users/0", map[string]any{"level": "WARN", "msg": "request", "method": "GET", "route": "/users/:id", "path": "/users/0", "status": 404.0, "request_id": ""}},
		{"/boom", map[string]any{"level": "ERROR", "msg": "request", "method": "GET", "route": "/boom", "path": "/boom", "status": 502.0, "request_id": ""}},
		{"/nope", map[string]any{"level": "WARN", "msg": "request", "method": "GET", "route": "", "path": "/nope", "status": 404.0, "request_id": ""}},
	}
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			serve(r, http.MethodGet, tt.url, nil)
			if got := requestRecord(t, read()); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("запись лога %v\nwant %v\n(статус и уровень читаются после c.Next())", got, tt.want)
			}
		})
	}
}

func TestRecovery(t *testing.T) {
	log, read := capture(t)
	r := gin.New()
	r.Use(Recovery(log))
	r.GET("/panic", func(c *gin.Context) { panic("boom") })
	r.GET("/ok", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })

	w := serve(r, http.MethodGet, "/panic", nil)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("после паники код %d, want 500", w.Code)
	}
	if got, want := decodeOne(t, w.Body.Bytes()), map[string]any{"error": "internal"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("после паники тело %v, want %v", got, want)
	}
	recs := read()
	if len(recs) != 1 || recs[0]["level"] != "ERROR" || recs[0]["msg"] != "panic" || recs[0]["panic"] != "boom" {
		t.Fatalf("лог после паники %v, want одну запись ERROR \"panic\" с panic=boom", recs)
	}

	w = serve(r, http.MethodGet, "/ok", nil)
	if w.Code != http.StatusOK || len(read()) != 0 {
		t.Fatalf("без паники: код %d, want 200 и пустой лог", w.Code)
	}
}
