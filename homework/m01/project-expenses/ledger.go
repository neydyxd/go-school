// Package projectexpenses: ядро мини-проекта модуля 01, учёт расходов.
// CLI-обвязку (команды add, rm, list, report и файл на диске) пишешь сам по ТЗ из урока,
// а здесь логика, которую CLI вызывает.
package projectexpenses

import (
	"errors"
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
	panic("TODO: реализуй Add")
}

// Remove удаляет расход по ID, порядок остальных (порядок добавления,
// он виден в Encode) сохраняется.
// Нет такого ID: ErrNotFound.
func (l *Ledger) Remove(id int) error {
	panic("TODO: реализуй Remove")
}

// List возвращает копию всех расходов, отсортированную по дате,
// при равной дате по ID. Изменение результата не должно менять журнал.
// Пустой журнал: слайс длины 0.
func (l *Ledger) List() []Expense {
	panic("TODO: реализуй List")
}

// TotalsByCategory возвращает сумму расходов по каждой категории.
// Пустой журнал: пустая не-nil мапа.
func (l *Ledger) TotalsByCategory() map[string]int64 {
	panic("TODO: реализуй TotalsByCategory")
}

// TotalsByMonth возвращает сумму расходов по месяцам, ключ в формате "2026-11"
// (год и месяц даты расхода в её собственном часовом поясе).
// Пустой журнал: пустая не-nil мапа.
func (l *Ledger) TotalsByMonth() map[string]int64 {
	panic("TODO: реализуй TotalsByMonth")
}

// Encode сериализует журнал в JSON такого вида:
//
//	{"last_id": 3, "expenses": [{"id": 1, "amount_kop": 25000, ...}]}
//
// Пустой журнал даёт "expenses": [] (а не null). Отступы на твой вкус.
func Encode(l *Ledger) ([]byte, error) {
	panic("TODO: реализуй Encode")
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
	panic("TODO: реализуй Decode")
}
