package l05auth

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// asUser подменяет Authenticate: кладёт в контекст пользователя из заголовка X-Test-User.
func asUser(c *gin.Context) {
	if id := c.GetHeader("X-Test-User"); id != "" {
		c.Set(userKey, User{ID: id, Role: c.GetHeader("X-Test-Role")})
	}
	c.Next()
}

func taskRouter(store *MemStore) *gin.Engine {
	r := gin.New()
	RegisterTaskRoutes(r.Group("", asUser), store)
	return r
}

func TestTaskRoutesIDOR(t *testing.T) {
	tests := []struct {
		name         string
		method, path string
		user, role   string
		body         string
		wantCode     int
		wantTitle    string // для 200
		wantError    string // для остальных
	}{
		{"своя задача", "GET", "/tasks/1", "alice", "user", "", 200, "купить молоко", ""},
		{"чужая задача: 404, не 403", "GET", "/tasks/2", "alice", "user", "", 404, "", "not found"},
		{"несуществующая задача", "GET", "/tasks/999", "alice", "user", "", 404, "", "not found"},
		{"аноним", "GET", "/tasks/1", "", "", "", 401, "", "unauthorized"},
		{"id не число", "GET", "/tasks/abc", "alice", "user", "", 400, "", "bad id"},
		{"id ноль", "GET", "/tasks/0", "alice", "user", "", 400, "", "bad id"},
		{"admin читает чужую", "GET", "/tasks/2", "root", "admin", "", 200, "секретный план", ""},
		{"правка своей", "PATCH", "/tasks/1", "alice", "user", `{"title":"купить кефир"}`, 200, "купить кефир", ""},
		{"правка чужой", "PATCH", "/tasks/2", "alice", "user", `{"title":"взлом"}`, 404, "", "not found"},
		{"admin не правит чужие", "PATCH", "/tasks/2", "root", "admin", `{"title":"взлом"}`, 404, "", "not found"},
		{"пустой title", "PATCH", "/tasks/1", "alice", "user", `{"title":"   "}`, 400, "", "bad request"},
		{"битый JSON", "PATCH", "/tasks/1", "alice", "user", `{"title":`, 400, "", "bad request"},
		{"аноним правит", "PATCH", "/tasks/1", "", "", `{"title":"x"}`, 401, "", "unauthorized"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewMemStore(
				Task{ID: 1, OwnerID: "alice", Title: "купить молоко"},
				Task{ID: 2, OwnerID: "bob", Title: "секретный план"},
			)
			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			if tt.user != "" {
				req.Header.Set("X-Test-User", tt.user)
				req.Header.Set("X-Test-Role", tt.role)
			}
			w := httptest.NewRecorder()
			taskRouter(store).ServeHTTP(w, req)

			if w.Code != tt.wantCode {
				t.Fatalf("%s %s от %q: код %d, want %d; тело %s", tt.method, tt.path, tt.user, w.Code, tt.wantCode, w.Body)
			}
			if tt.wantCode == 200 {
				var got Task
				if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil || got.Title != tt.wantTitle {
					t.Errorf("тело = %s, want задачу с title %q", w.Body, tt.wantTitle)
				}
			} else {
				var e map[string]string
				if err := json.Unmarshal(w.Body.Bytes(), &e); err != nil || e["error"] != tt.wantError {
					t.Errorf("тело = %s, want {\"error\":%q}", w.Body, tt.wantError)
				}
			}
			if bob, _ := store.Get(2); bob.Title != "секретный план" {
				t.Errorf("задача bob изменилась: %+v", bob)
			}
		})
	}

	t.Run("владельца нельзя сменить через тело", func(t *testing.T) {
		store := NewMemStore(Task{ID: 1, OwnerID: "alice", Title: "a"})
		req := httptest.NewRequest("PATCH", "/tasks/1", strings.NewReader(`{"title":"b","owner_id":"mallory"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Test-User", "alice")
		w := httptest.NewRecorder()
		taskRouter(store).ServeHTTP(w, req)
		if got, _ := store.Get(1); got.OwnerID != "alice" || got.Title != "b" {
			t.Errorf("после PATCH задача = %+v, want владелец alice и title b", got)
		}
	})
}
