package l01testing

import (
	"os"
	"path/filepath"
	"testing"
)

// WriteFixture создаёт файл во временном каталоге теста и возвращает путь к нему.
func WriteFixture(t testing.TB, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFixture(%q): %v", name, err)
	}
	return path
}
