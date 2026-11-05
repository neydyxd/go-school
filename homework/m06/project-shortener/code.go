// Package shortener: ядро URL-shortener на стандартной библиотеке
// (мини-проект модуля 06). Сервер (main, конфиг, graceful shutdown)
// собираешь сам по ТЗ из урока, а здесь логика, которую он вызывает.
package shortener

import "errors"

// ErrInvalidURL: ссылку нельзя сокращать.
var ErrInvalidURL = errors.New("invalid url")

// Alphabet: символы короткого кода (base62).
const Alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// NewCode возвращает случайный код длины n из символов Alphabet.
// Источник случайности: crypto/rand (math/rand предсказуем, а по
// предсказуемым кодам чужие ссылки перебираются). Для n <= 0 вернуть "".
func NewCode(n int) string {
	panic("TODO: реализуй NewCode")
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
	panic("TODO: реализуй ValidateURL")
}
