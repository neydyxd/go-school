package l04quality

import "sync"

// Hits считает обращения по ключу, безопасен для конкурентного использования.
type Hits struct {
	mu sync.Mutex
	m  map[string]int
}

func NewHits() *Hits {
	return &Hits{m: map[string]int{}}
}

func (h *Hits) Add(key string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.m[key]++
}

func (h *Hits) Get(key string) int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.m[key]
}

func (h *Hits) Total() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	total := 0
	for _, n := range h.m {
		total += n
	}
	return total
}
