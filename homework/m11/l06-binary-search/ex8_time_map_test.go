package l06binarysearch

import (
	"strconv"
	"testing"
)

func TestTimeMap(t *testing.T) {
	m := NewTimeMap()
	if v, ok := m.Get("нет такого", 10); ok || v != "" {
		t.Fatalf("Get(\"нет такого\", 10) = %q, %v, want \"\", false", v, ok)
	}
	m.Set("лимит", "100", 10)
	m.Set("лимит", "200", 20)
	m.Set("лимит", "300", 30)
	m.Set("таймаут", "5s", 15)
	tests := []struct {
		key    string
		ts     int
		want   string
		wantOK bool
	}{
		{"лимит", 5, "", false}, // раньше первой версии
		{"лимит", 10, "100", true},
		{"лимит", 15, "100", true},
		{"лимит", 20, "200", true},
		{"лимит", 29, "200", true},
		{"лимит", 30, "300", true},
		{"лимит", 1000, "300", true},
		{"таймаут", 14, "", false},
		{"таймаут", 15, "5s", true},
		{"таймаут", 100, "5s", true},
		{"ретраи", 100, "", false},
	}
	for _, tt := range tests {
		if v, ok := m.Get(tt.key, tt.ts); v != tt.want || ok != tt.wantOK {
			t.Errorf("Get(%q, %d) = %q, %v, want %q, %v", tt.key, tt.ts, v, ok, tt.want, tt.wantOK)
		}
	}
	// пустое значение это тоже значение
	m.Set("флаг", "", 1)
	if v, ok := m.Get("флаг", 2); !ok || v != "" {
		t.Errorf("Get(\"флаг\", 2) = %q, %v, want \"\", true", v, ok)
	}
}

func TestTimeMapBig(t *testing.T) {
	const n = 200_000
	m := NewTimeMap()
	failed, ts, got := false, 0, ""
	within(t, func() {
		for i := range n {
			m.Set("k", strconv.Itoa(i), 10*i)
		}
		for i := range n {
			q := 10*i + 5 // между версиями i и i+1
			if g, _ := m.Get("k", q); g != strconv.Itoa(i) {
				failed, ts, got = true, q, g
				return
			}
		}
	})
	if failed {
		t.Fatalf("Get(\"k\", %d) = %q на большом входе, want %q", ts, got, strconv.Itoa((ts-5)/10))
	}
}
