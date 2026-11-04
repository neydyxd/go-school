package l08structs

import (
	"encoding/json"
	"testing"
	"time"
)

var (
	t0 = time.Date(2026, 11, 4, 10, 0, 0, 0, time.UTC)
	t1 = t0.Add(time.Hour)
)

func TestNewArticle(t *testing.T) {
	a := NewArticle(1, "Слайсы", t0)
	if a.ID != 1 || a.Title != "Слайсы" || !a.CreatedAt.Equal(t0) || !a.UpdatedAt.Equal(t0) {
		t.Errorf("NewArticle(1, \"Слайсы\", t0) = %+v, want ID 1, Title Слайсы, CreatedAt = UpdatedAt = t0", a)
	}
}

func TestArticleRename(t *testing.T) {
	tests := []struct {
		name        string
		title       string
		wantOK      bool
		wantTitle   string
		wantUpdated time.Time
	}{
		{"новый заголовок", "Мапы", true, "Мапы", t1},
		{"пустой заголовок", "", false, "Слайсы", t0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := NewArticle(1, "Слайсы", t0)
			ok := a.Rename(tt.title, t1)
			if ok != tt.wantOK || a.Title != tt.wantTitle || !a.UpdatedAt.Equal(tt.wantUpdated) {
				t.Errorf("Rename(%q) = %v, Title %q, UpdatedAt %v; want %v, %q, %v",
					tt.title, ok, a.Title, a.UpdatedAt, tt.wantOK, tt.wantTitle, tt.wantUpdated)
			}
			if !a.CreatedAt.Equal(t0) {
				t.Errorf("Rename изменил CreatedAt: %v, want %v", a.CreatedAt, t0)
			}
		})
	}
}

func TestArticlePromotedTouch(t *testing.T) {
	a := NewArticle(2, "x", t0)
	a.Touch(t1) // метод встроенной структуры вызывается напрямую
	if !a.UpdatedAt.Equal(t1) {
		t.Errorf("a.Touch(t1): UpdatedAt = %v, want %v", a.UpdatedAt, t1)
	}
}

func TestArticleJSONIsFlat(t *testing.T) {
	b, err := json.Marshal(NewArticle(3, "JSON", t0))
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	want := `{"id":3,"title":"JSON","created_at":"2026-11-04T10:00:00Z","updated_at":"2026-11-04T10:00:00Z"}`
	if string(b) != want {
		t.Errorf("json.Marshal(статья) = %s, want %s", b, want)
	}
}
