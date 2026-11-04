package projectexpenses

import (
	"encoding/json"
	"errors"
	"maps"
	"strings"
	"testing"
	"time"
)

func day(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 12, 0, 0, 0, time.UTC)
}

// sample собирает журнал: еда 250 + 100 в октябре, транспорт 60 в ноябре, еда 40 в ноябре.
func sample(t *testing.T) *Ledger {
	t.Helper()
	var l Ledger
	adds := []struct {
		amount   int64
		category string
		date     time.Time
	}{
		{25000, "Еда", day(2026, 10, 30)},
		{6000, "транспорт", day(2026, 11, 2)},
		{10000, " еда ", day(2026, 10, 1)},
		{4000, "ЕДА", day(2026, 11, 2)},
	}
	for _, a := range adds {
		if _, err := l.Add(a.amount, a.category, "", a.date); err != nil {
			t.Fatalf("Add(%d, %q): %v", a.amount, a.category, err)
		}
	}
	return &l
}

func ids(es []Expense) []int {
	out := make([]int, len(es))
	for i, e := range es {
		out[i] = e.ID
	}
	return out
}

func equalInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestAdd(t *testing.T) {
	tests := []struct {
		name     string
		amount   int64
		category string
		wantCat  string
		wantErr  error
	}{
		{"обычный расход", 15000, "Кафе", "кафе", nil},
		{"пробелы по краям", 100, "  Такси ", "такси", nil},
		{"нулевая сумма", 0, "еда", "", ErrInvalidAmount},
		{"отрицательная сумма", -5, "еда", "", ErrInvalidAmount},
		{"пустая категория", 100, "", "", ErrEmptyCategory},
		{"категория из пробелов", 100, "   ", "", ErrEmptyCategory},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var l Ledger
			got, err := l.Add(tt.amount, tt.category, "заметка", day(2026, 11, 4))
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Add(%d, %q) error = %v, want %v", tt.amount, tt.category, err, tt.wantErr)
			}
			if tt.wantErr != nil {
				if n := len(l.List()); n != 0 {
					t.Errorf("после ошибки в журнале %d расходов, want 0", n)
				}
				return
			}
			if got.ID != 1 || got.Category != tt.wantCat || got.Amount != tt.amount || got.Note != "заметка" {
				t.Errorf("Add(%d, %q) = %+v, want ID 1, Category %q", tt.amount, tt.category, got, tt.wantCat)
			}
		})
	}
}

func TestAddIDsAfterError(t *testing.T) {
	var l Ledger
	l.Add(100, "a", "", day(2026, 1, 1))
	l.Add(-1, "a", "", day(2026, 1, 1))
	e, _ := l.Add(200, "b", "", day(2026, 1, 1))
	if e.ID != 2 {
		t.Errorf("ID после неудачного Add = %d, want 2: ошибка не должна тратить ID", e.ID)
	}
}

func TestRemove(t *testing.T) {
	l := sample(t)
	if err := l.Remove(2); err != nil {
		t.Fatalf("Remove(2): %v", err)
	}
	if err := l.Remove(2); !errors.Is(err, ErrNotFound) {
		t.Errorf("повторный Remove(2) = %v, want ErrNotFound", err)
	}
	if err := l.Remove(42); !errors.Is(err, ErrNotFound) {
		t.Errorf("Remove(42) = %v, want ErrNotFound", err)
	}
	if got, want := ids(l.List()), []int{3, 1, 4}; !equalInts(got, want) {
		t.Errorf("после Remove(2): ID в List() = %v, want %v", got, want)
	}
	e, _ := l.Add(1, "x", "", day(2026, 12, 1))
	if e.ID != 5 {
		t.Errorf("ID после удаления = %d, want 5: ID не переиспользуются", e.ID)
	}
}

func TestListSortedCopy(t *testing.T) {
	l := sample(t)
	list := l.List()
	// 3 (1 окт), 1 (30 окт), 2 и 4 (2 ноя, равная дата: по ID)
	if got, want := ids(list), []int{3, 1, 2, 4}; !equalInts(got, want) {
		t.Fatalf("ID в List() = %v, want %v", got, want)
	}
	list[0].Amount = 1
	if l.List()[0].Amount != 10000 {
		t.Errorf("изменение результата List() изменило журнал")
	}
	var empty Ledger
	if got := empty.List(); len(got) != 0 {
		t.Errorf("List() пустого журнала = %v, want пустой", got)
	}
}

func TestTotals(t *testing.T) {
	l := sample(t)
	if got, want := l.TotalsByCategory(), map[string]int64{"еда": 39000, "транспорт": 6000}; !maps.Equal(got, want) {
		t.Errorf("TotalsByCategory() = %v, want %v", got, want)
	}
	if got, want := l.TotalsByMonth(), map[string]int64{"2026-10": 35000, "2026-11": 10000}; !maps.Equal(got, want) {
		t.Errorf("TotalsByMonth() = %v, want %v", got, want)
	}
	var empty Ledger
	if got := empty.TotalsByCategory(); got == nil || len(got) != 0 {
		t.Errorf("TotalsByCategory() пустого журнала = %#v, want пустую не-nil мапу", got)
	}
	if got := empty.TotalsByMonth(); got == nil || len(got) != 0 {
		t.Errorf("TotalsByMonth() пустого журнала = %#v, want пустую не-nil мапу", got)
	}
}

func TestTotalsByMonthUsesOwnTimezone(t *testing.T) {
	msk := time.FixedZone("MSK", 3*60*60)
	var l Ledger
	// 1 декабря 01:00 по Москве это ещё 30 ноября по UTC, но месяц берём по дате расхода.
	l.Add(500, "еда", "", time.Date(2026, 12, 1, 1, 0, 0, 0, msk))
	if got, want := l.TotalsByMonth(), map[string]int64{"2026-12": 500}; !maps.Equal(got, want) {
		t.Errorf("TotalsByMonth() = %v, want %v", got, want)
	}
}

func TestEncodeDecodeRoundTrip(t *testing.T) {
	l := sample(t)
	l.Remove(4) // last_id остаётся 4
	data, err := Encode(l)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	var raw struct {
		LastID   int               `json:"last_id"`
		Expenses []json.RawMessage `json:"expenses"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Encode дал невалидный JSON: %v\n%s", err, data)
	}
	if raw.LastID != 4 || len(raw.Expenses) != 3 {
		t.Errorf("Encode: last_id = %d, расходов %d; want 4 и 3\n%s", raw.LastID, len(raw.Expenses), data)
	}
	if !strings.Contains(string(data), `"amount_kop"`) {
		t.Errorf("Encode: в JSON нет поля amount_kop\n%s", data)
	}

	back, err := Decode(data)
	if err != nil {
		t.Fatalf("Decode(Encode(l)): %v", err)
	}
	orig, got := l.List(), back.List()
	if len(orig) != len(got) {
		t.Fatalf("после Decode %d расходов, want %d", len(got), len(orig))
	}
	for i := range orig {
		o, g := orig[i], got[i]
		if o.ID != g.ID || o.Amount != g.Amount || o.Category != g.Category || o.Note != g.Note || !o.Date.Equal(g.Date) {
			t.Errorf("расход %d после Decode = %+v, want %+v", i, g, o)
		}
	}
	e, _ := back.Add(1, "x", "", day(2026, 12, 1))
	if e.ID != 5 {
		t.Errorf("Add после Decode выдал ID %d, want 5", e.ID)
	}
}

func TestEncodeEmpty(t *testing.T) {
	var l Ledger
	data, err := Encode(&l)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	compact := strings.Join(strings.Fields(string(data)), "")
	if !strings.Contains(compact, `"expenses":[]`) {
		t.Errorf("Encode(пустой журнал) = %s, want \"expenses\": []", data)
	}
}

func TestDecodeFixesLastID(t *testing.T) {
	data := `{"last_id": 1, "expenses": [{"id": 7, "amount_kop": 100, "category": "еда", "date": "2026-11-04T12:00:00Z"}]}`
	l, err := Decode([]byte(data))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	e, _ := l.Add(1, "x", "", day(2026, 12, 1))
	if e.ID != 8 {
		t.Errorf("Add после Decode выдал ID %d, want 8 (максимальный ID + 1)", e.ID)
	}
}

func TestDecodeNormalizesCategory(t *testing.T) {
	data := `{"last_id": 2, "expenses": [
		{"id": 1, "amount_kop": 100, "category": " Еда ", "date": "2026-11-04T12:00:00Z"},
		{"id": 2, "amount_kop": 50, "category": "еда", "date": "2026-11-04T12:00:00Z"}]}`
	l, err := Decode([]byte(data))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if got, want := l.TotalsByCategory(), map[string]int64{"еда": 150}; !maps.Equal(got, want) {
		t.Errorf("TotalsByCategory() после Decode = %v, want %v: категории из файла надо нормализовать", got, want)
	}
}

func TestRemoveKeepsInsertionOrder(t *testing.T) {
	l := sample(t) // ID 1, 2, 3, 4 в порядке добавления
	if err := l.Remove(2); err != nil {
		t.Fatalf("Remove(2): %v", err)
	}
	data, err := Encode(l)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	var f struct {
		Expenses []Expense `json:"expenses"`
	}
	if err := json.Unmarshal(data, &f); err != nil {
		t.Fatalf("Encode дал невалидный JSON: %v", err)
	}
	if got, want := ids(f.Expenses), []int{1, 3, 4}; !equalInts(got, want) {
		t.Errorf("порядок ID в Encode после Remove(2) = %v, want %v: Remove должен сохранять порядок добавления", got, want)
	}
}

func TestDecodeCorrupt(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{"не JSON", `{"last_id": `},
		{"отрицательный last_id", `{"last_id": -5, "expenses": []}`},
		{"повтор ID", `{"last_id": 2, "expenses": [{"id": 1, "amount_kop": 1, "category": "a", "date": "2026-11-04T12:00:00Z"}, {"id": 1, "amount_kop": 2, "category": "b", "date": "2026-11-04T12:00:00Z"}]}`},
		{"нулевой ID", `{"last_id": 1, "expenses": [{"id": 0, "amount_kop": 1, "category": "a", "date": "2026-11-04T12:00:00Z"}]}`},
		{"отрицательная сумма", `{"last_id": 1, "expenses": [{"id": 1, "amount_kop": -1, "category": "a", "date": "2026-11-04T12:00:00Z"}]}`},
		{"пустая категория", `{"last_id": 1, "expenses": [{"id": 1, "amount_kop": 1, "category": " ", "date": "2026-11-04T12:00:00Z"}]}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := Decode([]byte(tt.data)); !errors.Is(err, ErrCorrupt) {
				t.Errorf("Decode(%s) error = %v, want ErrCorrupt", tt.data, err)
			}
		})
	}
}
