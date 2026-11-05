package l01goquestions

import "errors"

// ErrBadStatus возвращается, когда строка или число не соответствуют ни одному статусу.
var ErrBadStatus = errors.New("неизвестный статус")

// Status: статус пользователя. Нулевое значение StatusUnknown значит «не задан».
type Status int

const (
	StatusUnknown Status = iota
	StatusActive
	StatusBlocked
	StatusDeleted
)

// String возвращает имя статуса: "unknown", "active", "blocked", "deleted".
// Для значения вне списка (Status(42), Status(-1)) возвращает "Status(42)", "Status(-1)"
// и не паникует.
func (s Status) String() string {
	panic("TODO: реализуй String")
}

// ParseStatus разбирает имя статуса без учёта регистра ("Active" тоже подходит).
// Любая другая строка, включая пустую и число "1", даёт ошибку, для которой
// errors.Is(err, ErrBadStatus) истинно; статус при этом StatusUnknown.
func ParseStatus(s string) (Status, error) {
	panic("TODO: реализуй ParseStatus")
}

// MarshalText нужен encoding/json: статус уходит в JSON строкой ("blocked"), а не числом.
// Для значения вне списка возвращает ошибку с ErrBadStatus: невалидное не должно уйти наружу.
func (s Status) MarshalText() ([]byte, error) {
	return nil, nil
}

// UnmarshalText разбирает строку через ParseStatus. При ошибке *s не меняется.
func (s *Status) UnmarshalText(b []byte) error {
	return nil
}
