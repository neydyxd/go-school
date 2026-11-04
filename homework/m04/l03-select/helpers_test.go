package l03select

import (
	"runtime"
	"testing"
	"testing/synctest"
	"time"
)

// within запускает f и ждёт её не дольше d. Зависшее решение не вешает
// тесты, а паника (например, panic("TODO")) превращается в понятный провал.
func within(t *testing.T, d time.Duration, what string, f func()) {
	t.Helper()
	done := make(chan any, 1)
	go func() {
		defer func() { done <- recover() }()
		f()
	}()
	select {
	case r := <-done:
		if r != nil {
			t.Fatalf("%s: паника: %v", what, r)
		}
	case <-time.After(d):
		t.Fatalf("%s: не завершилось за %v (зависание или дедлок?)", what, d)
	}
}

// noLeak ждёт до секунды, пока число горутин вернётся к before.
// Если не вернулось, какая-то горутина решения осталась висеть.
func noLeak(t *testing.T, before int) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for runtime.NumGoroutine() > before {
		if time.Now().After(deadline) {
			t.Fatalf("утечка горутин: было %d, осталось %d", before, runtime.NumGoroutine())
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// bubble запускает f в пузыре testing/synctest: время внутри фейковое и идёт
// скачками, только когда все горутины пузыря спят, а synctest.Wait()
// дожидается, пока все они уснут. Так тесты на «блокируется ли» и «через
// сколько вернулось» не зависят от скорости машины.
//
// Если решение зависло навсегда, within внутри пузыря провалит тест, а
// synctest сообщит о горутинах, которые остались заблокированы: эту панику
// превращаем в обычный провал, чтобы остальные тесты пакета тоже прогнались.
// Решение, которое крутится в цикле, не засыпая, остановит фейковое время
// навсегда. На этот случай сторож в реальном времени роняет тестовый
// бинарник с понятным сообщением, а не вешает его.
func bubble(t *testing.T, f func(t *testing.T)) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("в решении остались навсегда заблокированные горутины: %v", r)
		}
	}()
	done := make(chan struct{})
	name := t.Name()
	go func() {
		select {
		case <-done:
		case <-time.After(10 * time.Second):
			panic(name + ": тест не закончился за 10 с реального времени: похоже, решение крутится в цикле, не блокируясь (busy wait)")
		}
	}()
	defer close(done)
	synctest.Test(t, f)
}
