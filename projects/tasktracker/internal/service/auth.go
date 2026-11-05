package service

import (
	"context"
	"fmt"
	"net/mail"
	"strings"
	"time"
)

// UserRepo это хранилище пользователей. Реализация: repository.Users.
type UserRepo interface {
	// Create сохраняет пользователя и возвращает его с ID. Email занят: ErrConflict.
	Create(ctx context.Context, u User) (User, error)
	// GetByEmail: нет такого: ErrNotFound.
	GetByEmail(ctx context.Context, email string) (User, error)
}

// RefreshRepo хранит хеши refresh-токенов. Реализация: repository.RefreshTokens.
type RefreshRepo interface {
	Save(ctx context.Context, hash string, userID int64, family string, expires time.Time) error
	// MarkUsed атомарно помечает токен использованным и возвращает его запись.
	// Уже был использован: вернуть запись и used=true (это кража, отзывай семейство).
	MarkUsed(ctx context.Context, hash string, now time.Time) (userID int64, family string, used bool, err error)
	RevokeFamily(ctx context.Context, family string) error
	RevokeUser(ctx context.Context, userID int64) error
}

// PasswordHasher и TokenIssuer: узкие интерфейсы на то, что нужно сервису от пакета auth.
type PasswordHasher interface {
	Hash(password string) (string, error)
	Check(hash, password string) bool
}

type TokenIssuer interface {
	IssueAccess(userID int64, role string) (string, error)
}

type AuthService struct {
	users      UserRepo
	refresh    RefreshRepo
	hasher     PasswordHasher
	tokens     TokenIssuer
	accessTTL  time.Duration
	refreshTTL time.Duration
	now        func() time.Time
}

func NewAuthService(users UserRepo, refresh RefreshRepo, hasher PasswordHasher, tokens TokenIssuer, accessTTL, refreshTTL time.Duration) *AuthService {
	return &AuthService{users: users, refresh: refresh, hasher: hasher, tokens: tokens,
		accessTTL: accessTTL, refreshTTL: refreshTTL, now: time.Now}
}

// Register: ГОТОВЫЙ ОБРАЗЕЦ. Посмотри, как устроены валидация, нормализация,
// хеширование и проброс ошибок, и пиши остальные методы так же.
func (s *AuthService) Register(ctx context.Context, email, password string) (User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if a, err := mail.ParseAddress(email); err != nil || a.Address != email {
		return User{}, invalid("email", "некорректный email")
	}
	if n := len(password); n < 8 || n > 72 { // 72 байта: предел bcrypt
		return User{}, invalid("password", "от 8 до 72 байт")
	}
	hash, err := s.hasher.Hash(password)
	if err != nil {
		return User{}, fmt.Errorf("хеш пароля: %w", err)
	}
	u, err := s.users.Create(ctx, User{Email: email, PasswordHash: hash, Role: "user"})
	if err != nil {
		return User{}, fmt.Errorf("создать пользователя: %w", err)
	}
	return u, nil
}

// Login проверяет email и пароль и выдаёт пару токенов (новое семейство refresh).
//
// TODO:
//   - нет пользователя ИЛИ пароль не подходит: одинаковая ErrUnauthorized
//     (разные ошибки позволяют перебором узнать, какие email зарегистрированы);
//   - и одинаковое ВРЕМЯ ответа: если email не найден, всё равно вызови
//     s.hasher.Check с фиктивным хешем того же cost (посчитай его один раз,
//     например в NewAuthService). Иначе ответ без bcrypt приходит на ~60 мс
//     быстрее, и существование email видно по таймингу;
//   - пароль сверяй через s.hasher.Check, не через ==;
//   - refresh: auth.NewRefreshToken(), в базу только хеш, family: новый случайный id.
func (s *AuthService) Login(ctx context.Context, email, password string) (TokenPair, error) {
	return TokenPair{}, ErrNotImplemented
}

// Refresh меняет refresh-токен на новую пару (ротация).
//
// TODO: MarkUsed; не найден или просрочен: ErrUnauthorized; used=true: RevokeFamily
// и ErrUnauthorized; иначе новая пара в том же семействе.
func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (TokenPair, error) {
	return TokenPair{}, ErrNotImplemented
}

// Logout отзывает все refresh-токены пользователя. TODO.
func (s *AuthService) Logout(ctx context.Context, userID int64) error {
	return ErrNotImplemented
}
