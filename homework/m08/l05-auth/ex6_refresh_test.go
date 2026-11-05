package l05auth

import (
	"encoding/base64"
	"errors"
	"sync"
	"testing"
	"time"
)

func newStore(now *time.Time) *RefreshStore {
	return NewRefreshStore(30*24*time.Hour, func() time.Time { return *now })
}

func mustIssue(t *testing.T, s *RefreshStore, user string) string {
	t.Helper()
	tok, err := s.Issue(user)
	if err != nil {
		t.Fatalf("Issue(%q): %v", user, err)
	}
	return tok
}

func TestRefreshIssue(t *testing.T) {
	now := testNow
	s := newStore(&now)
	a := mustIssue(t, s, "u1")
	b := mustIssue(t, s, "u1")
	if raw, err := base64.RawURLEncoding.DecodeString(a); err != nil || len(raw) != 32 {
		t.Errorf("токен %q: want 32 случайных байта в base64.RawURLEncoding", a)
	}
	if a == b {
		t.Errorf("два Issue вернули одинаковый токен %q", a)
	}
	if _, plain := s.byHash[a]; plain {
		t.Errorf("токен лежит в хранилище открытым текстом: ключ должен быть hashToken(token)")
	}
	if _, ok := s.byHash[hashToken(a)]; !ok {
		t.Errorf("в byHash нет ключа hashToken(token)")
	}
	for k := range s.byHash {
		if k == a || k == b {
			t.Errorf("в byHash найден сам токен")
		}
	}
}

func TestRefreshRotate(t *testing.T) {
	now := testNow
	s := newStore(&now)
	t1 := mustIssue(t, s, "u1")
	laptop := mustIssue(t, s, "u1") // второе устройство, отдельное семейство

	t2, user, err := s.Rotate(t1)
	if err != nil || user != "u1" || t2 == "" || t2 == t1 {
		t.Fatalf("Rotate(t1) = %q, %q, %v; want новый токен, \"u1\", nil", t2, user, err)
	}

	// Злоумышленник украл t1 и пришёл с ним после настоящего клиента.
	if _, _, err := s.Rotate(t1); !errors.Is(err, ErrRefreshReused) {
		t.Fatalf("повторный Rotate(t1) err = %v, want ErrRefreshReused", err)
	}
	// Всё семейство отозвано: и выданный на смену t2 больше не работает.
	if _, _, err := s.Rotate(t2); !errors.Is(err, ErrInvalidRefresh) {
		t.Errorf("Rotate(t2) после обнаружения кражи err = %v, want ErrInvalidRefresh", err)
	}
	// Другое устройство того же пользователя не задето.
	if _, _, err := s.Rotate(laptop); err != nil {
		t.Errorf("Rotate(токен второго устройства) err = %v, want nil", err)
	}
	if _, _, err := s.Rotate("never-issued"); !errors.Is(err, ErrInvalidRefresh) {
		t.Errorf("Rotate(неизвестный) err = %v, want ErrInvalidRefresh", err)
	}
}

func TestRefreshExpiry(t *testing.T) {
	now := testNow
	s := newStore(&now)
	tok := mustIssue(t, s, "u1")
	now = now.Add(30*24*time.Hour - time.Second)
	fresh, _, err := s.Rotate(tok)
	if err != nil {
		t.Fatalf("Rotate за секунду до истечения: %v", err)
	}
	now = now.Add(30 * 24 * time.Hour) // ровно ttl после выдачи fresh
	if _, _, err := s.Rotate(fresh); !errors.Is(err, ErrInvalidRefresh) {
		t.Errorf("Rotate просроченного err = %v, want ErrInvalidRefresh", err)
	}
}

func TestRefreshRevokeUser(t *testing.T) {
	now := testNow
	s := newStore(&now)
	a := mustIssue(t, s, "u1")
	b := mustIssue(t, s, "u1")
	other := mustIssue(t, s, "u2")
	s.RevokeUser("u1")
	for _, tok := range []string{a, b} {
		if _, _, err := s.Rotate(tok); !errors.Is(err, ErrInvalidRefresh) {
			t.Errorf("Rotate после RevokeUser err = %v, want ErrInvalidRefresh", err)
		}
	}
	if _, user, err := s.Rotate(other); err != nil || user != "u2" {
		t.Errorf("токен другого пользователя: %q, %v; want u2, nil", user, err)
	}
}

func TestRefreshRotateConcurrent(t *testing.T) {
	now := testNow
	s := newStore(&now)
	tok := mustIssue(t, s, "u1")
	var mu sync.Mutex
	ok := 0
	var wg sync.WaitGroup
	for range 20 {
		wg.Go(func() {
			if _, _, err := s.Rotate(tok); err == nil {
				mu.Lock()
				ok++
				mu.Unlock()
			}
		})
	}
	wg.Wait()
	if ok != 1 {
		t.Errorf("20 одновременных Rotate одного токена: успешных %d, want 1", ok)
	}
}
