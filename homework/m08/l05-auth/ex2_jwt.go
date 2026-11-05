package l05auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	Issuer   = "tasktracker"
	Audience = "tasktracker-api"
	// MinSecretLen: секрет HS256 короче 32 байт перебирается офлайн по одному перехваченному токену.
	MinSecretLen = 32
)

var (
	ErrInvalidToken = errors.New("невалидный токен")
	ErrShortSecret  = errors.New("секрет короче 32 байт")
)

// Claims это полезная нагрузка access-токена: стандартные поля плюс роль.
type Claims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

// IssueAccess выпускает access-токен, подписанный HS256 секретом secret.
//
// В токене должны быть: sub=userID, role, iss=Issuer, aud=[Audience],
// iat=now, nbf=now, exp=now+ttl. Секрет короче MinSecretLen байт: ErrShortSecret.
func IssueAccess(secret []byte, userID, role string, now time.Time, ttl time.Duration) (string, error) {
	panic("TODO: реализуй IssueAccess")
}

// ParseAccess проверяет токен и возвращает его claims. Время проверки: now.
//
// Токен принимается, только если одновременно:
//   - алгоритм ровно HS256 (не none, не HS512, не RS256), подпись верна;
//   - exp есть и ещё не наступил, nbf и iat не в будущем;
//   - iss равен Issuer, aud содержит Audience, sub не пустой.
//
// Любая проблема: ошибка, которая через errors.Is совпадает с ErrInvalidToken.
func ParseAccess(secret []byte, token string, now time.Time) (*Claims, error) {
	panic("TODO: реализуй ParseAccess")
}
