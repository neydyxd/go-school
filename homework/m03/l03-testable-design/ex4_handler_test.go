package l03testabledesign

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestStatusHandlerGET(t *testing.T) {
	msk := time.FixedZone("MSK", 3*60*60)
	clock := fakeClock{time.Date(2026, 11, 4, 12, 30, 0, 0, msk)}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/status", nil)

	StatusHandler(clock).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /status: код %d, want 200", rec.Code)
	}
	// Result() это снимок ответа на момент первой записи: заголовки,
	// поставленные после w.Write, туда не попадут, как и в настоящем сервере.
	if ct := rec.Result().Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("GET /status: Content-Type %q, want application/json (заголовки ставятся до записи тела)", ct)
	}
	var body struct {
		Status string `json:"status"`
		Time   string `json:"time"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("GET /status: тело %q не JSON: %v", rec.Body.String(), err)
	}
	if body.Status != "ok" || body.Time != "2026-11-04T09:30:00Z" {
		t.Errorf("GET /status: тело %+v, want status ok и time 2026-11-04T09:30:00Z (UTC)", body)
	}
}

func TestStatusHandlerMethods(t *testing.T) {
	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodDelete} {
		t.Run(method, func(t *testing.T) {
			rec := httptest.NewRecorder()
			StatusHandler(at(9, 0)).ServeHTTP(rec, httptest.NewRequest(method, "/status", nil))
			if rec.Code != http.StatusMethodNotAllowed {
				t.Errorf("%s /status: код %d, want 405", method, rec.Code)
			}
			if allow := rec.Result().Header.Get("Allow"); allow != "GET" {
				t.Errorf("%s /status: Allow %q, want GET", method, allow)
			}
		})
	}
}
