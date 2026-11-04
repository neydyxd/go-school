package testredis

import (
	"context"
	"testing"
)

func TestPrefixIsolation(t *testing.T) {
	c, p := New(t)
	ctx := context.Background()
	if err := c.Set(ctx, p+"k", "v", 0).Err(); err != nil {
		t.Fatal(err)
	}
	if got := c.Get(ctx, p+"k").Val(); got != "v" {
		t.Fatalf("Get = %q, want v", got)
	}
	_, p2 := New(t)
	if p == p2 {
		t.Fatal("префиксы совпали")
	}
}
