package shortener

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

const base = "https://sho.rt"

func newAPI() (http.Handler, *Store) {
	s := NewStore(seq("abc", "def", "ghi"), fixedNow)
	return NewHandler(s, base), s
}

func do(t *testing.T, h http.Handler, method, path, ctype, body string) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	if ctype != "" {
		r.Header.Set("Content-Type", ctype)
	}
	w := httptest.NewRecorder()
	within(t, method+" "+path, func() { h.ServeHTTP(w, r) })
	return w
}

func errCode(t *testing.T, w *httptest.ResponseRecorder) string {
	t.Helper()
	if ct := w.Result().Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("ошибка с Content-Type %q, want application/json", ct)
	}
	var body struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("тело ошибки не JSON: %q", w.Body.String())
	}
	if body.Error.Message == "" {
		t.Errorf("у ошибки пустой message: %s", w.Body.String())
	}
	return body.Error.Code
}

func TestCreateLink(t *testing.T) {
	h, _ := newAPI()
	w := do(t, h, "POST", "/api/links", "application/json", `{"url":"https://go.dev/doc/"}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("POST /api/links: код %d, тело %s; want 201", w.Code, w.Body.String())
	}
	if ct := w.Result().Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type %q, want application/json", ct)
	}
	if loc := w.Result().Header.Get("Location"); loc != base+"/abc" {
		t.Errorf("Location %q, want %s/abc", loc, base)
	}
	var got map[string]string
	json.Unmarshal(w.Body.Bytes(), &got)
	want := map[string]string{"code": "abc", "short_url": base + "/abc", "url": "https://go.dev/doc/"}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Errorf("тело %v, want %v", got, want)
	}
}

func TestCreateLinkErrors(t *testing.T) {
	tests := []struct {
		name, ctype, body string
		wantStatus        int
		wantCode          string
	}{
		{"нет Content-Type", "", `{"url":"https://go.dev"}`, 415, "unsupported_media_type"},
		{"форма вместо JSON", "application/x-www-form-urlencoded", `url=https://go.dev`, 415, "unsupported_media_type"},
		{"битый JSON", "application/json", `{"url":`, 400, "invalid_json"},
		{"пустое тело", "application/json", ``, 400, "invalid_json"},
		{"лишнее поле", "application/json", `{"url":"https://go.dev","admin":true}`, 400, "invalid_json"},
		{"два объекта", "application/json", `{"url":"https://a.ru"}{"url":"https://b.ru"}`, 400, "invalid_json"},
		{"url не строка", "application/json", `{"url":42}`, 400, "invalid_json"},
		{"тело больше 4 КБ", "application/json", `{"url":"https://a.ru/` + strings.Repeat("x", 5000) + `"}`, 413, "body_too_large"},
		{"нет url", "application/json", `{}`, 422, "invalid_url"},
		{"javascript:", "application/json", `{"url":"javascript:alert(1)"}`, 422, "invalid_url"},
		{"userinfo", "application/json", `{"url":"https://bank.ru@evil.example"}`, 422, "invalid_url"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, s := newAPI()
			w := do(t, h, "POST", "/api/links", tt.ctype, tt.body)
			if w.Code != tt.wantStatus {
				t.Fatalf("код %d, тело %s; want %d", w.Code, w.Body.String(), tt.wantStatus)
			}
			if c := errCode(t, w); c != tt.wantCode {
				t.Errorf("error.code %q, want %q", c, tt.wantCode)
			}
			if _, err := stats(t, s, "abc"); err == nil {
				t.Error("при ошибке ссылка всё равно сохранена")
			}
		})
	}
}

func TestCreateLinkExhausted(t *testing.T) {
	s := NewStore(func() string { return "same" }, fixedNow)
	h := NewHandler(s, base)
	do(t, h, "POST", "/api/links", "application/json", `{"url":"https://a.ru"}`)
	w := do(t, h, "POST", "/api/links", "application/json", `{"url":"https://b.ru"}`)
	if w.Code != 500 {
		t.Fatalf("код %d, want 500", w.Code)
	}
	if c := errCode(t, w); c != "internal" {
		t.Errorf("error.code %q, want internal", c)
	}
}

func TestRedirectAndStats(t *testing.T) {
	h, _ := newAPI()
	do(t, h, "POST", "/api/links", "application/json", `{"url":"https://go.dev/blog/"}`)

	for range 2 {
		w := do(t, h, "GET", "/abc", "", "")
		if w.Code != http.StatusFound {
			t.Fatalf("GET /abc: код %d, want 302 (не 301: его кеширует браузер)", w.Code)
		}
		if loc := w.Result().Header.Get("Location"); loc != "https://go.dev/blog/" {
			t.Errorf("GET /abc: Location %q, want https://go.dev/blog/", loc)
		}
	}

	w := do(t, h, "GET", "/api/links/abc/stats", "", "")
	if w.Code != 200 {
		t.Fatalf("GET stats: код %d, want 200", w.Code)
	}
	if ct := w.Result().Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("stats: Content-Type %q, want application/json", ct)
	}
	var got Link
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("stats: тело не JSON: %q", w.Body.String())
	}
	want := Link{Code: "abc", URL: "https://go.dev/blog/", Hits: 2, CreatedAt: t0}
	if got != want {
		t.Errorf("stats = %+v, want %+v", got, want)
	}
	if !strings.Contains(w.Body.String(), `"created_at":"2026-11-05T10:00:00Z"`) {
		t.Errorf("stats: тело %s, want поле created_at в RFC 3339", w.Body.String())
	}
}

func TestNotFoundAndMethods(t *testing.T) {
	h, _ := newAPI()
	tests := []struct {
		name, method, path string
		wantStatus         int
		wantAllow          string
	}{
		{"неизвестный код", "GET", "/zzz", 404, ""},
		{"stats неизвестного кода", "GET", "/api/links/zzz/stats", 404, ""},
		{"GET на создание", "GET", "/api/links", 405, "POST"},
		{"DELETE на код", "DELETE", "/abc", 405, "GET, HEAD"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := do(t, h, tt.method, tt.path, "", "")
			if w.Code != tt.wantStatus {
				t.Fatalf("%s %s: код %d, want %d", tt.method, tt.path, w.Code, tt.wantStatus)
			}
			if got := w.Result().Header.Get("Allow"); got != tt.wantAllow {
				t.Errorf("%s %s: Allow %q, want %q", tt.method, tt.path, got, tt.wantAllow)
			}
			if tt.wantStatus == 404 {
				if c := errCode(t, w); c != "not_found" {
					t.Errorf("error.code %q, want not_found", c)
				}
			}
		})
	}
}

func TestConcurrentAPI(t *testing.T) {
	var mu sync.Mutex
	n := 0
	s := NewStore(func() string { mu.Lock(); defer mu.Unlock(); n++; return fmt.Sprintf("k%d", n) }, fixedNow)
	srv := httptest.NewServer(NewHandler(s, base))
	defer func() {
		// srv.Close ждёт завершения всех хендлеров; если хендлер завис на
		// мьютексе, ждать не будем, тест уже упал с понятным сообщением.
		closed := make(chan struct{})
		go func() { srv.Close(); close(closed) }()
		select {
		case <-closed:
		case <-time.After(2 * time.Second):
			t.Error("хендлеры не завершились за 2с после теста: взаимоблокировка в Store?")
		}
	}()
	client := srv.Client()
	client.Timeout = 2 * time.Second // зависший хендлер не повесит тест
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }

	resp, err := client.Post(srv.URL+"/api/links", "application/json", strings.NewReader(`{"url":"https://hot.example"}`))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	var wg sync.WaitGroup
	for range 20 {
		wg.Go(func() {
			for range 10 {
				r, err := client.Get(srv.URL + "/k1")
				if err != nil {
					t.Error(err)
					return
				}
				r.Body.Close()
			}
		})
		wg.Go(func() {
			r, err := client.Post(srv.URL+"/api/links", "application/json", strings.NewReader(`{"url":"https://x.example"}`))
			if err == nil {
				r.Body.Close()
			}
		})
	}
	within(t, "wg.Wait для 200 параллельных запросов", wg.Wait)
	if st, _ := stats(t, s, "k1"); st.Hits != 200 {
		t.Errorf("после 200 параллельных переходов Hits = %d, want 200", st.Hits)
	}
}
