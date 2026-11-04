package l07problems

import (
	"strings"
	"testing"
	"time"
)

func TestCheckLeaksClean(t *testing.T) {
	start := time.Now()
	err := CheckLeaks(func() {
		done := make(chan struct{})
		go func() { close(done) }()
		<-done
	}, time.Second)
	if err != nil {
		t.Errorf("CheckLeaks(чистая функция) = %v, want nil", err)
	}
	if d := time.Since(start); d > 300*time.Millisecond {
		t.Errorf("CheckLeaks думал %v над чистой функцией, want почти мгновенно: проверяй сразу и периодически, а не в конце таймаута", d)
	}
}

func TestCheckLeaksWaits(t *testing.T) {
	start := time.Now()
	err := CheckLeaks(func() {
		go func() { time.Sleep(50 * time.Millisecond) }()
	}, 2*time.Second)
	if err != nil {
		t.Errorf("горутина завершается через 50ms, а CheckLeaks вернул %v: надо подождать, а не проверять один раз", err)
	}
	if d := time.Since(start); d > 500*time.Millisecond {
		t.Errorf("CheckLeaks думал %v: горутина завершилась через 50ms, проверяй периодически, а не в конце таймаута 2s", d)
	}
}

// leakyWorker висит, пока не закроют stop. Имя функции должно попасть в
// стеки, которые вернёт CheckLeaks.
func leakyWorker(stop chan struct{}) { <-stop }

func TestCheckLeaksFindsLeak(t *testing.T) {
	stop := make(chan struct{})
	t.Cleanup(func() { close(stop) }) // после теста отпускаем «утёкшие» горутины
	start := time.Now()
	err := CheckLeaks(func() {
		go leakyWorker(stop)
		go leakyWorker(stop)
	}, 200*time.Millisecond)
	if err == nil {
		t.Fatal("CheckLeaks не заметил две висящие горутины")
	}
	if d := time.Since(start); d < 200*time.Millisecond || d > 2*time.Second {
		t.Errorf("CheckLeaks думал %v, want около таймаута 200ms", d)
	}
	msg := err.Error()
	first, _, _ := strings.Cut(msg, "\n")
	if first != "лишних горутин: 2" {
		t.Errorf("первая строка ошибки %q, want %q", first, "лишних горутин: 2")
	}
	if !strings.Contains(msg, "leakyWorker") {
		t.Errorf("в ошибке нет стека с leakyWorker:\n%.500s", msg)
	}
}
