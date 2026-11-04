package l07maps

import "testing"

func TestLRU(t *testing.T) {
	c := NewLRU(2)
	c.Put("a", 1)
	c.Put("b", 2)
	if v, ok := c.Get("a"); !ok || v != 1 {
		t.Fatalf("Get(\"a\") = %d, %v, want 1, true", v, ok)
	}
	c.Put("c", 3) // вытесняет b: a только что читали
	steps := []struct {
		name   string
		key    string
		want   int
		wantOK bool
	}{
		{"b вытеснен", "b", 0, false},
		{"a остался", "a", 1, true},
		{"c добавлен", "c", 3, true},
	}
	for _, s := range steps {
		t.Run(s.name, func(t *testing.T) {
			if v, ok := c.Get(s.key); v != s.want || ok != s.wantOK {
				t.Errorf("Get(%q) = %d, %v, want %d, %v", s.key, v, ok, s.want, s.wantOK)
			}
		})
	}
	if c.Len() != 2 {
		t.Errorf("Len() = %d, want 2", c.Len())
	}
}

func TestLRUUpdateRefreshes(t *testing.T) {
	c := NewLRU(2)
	c.Put("a", 1)
	c.Put("b", 2)
	c.Put("a", 10) // обновление делает a свежим
	c.Put("c", 3)  // вытесняет b
	if _, ok := c.Get("b"); ok {
		t.Errorf("Get(\"b\") нашёл ключ, want вытеснен: Put существующего ключа должен освежать его")
	}
	if v, ok := c.Get("a"); !ok || v != 10 {
		t.Errorf("Get(\"a\") = %d, %v, want 10, true", v, ok)
	}
	if c.Len() != 2 {
		t.Errorf("Len() = %d, want 2: обновление не должно добавлять дубликат", c.Len())
	}
}

func TestLRUZeroCapacity(t *testing.T) {
	for _, capacity := range []int{0, -1} {
		c := NewLRU(capacity)
		c.Put("a", 1)
		if _, ok := c.Get("a"); ok || c.Len() != 0 {
			t.Errorf("NewLRU(%d): после Put Len() = %d, want 0", capacity, c.Len())
		}
	}
}

func TestLRUManyEvictions(t *testing.T) {
	c := NewLRU(3)
	keys := []string{"k0", "k1", "k2", "k3", "k4", "k5", "k6", "k7", "k8", "k9"}
	for i, k := range keys {
		c.Put(k, i)
	}
	if c.Len() != 3 {
		t.Fatalf("Len() = %d, want 3", c.Len())
	}
	for i, k := range keys {
		_, ok := c.Get(k)
		if want := i >= 7; ok != want {
			t.Errorf("Get(%q) ok = %v, want %v", k, ok, want)
		}
	}
}
