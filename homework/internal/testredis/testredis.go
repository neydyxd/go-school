// Package testredis даёт тесту клиент учебного Redis и уникальный префикс ключей.
//
// Redis поднимается из папки homework: docker compose up -d.
// Адрес можно переопределить переменной GOSCHOOL_REDIS_ADDR.
package testredis

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// New возвращает клиента и префикс вида "t:<случайное>:". Тест обязан класть ключи
// только под этим префиксом: так параллельные пакеты не мешают друг другу.
// После теста все ключи с префиксом удаляются.
func New(t testing.TB) (*redis.Client, string) {
	t.Helper()
	addr := os.Getenv("GOSCHOOL_REDIS_ADDR")
	if addr == "" {
		addr = "localhost:56379"
	}
	c := redis.NewClient(&redis.Options{Addr: addr})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := c.Ping(ctx).Err(); err != nil {
		t.Fatalf("Redis недоступен (%v): подними базу: cd homework && docker compose up -d", err)
	}
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	prefix := "t:" + hex.EncodeToString(b) + ":"
	t.Cleanup(func() {
		ctx := context.Background()
		iter := c.Scan(ctx, 0, prefix+"*", 100).Iterator()
		for iter.Next(ctx) {
			c.Del(ctx, iter.Val())
		}
		_ = c.Close()
	})
	return c, prefix
}
