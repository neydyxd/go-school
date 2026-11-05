package l01goquestions

import (
	"errors"
	"fmt"
	"net/url"
	"time"
)

var (
	// ErrInvalidURL: baseURL не абсолютный http(s) URL с хостом.
	ErrInvalidURL = errors.New("некорректный base URL")
	// ErrInvalidOption: опция получила недопустимое значение.
	ErrInvalidOption = errors.New("некорректная опция")
)

// Client: настройки HTTP-клиента.
type Client struct {
	BaseURL string
	Timeout time.Duration
	Retries int
	Headers map[string]string
}

// Option меняет настройки клиента или возвращает ошибку.
type Option func(*Client) error

// WithTimeout задаёт таймаут.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) error {
		if d <= 0 {
			return fmt.Errorf("%w: таймаут %v должен быть больше нуля", ErrInvalidOption, d)
		}
		c.Timeout = d
		return nil
	}
}

// WithRetries задаёт число повторов 0..10.
func WithRetries(n int) Option {
	return func(c *Client) error {
		if n < 0 || n > 10 {
			return fmt.Errorf("%w: повторов %d, допустимо 0..10", ErrInvalidOption, n)
		}
		c.Retries = n
		return nil
	}
}

// WithHeader добавляет заголовок.
func WithHeader(name, value string) Option {
	return func(c *Client) error {
		if name == "" {
			return fmt.Errorf("%w: пустое имя заголовка", ErrInvalidOption)
		}
		c.Headers[name] = value // мапа своя у каждого клиента: создаётся в NewClient
		return nil
	}
}

// NewClient создаёт клиента со значениями по умолчанию и применяет опции.
func NewClient(baseURL string, opts ...Option) (*Client, error) {
	u, err := url.Parse(baseURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return nil, fmt.Errorf("%w: %q", ErrInvalidURL, baseURL)
	}
	c := &Client{BaseURL: baseURL, Timeout: 10 * time.Second, Retries: 2, Headers: map[string]string{}}
	for _, o := range opts {
		if err := o(c); err != nil {
			return nil, err
		}
	}
	return c, nil
}
