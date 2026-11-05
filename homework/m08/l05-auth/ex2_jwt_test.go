package l05auth

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	testSecret = []byte("0123456789abcdef0123456789abcdef") // ровно 32 байта
	testNow    = time.Date(2026, 11, 5, 12, 0, 0, 0, time.UTC)
)

// sign подписывает произвольные claims, чтобы собирать «вредные» токены.
func sign(t *testing.T, m jwt.SigningMethod, key any, claims jwt.Claims) string {
	t.Helper()
	s, err := jwt.NewWithClaims(m, claims).SignedString(key)
	if err != nil {
		t.Fatalf("не удалось подписать тестовый токен: %v", err)
	}
	return s
}

func validClaims() Claims {
	return Claims{
		Role: "user",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "u1",
			Issuer:    Issuer,
			Audience:  jwt.ClaimStrings{Audience},
			IssuedAt:  jwt.NewNumericDate(testNow),
			NotBefore: jwt.NewNumericDate(testNow),
			ExpiresAt: jwt.NewNumericDate(testNow.Add(15 * time.Minute)),
		},
	}
}

func TestIssueAccess(t *testing.T) {
	tok, err := IssueAccess(testSecret, "u42", "admin", testNow, 15*time.Minute)
	if err != nil {
		t.Fatalf("IssueAccess: %v", err)
	}
	parts := strings.Split(tok, ".")
	if len(parts) != 3 {
		t.Fatalf("токен %q не похож на JWT: частей %d, want 3", tok, len(parts))
	}
	hdr, _ := base64.RawURLEncoding.DecodeString(parts[0])
	var h map[string]any
	if err := json.Unmarshal(hdr, &h); err != nil || h["alg"] != "HS256" {
		t.Errorf("заголовок токена = %s, want alg HS256", hdr)
	}

	// Разбираем «вручную» библиотекой, чтобы не зависеть от твоего ParseAccess.
	var c Claims
	_, err = jwt.ParseWithClaims(tok, &c, func(*jwt.Token) (any, error) { return testSecret, nil },
		jwt.WithTimeFunc(func() time.Time { return testNow }))
	if err != nil {
		t.Fatalf("выпущенный токен не проходит проверку: %v", err)
	}
	checks := []struct{ name, got, want string }{
		{"sub", c.Subject, "u42"},
		{"role", c.Role, "admin"},
		{"iss", c.Issuer, Issuer},
	}
	for _, ch := range checks {
		if ch.got != ch.want {
			t.Errorf("%s = %q, want %q", ch.name, ch.got, ch.want)
		}
	}
	if len(c.Audience) != 1 || c.Audience[0] != Audience {
		t.Errorf("aud = %v, want [%s]", c.Audience, Audience)
	}
	if c.ExpiresAt == nil || !c.ExpiresAt.Equal(testNow.Add(15*time.Minute)) {
		t.Errorf("exp = %v, want %v", c.ExpiresAt, testNow.Add(15*time.Minute))
	}
	if c.IssuedAt == nil || c.NotBefore == nil {
		t.Errorf("iat = %v, nbf = %v: оба должны быть заполнены", c.IssuedAt, c.NotBefore)
	}

	t.Run("короткий секрет", func(t *testing.T) {
		if _, err := IssueAccess([]byte("secret"), "u1", "user", testNow, time.Minute); !errors.Is(err, ErrShortSecret) {
			t.Errorf("IssueAccess(короткий секрет) err = %v, want ErrShortSecret", err)
		}
	})
}

func TestParseAccess(t *testing.T) {
	good := sign(t, jwt.SigningMethodHS256, testSecret, validClaims())

	c, err := ParseAccess(testSecret, good, testNow.Add(time.Minute))
	if err != nil {
		t.Fatalf("ParseAccess(валидный токен) = %v, want nil", err)
	}
	if c.Subject != "u1" || c.Role != "user" {
		t.Errorf("ParseAccess вернул sub=%q role=%q, want u1/user", c.Subject, c.Role)
	}

	without := func(edit func(*Claims)) Claims {
		c := validClaims()
		edit(&c)
		return c
	}
	noneToken := sign(t, jwt.SigningMethodNone, jwt.UnsafeAllowNoneSignatureType, validClaims())

	// Подменяем payload, оставляя старую подпись: роль user -> admin.
	parts := strings.Split(good, ".")
	evil := validClaims()
	evil.Role = "admin"
	evilPayload, _ := json.Marshal(evil)
	tampered := parts[0] + "." + base64.RawURLEncoding.EncodeToString(evilPayload) + "." + parts[2]

	tests := []struct {
		name  string
		token string
		now   time.Time
	}{
		{"просрочен", good, testNow.Add(16 * time.Minute)},
		{"ещё не действует (nbf в будущем)", good, testNow.Add(-time.Minute)},
		{"чужой секрет", sign(t, jwt.SigningMethodHS256, []byte("another-secret-another-secret-12"), validClaims()), testNow},
		{"alg=none", noneToken, testNow},
		{"HS512 тем же секретом (подмена алгоритма)", sign(t, jwt.SigningMethodHS512, testSecret, validClaims()), testNow},
		{"подменён payload", tampered, testNow},
		{"нет exp", sign(t, jwt.SigningMethodHS256, testSecret, without(func(c *Claims) { c.ExpiresAt = nil })), testNow},
		{"чужой iss", sign(t, jwt.SigningMethodHS256, testSecret, without(func(c *Claims) { c.Issuer = "evil" })), testNow},
		{"чужой aud", sign(t, jwt.SigningMethodHS256, testSecret, without(func(c *Claims) { c.Audience = jwt.ClaimStrings{"billing-api"} })), testNow},
		{"пустой sub", sign(t, jwt.SigningMethodHS256, testSecret, without(func(c *Claims) { c.Subject = "" })), testNow},
		{"iat в будущем", sign(t, jwt.SigningMethodHS256, testSecret, without(func(c *Claims) {
			c.IssuedAt = jwt.NewNumericDate(testNow.Add(time.Hour))
			c.NotBefore = nil
		})), testNow},
		{"мусор", "not.a.jwt", testNow},
		{"пустая строка", "", testNow},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := ParseAccess(testSecret, tt.token, tt.now)
			if err == nil {
				t.Fatalf("ParseAccess(%s) = %+v, nil; want ошибку", tt.name, c)
			}
			if !errors.Is(err, ErrInvalidToken) {
				t.Errorf("ParseAccess(%s) err = %v, want errors.Is(err, ErrInvalidToken)", tt.name, err)
			}
		})
	}
}
