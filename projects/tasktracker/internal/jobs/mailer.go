package jobs

import (
	"context"
	"log/slog"
	"sync"
)

// Email это одно письмо. IdempotencyKey передаётся провайдеру (у SendGrid,
// Postmark, Unisender он есть в API): повтор с тем же ключом не шлёт письмо второй раз.
type Email struct {
	To             string
	Subject        string
	Body           string
	IdempotencyKey string
}

// Mailer отправляет письма. Интерфейс объявлен здесь, у потребителя.
type Mailer interface {
	Send(ctx context.Context, e Email) error
}

// LogMailer «отправляет» письма в лог. Для локального запуска. Готово.
type LogMailer struct{ Log *slog.Logger }

func (m LogMailer) Send(ctx context.Context, e Email) error {
	m.Log.InfoContext(ctx, "письмо", "to", e.To, "subject", e.Subject, "key", e.IdempotencyKey)
	return nil
}

// FakeMailer запоминает письма для тестов, дубли по IdempotencyKey не
// записывает (как настоящий провайдер). Err, если задан, возвращается из Send. Готово.
type FakeMailer struct {
	mu   sync.Mutex
	sent []Email
	keys map[string]bool
	Err  error
}

func (m *FakeMailer) Send(_ context.Context, e Email) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.Err != nil {
		return m.Err
	}
	if m.keys == nil {
		m.keys = map[string]bool{}
	}
	if e.IdempotencyKey != "" && m.keys[e.IdempotencyKey] {
		return nil
	}
	m.keys[e.IdempotencyKey] = true
	m.sent = append(m.sent, e)
	return nil
}

// Sent возвращает копию отправленных писем.
func (m *FakeMailer) Sent() []Email {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]Email(nil), m.sent...)
}
