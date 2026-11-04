// Package projectexpenses: ядро мини-проекта модуля 01, учёт расходов.
// CLI-обвязку (команды add, rm, list, report и файл на диске) пишешь сам по ТЗ из урока,
// а здесь логика, которую CLI вызывает.
package projectexpenses

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
)

// Ошибки, которые возвращают методы. Проверяй их через errors.Is.
var (
	ErrInvalidAmount = errors.New("сумма должна быть больше нуля")
	ErrEmptyCategory = errors.New("категория не может быть пустой")
	ErrNotFound      = errors.New("расход не найден")
	ErrCorrupt       = errors.New("файл расходов повреждён")
)

// Expense: один расход. Деньги храним в копейках целым числом:
// float64 для денег не годится (0.1 + 0.2 != 0.3).
type Expense struct {
	ID       int       `json:"id"`
	Amount   int64     `json:"amount_kop"`
	Category string    `json:"category"`
	Note     string    `json:"note,omitempty"`
	Date     time.Time `json:"date"`
}

// Ledger: журнал расходов. Нулевое значение готово к работе: var l Ledger.
type Ledger struct {
	lastID   int       // последний выданный ID; ID не переиспользуются даже после удаления
	expenses []Expense // в порядке добавления
}

// Add добавляет расход и возвращает его с присвоенным ID (1, 2, 3...).
//
// Категория очищается от пробелов по краям и приводится к нижнему регистру:
// " Еда " и "еда" это одна категория.
// amount <= 0: ErrInvalidAmount. Пустая после очистки категория: ErrEmptyCategory.
// При ошибке журнал не меняется и ID не тратится.
func (l *Ledger) Add(amount int64, category, note string, date time.Time) (Expense, error) {
	category = normalizeCategory(category)
	if err := validate(amount, category); err != nil {
		return Expense{}, err
	}
	l.lastID++
	e := Expense{ID: l.lastID, Amount: amount, Category: category, Note: note, Date: date}
	l.expenses = append(l.expenses, e)
	return e, nil
}

func normalizeCategory(c string) string {
	return strings.ToLower(strings.TrimSpace(c))
}

func validate(amount int64, category string) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	if category == "" {
		return ErrEmptyCategory
	}
	return nil
}

// Remove удаляет расход по ID, порядок остальных (порядок добавления,
// он виден в Encode) сохраняется.
// Нет такого ID: ErrNotFound.
func (l *Ledger) Remove(id int) error {
	i := slices.IndexFunc(l.expenses, func(e Expense) bool { return e.ID == id })
	if i < 0 {
		return ErrNotFound
	}
	l.expenses = slices.Delete(l.expenses, i, i+1)
	return nil
}

// List возвращает копию всех расходов, отсортированную по дате,
// при равной дате по ID. Изменение результата не должно менять журнал.
// Пустой журнал: слайс длины 0.
func (l *Ledger) List() []Expense {
	out := slices.Clone(l.expenses)
	slices.SortFunc(out, func(a, b Expense) int {
		if c := a.Date.Compare(b.Date); c != 0 {
			return c
		}
		return cmp.Compare(a.ID, b.ID)
	})
	return out
}

// TotalsByCategory возвращает сумму расходов по каждой категории.
// Пустой журнал: пустая не-nil мапа.
func (l *Ledger) TotalsByCategory() map[string]int64 {
	return l.totals(func(e Expense) string { return e.Category })
}

// TotalsByMonth возвращает сумму расходов по месяцам, ключ в формате "2026-11"
// (год и месяц даты расхода в её собственном часовом поясе).
// Пустой журнал: пустая не-nil мапа.
func (l *Ledger) TotalsByMonth() map[string]int64 {
	return l.totals(func(e Expense) string { return e.Date.Format("2006-01") })
}

func (l *Ledger) totals(key func(Expense) string) map[string]int64 {
	out := make(map[string]int64)
	for _, e := range l.expenses {
		out[key(e)] += e.Amount
	}
	return out
}

// file: формат хранения на диске.
type file struct {
	LastID   int       `json:"last_id"`
	Expenses []Expense `json:"expenses"`
}

// Encode сериализует журнал в JSON такого вида:
//
//	{"last_id": 3, "expenses": [{"id": 1, "amount_kop": 25000, ...}]}
//
// Пустой журнал даёт "expenses": [] (а не null). Отступы на твой вкус.
func Encode(l *Ledger) ([]byte, error) {
	f := file{LastID: l.lastID, Expenses: l.expenses}
	if f.Expenses == nil {
		f.Expenses = []Expense{} // nil-слайс кодируется как null
	}
	return json.MarshalIndent(f, "", "  ")
}

// Decode восстанавливает журнал из JSON, который сделал Encode.
//
// Невалидный JSON или данные, нарушающие правила журнала (last_id < 0, ID <= 0,
// повторяющийся ID, сумма <= 0, пустая категория), дают ошибку,
// для которой errors.Is(err, ErrCorrupt) == true.
// Если last_id в файле меньше максимального ID расходов, бери максимальный ID:
// новые расходы не должны получить уже занятый номер.
// Файл могли править руками, поэтому категории нормализуются так же, как в Add:
// " Еда " из файла становится "еда".
func Decode(data []byte) (*Ledger, error) {
	var f file
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCorrupt, err)
	}
	if f.LastID < 0 {
		return nil, fmt.Errorf("%w: last_id %d", ErrCorrupt, f.LastID)
	}
	seen := make(map[int]struct{}, len(f.Expenses))
	for i := range f.Expenses {
		e := &f.Expenses[i] // указатель: нормализуем категорию прямо в слайсе
		if e.ID <= 0 {
			return nil, fmt.Errorf("%w: ID %d", ErrCorrupt, e.ID)
		}
		if _, dup := seen[e.ID]; dup {
			return nil, fmt.Errorf("%w: ID %d повторяется", ErrCorrupt, e.ID)
		}
		seen[e.ID] = struct{}{}
		e.Category = normalizeCategory(e.Category)
		if err := validate(e.Amount, e.Category); err != nil {
			return nil, fmt.Errorf("%w: расход %d: %v", ErrCorrupt, e.ID, err)
		}
		f.LastID = max(f.LastID, e.ID)
	}
	return &Ledger{lastID: f.LastID, expenses: f.Expenses}, nil
}
