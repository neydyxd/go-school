package l05packages

import "testing"

func TestSplitPathVersion(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		wantPrefix string
		wantMajor  int
		wantOK     bool
	}{
		{"pgx v5", "github.com/jackc/pgx/v5", "github.com/jackc/pgx", 5, true},
		{"v2 у короткого пути", "example.com/v2", "example.com", 2, true},
		{"двузначная версия", "example.com/lib/v12", "example.com/lib", 12, true},
		{"без суффикса", "example.com/lib", "example.com/lib", 1, true},
		{"один элемент", "mylib", "mylib", 1, true},
		{"похоже на версию, но нет", "example.com/v2x", "example.com/v2x", 1, true},
		{"просто v", "example.com/v", "example.com/v", 1, true},
		{"слово version", "example.com/version", "example.com/version", 1, true},
		{"суффикс в середине не считается", "example.com/v2/sub", "example.com/v2/sub", 1, true},
		{"пустой путь", "", "", 0, false},
		{"v1 не пишут", "example.com/lib/v1", "", 0, false},
		{"v0 не пишут", "example.com/lib/v0", "", 0, false},
		{"ведущий ноль", "example.com/lib/v02", "", 0, false},
		{"нет пути перед суффиксом", "/v2", "", 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, m, ok := SplitPathVersion(tt.path)
			if p != tt.wantPrefix || m != tt.wantMajor || ok != tt.wantOK {
				t.Errorf("SplitPathVersion(%q) = %q, %d, %v, want %q, %d, %v", tt.path, p, m, ok, tt.wantPrefix, tt.wantMajor, tt.wantOK)
			}
		})
	}
}

func TestCompatibleVersion(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		version string
		want    bool
	}{
		{"v1 без суффикса", "example.com/lib", "v1.4.2", true},
		{"v0 без суффикса", "example.com/lib", "v0.3.0", true},
		{"v2 без суффикса нельзя", "example.com/lib", "v2.0.0", false},
		{"v5 с /v5", "github.com/jackc/pgx/v5", "v5.7.1", true},
		{"пререлиз тоже версия", "github.com/jackc/pgx/v5", "v5.0.0-rc.1", true},
		{"v4 с /v5 нельзя", "github.com/jackc/pgx/v5", "v4.18.0", false},
		{"v1 с /v2 нельзя", "example.com/lib/v2", "v1.0.0", false},
		{"v10 не v1", "example.com/lib", "v10.0.0", false},
		{"версия без v", "example.com/lib", "1.0.0", false},
		{"версия без точки", "example.com/lib", "v1", false},
		{"мусор вместо версии", "example.com/lib", "vX.1.0", false},
		{"кривой путь", "example.com/lib/v1", "v1.0.0", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CompatibleVersion(tt.path, tt.version); got != tt.want {
				t.Errorf("CompatibleVersion(%q, %q) = %v, want %v", tt.path, tt.version, got, tt.want)
			}
		})
	}
}
