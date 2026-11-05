package l01goquestions

import (
	"errors"
	"time"
)

var (
	// ErrInvalidURL: baseURL не абсолютный http(s) URL с хостом.
	ErrInvalidURL = errors.New("некорректный base URL")
	// ErrInvalidOption: опция получила недопустимое значение.
	ErrInvalidOption = errors.New("некорректная опция")
)

// Client: настройки HTTP-клиента. Поля открыты только для чтения в тестах.
type Client struct {
	BaseURL string
	Timeout time.Duration
	Retries int
	Headers map[string]string
}

// Option меняет настройки клиента или возвращает ошибку, если значение недопустимо.
type Option func(*Client) error

// WithTimeout задаёт таймаут. d <= 0 даёт ошибку с ErrInvalidOption.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) error { return nil }
}

// WithRetries задаёт число повторов, допустимо 0..10. Иначе ошибка с ErrInvalidOption.
func WithRetries(n int) Option {
	return func(c *Client) error { return nil }
}

// WithHeader добавляет заголовок, который клиент будет отправлять. Пустое имя
// даёт ошибку с ErrInvalidOption. Повторный WithHeader с тем же именем перезаписывает значение.
func WithHeader(name, value string) Option {
	return func(c *Client) error { return nil }
}

// NewClient создаёт клиента.
//   - baseURL обязан быть абсолютным URL со схемой http или https и непустым хостом,
//     иначе ошибка с ErrInvalidURL.
//   - Значения по умолчанию: Timeout 10s, Retries 2, Headers пустая (не nil) мапа.
//   - Опции применяются по порядку, поздняя перекрывает раннюю. Первая ошибка
//     опции возвращается (errors.Is(err, ErrInvalidOption)), клиент при этом nil.
//   - Клиенты не делят состояние: изменение Headers одного клиента не видно другим,
//     даже если они созданы с одним и тем же набором опций.
func NewClient(baseURL string, opts ...Option) (*Client, error) {
	panic("TODO: реализуй NewClient")
}
