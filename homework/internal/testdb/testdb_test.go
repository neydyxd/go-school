package testdb

import (
	"reflect"
	"testing"
)

func TestIsolationAndFormatting(t *testing.T) {
	a := New(t)
	b := New(t)
	Exec(t, a, `CREATE TABLE items (id int, price numeric(10,2), note text, uid uuid); INSERT INTO items VALUES (1, 9.90, NULL, '8f14e45f-ceea-467f-a0e6-0f1d7f6c2c11')`)

	got := Rows(t, a, `SELECT id, price, note, uid FROM items`)
	want := [][]string{{"1", "9.90", "NULL", "8f14e45f-ceea-467f-a0e6-0f1d7f6c2c11"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Rows = %v, want %v", got, want)
	}
	// у второго теста своя схема: таблицы items там нет
	if n := Rows(t, b, `SELECT count(*) FROM information_schema.tables WHERE table_schema = current_schema() AND table_name = 'items'`); n[0][0] != "0" {
		t.Fatalf("схемы не изолированы: в соседней схеме видна items")
	}
}
