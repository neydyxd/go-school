package l03testabledesign

import "errors"

// User это пользователь сервиса.
type User struct {
	ID    int
	Email string
}

var (
	ErrNotFound     = errors.New("not found")
	ErrExists       = errors.New("already exists")
	ErrInvalidEmail = errors.New("invalid email")
)

// Store это хранилище пользователей. В проде за ним стоит PostgreSQL,
// в тестах подставляется MemStore.
type Store interface {
	ByEmail(email string) (User, error)
	Save(u User) (User, error)
}

// MemStore это фейк: настоящая, но упрощённая реализация Store в памяти.
//
// Требования:
//   - NewMemStore возвращает пустое хранилище, готовое к работе;
//   - Save присваивает ID по порядку, начиная с 1, и возвращает сохранённого
//     пользователя; ID во входном u игнорируется;
//   - Save с уже занятым email возвращает ошибку ErrExists (можно обёрнутую);
//   - ByEmail возвращает ErrNotFound (можно обёрнутую), если email не найден.
type MemStore struct {
	// TODO: поля
}

func NewMemStore() *MemStore {
	panic("TODO: реализуй NewMemStore")
}

func (s *MemStore) ByEmail(email string) (User, error) {
	panic("TODO: реализуй ByEmail")
}

func (s *MemStore) Save(u User) (User, error) {
	panic("TODO: реализуй Save")
}

// Register регистрирует пользователя.
//
// Требования:
//   - email нормализуется: strings.TrimSpace и strings.ToLower;
//   - валидный email: ровно один "@", слева и справа от него непустые части,
//     иначе ErrInvalidEmail, а в хранилище никто не ходит;
//   - если ByEmail нашёл пользователя, вернуть ErrExists;
//   - если ByEmail вернул ErrNotFound, сохранить через Save;
//   - любую другую ошибку хранилища вернуть обёрнутой через %w (её проверят errors.Is).
func Register(s Store, email string) (User, error) {
	panic("TODO: реализуй Register")
}
