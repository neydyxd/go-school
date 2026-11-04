package l06stdlib

import (
	"errors"
	"time"
)

// ErrNoName: в событии нет имени.
var ErrNoName = errors.New("event without name")

// Event: событие деплоя. Добавь структурные теги так, чтобы JSON выглядел так:
//
//	{"name":"deploy","at":"2026-03-09T14:05:07Z","tags":["prod"]}
//
// Поле tags не выводится, если тегов нет (nil или пустой слайс).
type Event struct {
	Name string
	At   time.Time
	Tags []string
}

// EncodeEvent кодирует событие в JSON (без переводов строк и отступов).
func EncodeEvent(e Event) ([]byte, error) {
	panic("TODO: реализуй EncodeEvent")
}

// DecodeEvent строго разбирает JSON события:
//   - неизвестное поле (например "nmae") это ошибка;
//   - пустое или отсутствующее имя: ошибка с ErrNoName;
//   - мусор после объекта (например "{...} {...}") это ошибка.
func DecodeEvent(data []byte) (Event, error) {
	panic("TODO: реализуй DecodeEvent")
}
