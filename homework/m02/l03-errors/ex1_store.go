package l03errors

import "errors"

// Сигнальные (sentinel) ошибки пакета. Сравнивать их нужно через errors.Is.
var (
	ErrNotFound = errors.New("not found")
	ErrEmptyKey = errors.New("empty key")
)

// Store: простое хранилище строк по ключу.
type Store struct {
	data map[string]string
}

// NewStore уже написана.
func NewStore(data map[string]string) *Store {
	return &Store{data: data}
}

// Get возвращает значение по ключу.
//   - пустой ключ: "", ErrEmptyKey (без обёртки);
//   - ключа нет: "", ошибка с текстом get "<key>": not found,
//     для которой errors.Is(err, ErrNotFound) == true;
//   - ключ есть: значение, nil (даже если значение пустая строка).
//
// Хранилище с nil-мапой внутри должно работать как пустое.
func (s *Store) Get(key string) (string, error) {
	panic("TODO: реализуй Get")
}
