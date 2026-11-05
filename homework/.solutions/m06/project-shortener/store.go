package shortener

import (
	"errors"
	"sync"
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
	mu    sync.Mutex
	links map[string]*Link
	gen   func() string
	now   func() time.Time
}

// NewStore создаёт пустое хранилище. gen выдаёт новый код (в проде
// func() string { return NewCode(7) }, в тестах предсказуемая
// последовательность), now выдаёт текущее время (в проде time.Now).
func NewStore(gen func() string, now func() time.Time) *Store {
	return &Store{links: map[string]*Link{}, gen: gen, now: now}
}

// Create сохраняет url под новым кодом и возвращает Link (Hits 0,
// CreatedAt = now()). URL уже проверен вызывающим. Если gen выдал занятый
// код, попросить новый; пять занятых подряд: ErrCodeExhausted.
// Одинаковые url можно сокращать много раз: каждый раз новый код.
func (s *Store) Create(url string) (Link, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for range 5 {
		code := s.gen()
		if _, taken := s.links[code]; taken {
			continue
		}
		l := &Link{Code: code, URL: url, CreatedAt: s.now()}
		s.links[code] = l
		return *l, nil
	}
	return Link{}, ErrCodeExhausted
}

// Resolve возвращает url по коду и увеличивает счётчик переходов на 1.
// Нет кода: ErrNotFound.
func (s *Store) Resolve(code string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	l, ok := s.links[code]
	if !ok {
		return "", ErrNotFound
	}
	l.Hits++
	return l.URL, nil
}

// Stats возвращает копию Link по коду, счётчик не меняет. Нет кода: ErrNotFound.
func (s *Store) Stats(code string) (Link, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	l, ok := s.links[code]
	if !ok {
		return Link{}, ErrNotFound
	}
	return *l, nil
}
