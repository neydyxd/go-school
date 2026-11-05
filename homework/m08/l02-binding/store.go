package l02binding

import (
	"errors"
	"slices"
	"sync"
)

// Этот файл готов, менять его не нужно: хранилище в памяти и формат ответов.

// Product это товар каталога. Цена в копейках, чтобы не связываться с float.
type Product struct {
	ID       int64    `json:"id"`
	Name     string   `json:"name"`
	Slug     string   `json:"slug"`
	Price    int64    `json:"price"`
	Category string   `json:"category"`
	Tags     []string `json:"tags"`
}

// ErrSlugTaken возвращает Store.Add, если товар с таким slug уже есть.
var ErrSlugTaken = errors.New("slug уже занят")

// Store хранит товары в памяти, ID выдаются по порядку с 1.
type Store struct {
	mu    sync.Mutex
	items []Product
}

// NewStore создаёт хранилище и добавляет в него товары (ID присваиваются заново).
func NewStore(items ...Product) *Store {
	s := &Store{}
	for _, p := range items {
		if _, err := s.Add(p); err != nil {
			panic(err)
		}
	}
	return s
}

// Add сохраняет товар, присваивает ему ID и возвращает сохранённую копию.
func (s *Store) Add(p Product) (Product, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, it := range s.items {
		if it.Slug == p.Slug {
			return Product{}, ErrSlugTaken
		}
	}
	p.ID = int64(len(s.items) + 1)
	p.Tags = slices.Clone(p.Tags)
	s.items = append(s.items, p)
	return p, nil
}

// All возвращает копию всех товаров в порядке добавления.
func (s *Store) All() []Product {
	s.mu.Lock()
	defer s.mu.Unlock()
	return slices.Clone(s.items)
}

// Get ищет товар по ID.
func (s *Store) Get(id int64) (Product, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, it := range s.items {
		if it.ID == id {
			return it, true
		}
	}
	return Product{}, false
}

// Единый формат ответа: успех {"data": ..., "meta": ...}, ошибка {"error": {...}}.

// ErrorBody описывает ошибку для клиента. Code машиночитаемый, Message для человека.
type ErrorBody struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

// Meta это сведения о странице списка.
type Meta struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// Коды ошибок API.
const (
	CodeBadRequest = "bad_request"       // тело или параметры не разбираются
	CodeValidation = "validation_failed" // разобрались, но не прошли проверки
	CodeNotFound   = "not_found"
	CodeConflict   = "conflict"
)
