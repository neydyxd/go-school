package l05productionserver

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
)

func newTestLogger(level slog.Level) (*slog.Logger, *bytes.Buffer) {
	var buf bytes.Buffer
	h := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey && len(groups) == 0 {
				return slog.Attr{} // время мешает сравнению
			}
			return a
		},
	})
	return slog.New(ContextHandler{h}), &buf
}

func lines(t *testing.T, buf *bytes.Buffer) []map[string]any {
	t.Helper()
	var out []map[string]any
	for _, l := range strings.Split(strings.TrimSpace(buf.String()), "\n") {
		if l == "" {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(l), &m); err != nil {
			t.Fatalf("строка лога не JSON: %q", l)
		}
		out = append(out, m)
	}
	return out
}

func TestContextHandler(t *testing.T) {
	ctx := WithRequestID(context.Background(), "req-7")

	t.Run("request_id из контекста", func(t *testing.T) {
		log, buf := newTestLogger(slog.LevelInfo)
		log.InfoContext(ctx, "создан пользователь", "user_id", 42)
		got := lines(t, buf)
		if len(got) != 1 || got[0]["request_id"] != "req-7" || got[0]["user_id"] != float64(42) || got[0]["msg"] != "создан пользователь" {
			t.Errorf("лог %v, want одна запись с request_id=req-7, user_id=42", got)
		}
	})
	t.Run("без ID в контексте атрибута нет", func(t *testing.T) {
		log, buf := newTestLogger(slog.LevelInfo)
		log.InfoContext(context.Background(), "старт")
		log.Info("без контекста")
		for _, m := range lines(t, buf) {
			if _, ok := m["request_id"]; ok {
				t.Errorf("запись %v: request_id есть, а в контексте его не было", m)
			}
		}
	})
	t.Run("With не теряет обёртку", func(t *testing.T) {
		log, buf := newTestLogger(slog.LevelInfo)
		log.With("component", "billing").InfoContext(ctx, "списание")
		got := lines(t, buf)
		if len(got) != 1 || got[0]["request_id"] != "req-7" || got[0]["component"] != "billing" {
			t.Errorf("лог %v, want component=billing и request_id=req-7: WithAttrs должен вернуть ContextHandler", got)
		}
	})
	t.Run("WithGroup не теряет обёртку", func(t *testing.T) {
		log, buf := newTestLogger(slog.LevelInfo)
		log.WithGroup("http").InfoContext(ctx, "запрос", "status", 200)
		got := lines(t, buf)
		if len(got) != 1 {
			t.Fatalf("лог %v, want одна запись", got)
		}
		grp, _ := got[0]["http"].(map[string]any)
		if grp == nil || grp["request_id"] != "req-7" || grp["status"] != float64(200) {
			t.Errorf("лог %v, want группа http со status=200 и request_id=req-7", got)
		}
	})
	t.Run("уровень фильтрует как раньше", func(t *testing.T) {
		log, buf := newTestLogger(slog.LevelWarn)
		log.InfoContext(ctx, "не попадёт")
		log.DebugContext(ctx, "тоже")
		log.WarnContext(ctx, "попадёт")
		got := lines(t, buf)
		if len(got) != 1 || got[0]["level"] != "WARN" {
			t.Errorf("лог %v, want только WARN", got)
		}
	})
}
