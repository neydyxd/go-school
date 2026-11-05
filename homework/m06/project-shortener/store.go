package shortener

import (
	"errors"
	"time"
)

var (
	// ErrNotFound: кода нет в хранилище.
	ErrNotFound = errors.New("link not found")
	// ErrCodeExhausted: генератор пять раз подряд выдал занятый код.
	ErrCodeExhausted = errors.New("could not generate unique code")
)

// Link: сохранённая ссылка.
type Link struct {
	Code      string    `json:"code"`
	URL       string    `json:"url"`
	Hits      int64     `json:"hits"`
	CreatedAt time.Time `json:"created_at"`
}

// Store: потокобезопасное хранилище ссылок в памяти. Методы вызываются
// из множества горутин одновременно (по горутине на HTTP-запрос).
type Store struct {
	// TODO: поля (мапа, мьютекс, gen, now)
}

// NewStore создаёт пустое хранилище. gen выдаёт новый код (в проде
// func() string { return NewCode(7) }, в тестах предсказуемая
// последовательность), now выдаёт текущее время (в проде time.Now).
func NewStore(gen func() string, now func() time.Time) *Store {
	panic("TODO: реализуй NewStore")
}

// Create сохраняет url под новым кодом и возвращает Link (Hits 0,
// CreatedAt = now()). URL уже проверен вызывающим. Если gen выдал занятый
// код, попросить новый; пять занятых подряд: ErrCodeExhausted.
// Одинаковые url можно сокращать много раз: каждый раз новый код.
func (s *Store) Create(url string) (Link, error) {
	panic("TODO: реализуй Create")
}

// Resolve возвращает url по коду и увеличивает счётчик переходов на 1.
// Нет кода: ErrNotFound.
func (s *Store) Resolve(code string) (string, error) {
	panic("TODO: реализуй Resolve")
}

// Stats возвращает копию Link по коду, счётчик не меняет. Нет кода: ErrNotFound.
func (s *Store) Stats(code string) (Link, error) {
	panic("TODO: реализуй Stats")
}
