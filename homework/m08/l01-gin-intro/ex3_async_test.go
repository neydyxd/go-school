package l01ginintro

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

type sent struct{ user, id string }

// Запускай с -race: go test -race ./m08/l01-gin-intro/ -run TestReport
func TestReportHandlerUsesCopy(t *testing.T) {
	gate := make(chan struct{})
	got := make(chan sent, 1)
	h := ReportHandler(func() { <-gate }, func(user, id string) { got <- sent{user, id} })

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/reports/7", nil)
	c.Params = gin.Params{{Key: "id", Value: "7"}}
	c.Set("user", "alice")

	h(c)

	// Ответ должен быть готов к моменту выхода из обработчика, а не когда-нибудь из горутины.
	if w.Code != http.StatusAccepted {
		t.Fatalf("код %d, want 202", w.Code)
	}
	if body := decodeOne(t, w.Body.Bytes()); !reflect.DeepEqual(body, map[string]any{"status": "accepted"}) {
		t.Fatalf("тело %v, want {\"status\": \"accepted\"}", body)
	}

	// Так Gin переиспользует контекст: следующий запрос перезаписывает параметры и ключи.
	c.Params[0].Value = "999"
	c.Set("user", "mallory")
	close(gate)

	select {
	case s := <-got:
		if want := (sent{"alice", "7"}); s != want {
			t.Fatalf("send(%q, %q), want send(%q, %q): горутина читает исходный контекст после возврата обработчика",
				s.user, s.id, want.user, want.id)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("send не вызван за 2 секунды")
	}
}

func TestReportHandlerThroughRouter(t *testing.T) {
	got := make(chan sent, 10)
	r := gin.New()
	r.POST("/reports/:id", func(c *gin.Context) { c.Set("user", c.GetHeader("X-User")) },
		ReportHandler(func() {}, func(user, id string) { got <- sent{user, id} }))

	want := map[sent]bool{}
	for _, s := range []sent{{"alice", "1"}, {"bob", "2"}, {"carol", "3"}} {
		w := serve(r, http.MethodPost, "/reports/"+s.id, map[string]string{"X-User": s.user})
		if w.Code != http.StatusAccepted {
			t.Fatalf("POST /reports/%s: код %d, want 202", s.id, w.Code)
		}
		want[s] = true
	}
	for range 3 {
		select {
		case s := <-got:
			if !want[s] {
				t.Fatalf("send(%q, %q): такой пары не было в запросах", s.user, s.id)
			}
			delete(want, s)
		case <-time.After(2 * time.Second):
			t.Fatalf("за 2 секунды не дошли отчёты: %v", want)
		}
	}
}
