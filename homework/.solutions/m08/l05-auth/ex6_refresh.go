package l05auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"sync"
	"time"
)

var (
	ErrInvalidRefresh = errors.New("refresh-токен недействителен")
	ErrRefreshReused  = errors.New("refresh-токен использован повторно")
)

type refreshRecord struct {
	userID  string
	family  string // все токены одной цепочки ротации (одно устройство)
	expires time.Time
	used    bool
}

// RefreshStore хранит refresh-токены. Ключ карты byHash: hashToken(token).
// Сам токен в хранилище не попадает: утечка базы не должна давать готовые токены.
type RefreshStore struct {
	mu     sync.Mutex
	byHash map[string]*refreshRecord
	ttl    time.Duration
	now    func() time.Time
}

// NewRefreshStore готов, менять не нужно.
func NewRefreshStore(ttl time.Duration, now func() time.Time) *RefreshStore {
	return &RefreshStore{byHash: make(map[string]*refreshRecord), ttl: ttl, now: now}
}

// hashToken готов: hex(SHA-256(token)). Медленный хеш вроде bcrypt тут не нужен,
// у токена 256 бит случайности, перебирать нечего.
func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func newOpaqueToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b) // crypto/rand.Read не возвращает ошибку с Go 1.24
	return base64.RawURLEncoding.EncodeToString(b)
}

// Issue выпускает новый refresh-токен для userID (вход с нового устройства,
// то есть новое семейство). Токен: 32 байта из crypto/rand в base64.RawURLEncoding
// (43 символа). Срок жизни: now()+ttl.
func (s *RefreshStore) Issue(userID string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.issueLocked(userID, newOpaqueToken()), nil
}

func (s *RefreshStore) issueLocked(userID, family string) string {
	tok := newOpaqueToken()
	s.byHash[hashToken(tok)] = &refreshRecord{userID: userID, family: family, expires: s.now().Add(s.ttl)}
	return tok
}

// Rotate меняет refresh-токен на новый (ротация) и возвращает владельца.
//
//   - токен неизвестен или просрочен (now() не раньше expires): ErrInvalidRefresh;
//   - токен уже был использован: это кража (украденный токен и настоящий клиент
//     пришли оба). Отозвать (удалить) ВСЕ токены его семейства и вернуть ErrRefreshReused;
//   - иначе пометить старый использованным и выпустить новый в том же семействе.
//
// Два одновременных Rotate одного токена: успешен ровно один.
func (s *RefreshStore) Rotate(token string) (newToken, userID string, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	rec, ok := s.byHash[hashToken(token)]
	if !ok || !s.now().Before(rec.expires) {
		return "", "", ErrInvalidRefresh
	}
	if rec.used {
		for h, r := range s.byHash {
			if r.family == rec.family {
				delete(s.byHash, h)
			}
		}
		return "", "", ErrRefreshReused
	}
	rec.used = true
	return s.issueLocked(rec.userID, rec.family), rec.userID, nil
}

// RevokeUser удаляет все refresh-токены пользователя («выйти на всех устройствах»).
func (s *RefreshStore) RevokeUser(userID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for h, r := range s.byHash {
		if r.userID == userID {
			delete(s.byHash, h)
		}
	}
}
