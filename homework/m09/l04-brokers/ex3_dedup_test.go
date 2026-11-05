package l04brokers

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestDeduper(t *testing.T) {
	tests := []struct {
		name     string
		capacity int
		ids      []string
		want     []bool
	}{
		{"повтор сразу", 3, []string{"a", "a"}, []bool{false, true}},
		{"повтор внутри окна", 3, []string{"a", "b", "c", "a"}, []bool{false, false, false, true}},
		{"самый давний вытеснен", 3, []string{"a", "b", "c", "d", "a"}, []bool{false, false, false, false, false}},
		{"вытесняется только один", 3, []string{"a", "b", "c", "d", "b", "c", "d"}, []bool{false, false, false, false, true, true, true}},
		{"дубль не продлевает жизнь", 2, []string{"a", "b", "a", "c", "a"}, []bool{false, false, true, false, false}},
		{"дубль не занимает место", 2, []string{"a", "a", "a", "b", "a"}, []bool{false, true, true, false, true}},
		{"окно 0 это окно 1", 0, []string{"a", "a", "b", "a"}, []bool{false, true, false, false}},
		{"пустой id тоже id", 2, []string{"", ""}, []bool{false, true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDeduper(tt.capacity)
			var got []bool
			for _, id := range tt.ids {
				got = append(got, d.Seen(id))
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDeduper(%d), Seen по %q = %v, want %v", tt.capacity, tt.ids, got, tt.want)
			}
		})
	}
}

// Окно на 100 тысяч id и полмиллиона сообщений: решение за O(capacity) на вызов
// (поиск или сдвиг слайса) не уложится в таймаут.
func TestDeduperFast(t *testing.T) {
	done := make(chan int, 1)
	go func() {
		d := NewDeduper(100_000)
		dups := 0
		for i := range 500_000 {
			if d.Seen(fmt.Sprint(i % 150_000)) {
				dups++
			}
		}
		done <- dups
	}()
	select {
	case dups := <-done:
		// i%150000 повторяется через 150000 разных id, а окно 100000: дублей нет
		if dups != 0 {
			t.Errorf("найдено %d дублей, want 0: id повторяется только после 150000 других", dups)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("500000 вызовов Seen не уложились в 5 секунд: нужен O(1) на вызов")
	}
}
