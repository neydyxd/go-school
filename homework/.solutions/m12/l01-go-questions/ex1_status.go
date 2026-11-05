package l01goquestions

import (
	"errors"
	"fmt"
	"strings"
)

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

var statusNames = [...]string{"unknown", "active", "blocked", "deleted"}

func (s Status) valid() bool { return s >= 0 && int(s) < len(statusNames) }

// String возвращает имя статуса или "Status(N)" для значения вне списка.
func (s Status) String() string {
	if !s.valid() {
		return fmt.Sprintf("Status(%d)", int(s))
	}
	return statusNames[s]
}

// ParseStatus разбирает имя статуса без учёта регистра.
func ParseStatus(s string) (Status, error) {
	for i, name := range statusNames {
		if strings.EqualFold(s, name) {
			return Status(i), nil
		}
	}
	return StatusUnknown, fmt.Errorf("%w: %q", ErrBadStatus, s)
}

// MarshalText отдаёт имя статуса; невалидное значение даёт ошибку.
func (s Status) MarshalText() ([]byte, error) {
	if !s.valid() {
		return nil, fmt.Errorf("%w: %d", ErrBadStatus, int(s))
	}
	return []byte(statusNames[s]), nil
}

// UnmarshalText разбирает строку через ParseStatus. При ошибке *s не меняется.
func (s *Status) UnmarshalText(b []byte) error {
	v, err := ParseStatus(string(b))
	if err != nil {
		return err
	}
	*s = v
	return nil
}
