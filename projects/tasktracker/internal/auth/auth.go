// Package auth: пароли, access-токены (JWT) и refresh-токены.
package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidToken   = errors.New("invalid token")
	ErrNotImplemented = errors.New("TODO: не реализовано")
)

// ---- Пароли: готово ----

// Hasher хеширует пароли bcrypt. В проде Cost не меньше bcrypt.DefaultCost,
// в тестах можно bcrypt.MinCost, чтобы не ждать.
type Hasher struct{ Cost int }

func (h Hasher) Hash(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), h.Cost)
	return string(b), err
}

func (h Hasher) Check(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// ---- Access-токены: TODO ----

// Claims это то, что сервис узнаёт из валидного access-токена.
type Claims struct {
	UserID int64
	Role   string
}

// Tokens выпускает и проверяет access-токены (JWT, HS256).
type Tokens struct {
	secret    []byte
	accessTTL time.Duration
	now       func() time.Time
}

func NewTokens(secret []byte, accessTTL time.Duration) *Tokens {
	return &Tokens{secret: secret, accessTTL: accessTTL, now: time.Now}
}

// IssueAccess выпускает JWT: sub = userID строкой, role, iss "tasktracker",
// aud "tasktracker-api", iat, nbf, exp = now + accessTTL. Алгоритм HS256.
//
// TODO: реализуй через github.com/golang-jwt/jwt/v5 (как в домашке m08/l05, ex2).
// Зависимость добавь сам: go get github.com/golang-jwt/jwt/v5@v5.3.1
func (t *Tokens) IssueAccess(userID int64, role string) (string, error) {
	return "", ErrNotImplemented
}

// ParseAccess проверяет подпись, алгоритм (только HS256), exp, iss, aud
// и возвращает Claims. Любая проблема: ошибка, совпадающая с ErrInvalidToken.
//
// TODO: реализуй. Не забудь jwt.WithValidMethods и jwt.WithExpirationRequired.
func (t *Tokens) ParseAccess(token string) (Claims, error) {
	return Claims{}, ErrNotImplemented
}

// ---- Refresh-токены: готово ----

// NewRefreshToken возвращает случайный токен для клиента и его хеш для базы.
func NewRefreshToken() (plain, hash string) {
	b := make([]byte, 32)
	_, _ = rand.Read(b) // с Go 1.24 crypto/rand.Read не возвращает ошибку
	plain = base64.RawURLEncoding.EncodeToString(b)
	return plain, HashRefresh(plain)
}

// HashRefresh: SHA-256 в hex. Медленный хеш не нужен, в токене 256 бит случайности.
func HashRefresh(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(sum[:])
}
