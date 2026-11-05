package l05auth

import (
	"errors"
	"fmt"
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
	if len(secret) < MinSecretLen {
		return "", ErrShortSecret
	}
	claims := Claims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			Issuer:    Issuer,
			Audience:  jwt.ClaimStrings{Audience},
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
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
	var c Claims
	_, err := jwt.ParseWithClaims(token, &c,
		func(*jwt.Token) (any, error) { return secret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
		jwt.WithIssuer(Issuer),
		jwt.WithAudience(Audience),
		jwt.WithTimeFunc(func() time.Time { return now }),
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidToken, err)
	}
	if c.Subject == "" {
		return nil, fmt.Errorf("%w: пустой sub", ErrInvalidToken)
	}
	return &c, nil
}
