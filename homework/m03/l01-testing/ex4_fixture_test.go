package l01testing

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// recTB оборачивает настоящий t: TempDir, Cleanup и остальное работают как
// обычно, а вызовы Helper и Fatal запоминаются. Fatal не роняет настоящий тест,
// а только останавливает горутину, в которой вызван хелпер (как и настоящий Fatal).
type recTB struct {
	testing.TB
	helper bool
	fatal  string
}

func (r *recTB) Helper() { r.helper = true }
func (r *recTB) Fatalf(format string, a ...any) {
	r.fatal = fmt.Sprintf(format, a...)
	runtime.Goexit()
}
func (r *recTB) Fatal(a ...any) {
	r.fatal = fmt.Sprint(a...)
	runtime.Goexit()
}
func (r *recTB) FailNow() {
	r.fatal = "FailNow()"
	runtime.Goexit()
}

// callFixture вызывает WriteFixture в отдельной горутине, чтобы перехваченный
// Fatal завершил только её.
func callFixture(t *testing.T, name, content string) (*recTB, string) {
	rec := &recTB{TB: t}
	var path string
	done := make(chan struct{})
	go func() {
		defer close(done)
		path = WriteFixture(rec, name, content)
	}()
	<-done
	return rec, path
}

func TestWriteFixture(t *testing.T) {
	var paths []string
	t.Run("файлы создаются", func(t *testing.T) {
		files := []struct{ name, content string }{
			{"config.json", `{"port": 8080}`},
			{"пустой.txt", ""},
			{"данные.csv", "id,имя\n1,Гофер\n"},
		}
		for _, f := range files {
			rec, path := callFixture(t, f.name, f.content)
			if rec.fatal != "" {
				t.Fatalf("WriteFixture(%q) вызвал Fatal: %s", f.name, rec.fatal)
			}
			if !rec.helper {
				t.Errorf("WriteFixture(%q) не вызвал t.Helper()", f.name)
			}
			paths = append(paths, path)
			if !filepath.IsAbs(path) {
				t.Errorf("WriteFixture(%q) = %q, want абсолютный путь", f.name, path)
			}
			if filepath.Base(path) != f.name {
				t.Errorf("WriteFixture(%q) = %q, want файл с именем %q", f.name, path, f.name)
			}
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("WriteFixture(%q): файл не читается: %v", f.name, err)
			}
			if string(data) != f.content {
				t.Errorf("WriteFixture(%q) записал %q, want %q", f.name, data, f.content)
			}
		}
	})
	if len(paths) == 0 {
		t.Fatal("подтест не создал ни одного файла")
	}
	for _, p := range paths {
		if _, err := os.Stat(p); !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("файл %s остался после окончания теста: бери каталог из t.TempDir()", p)
		}
	}
}

func TestWriteFixtureFails(t *testing.T) {
	// Нулевой байт в имени файла запрещён на всех ОС: запись обязана упасть.
	rec, _ := callFixture(t, "плохое\x00имя.txt", "data")
	if rec.fatal == "" {
		t.Error("WriteFixture с невозможным именем файла не вызвал t.Fatalf: ошибку записи нельзя игнорировать")
	}
}
