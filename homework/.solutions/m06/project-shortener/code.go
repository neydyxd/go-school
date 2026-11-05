// Package shortener: ядро URL-shortener на стандартной библиотеке
// (мини-проект модуля 06). Сервер (main, конфиг, graceful shutdown)
// собираешь сам по ТЗ из урока, а здесь логика, которую он вызывает.
package shortener

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"net/url"
)

// ErrInvalidURL: ссылку нельзя сокращать.
var ErrInvalidURL = errors.New("invalid url")

// Alphabet: символы короткого кода (base62).
const Alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// NewCode возвращает случайный код длины n из символов Alphabet.
// Источник случайности: crypto/rand (math/rand предсказуем, а по
// предсказуемым кодам чужие ссылки перебираются). Для n <= 0 вернуть "".
func NewCode(n int) string {
	if n <= 0 {
		return ""
	}
	b := make([]byte, n)
	limit := big.NewInt(int64(len(Alphabet)))
	for i := range b {
		// rand.Int даёт равномерное распределение без перекоса от взятия по модулю
		k, err := rand.Int(rand.Reader, limit)
		if err != nil {
			panic(err) // crypto/rand.Reader с Go 1.24 не возвращает ошибок
		}
		b[i] = Alphabet[k.Int64()]
	}
	return string(b)
}

// ValidateURL проверяет ссылку перед сокращением. Ошибка оборачивает
// ErrInvalidURL (errors.Is) и объясняет причину.
//
//   - не длиннее 2048 байт;
//   - разбирается url.Parse;
//   - схема только http или https (никаких javascript:, ftp:, file:);
//   - есть хост (u.Host не пустой);
//   - нет userinfo: "https://bank.ru@evil.example" выглядит как ссылка
//     на банк, а ведёт на evil.example.
func ValidateURL(raw string) error {
	if len(raw) > 2048 {
		return fmt.Errorf("%w: длиннее 2048 байт", ErrInvalidURL)
	}
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidURL, err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("%w: схема %q, нужна http или https", ErrInvalidURL, u.Scheme)
	}
	if u.Host == "" {
		return fmt.Errorf("%w: нет хоста", ErrInvalidURL)
	}
	if u.User != nil {
		return fmt.Errorf("%w: userinfo в ссылке запрещено", ErrInvalidURL)
	}
	return nil
}
