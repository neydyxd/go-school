// Package service это слой бизнес-логики: правила домена, без HTTP и без SQL.
package service

import (
	"context"

	"goschool/homework/m08/l04-architecture/library"
)

// BookRepository описан здесь, на стороне потребителя: сервису нужны ровно эти три метода.
// Реализация живёт в пакете postgres, в тестах её подменяет fake в памяти.
type BookRepository interface {
	// Create сохраняет книгу и возвращает её с присвоенным ID. ISBN занят: library.ErrISBNTaken.
	Create(ctx context.Context, b library.Book) (library.Book, error)
	// Get возвращает книгу. Нет такой: library.ErrNotFound.
	Get(ctx context.Context, id int64) (library.Book, error)
	// TakeCopy атомарно уменьшает Copies на 1 и возвращает книгу после изменения.
	// Нет книги: library.ErrNotFound; экземпляров 0: library.ErrNoCopies.
	TakeCopy(ctx context.Context, id int64) (library.Book, error)
}

// Service реализует сценарии библиотеки поверх репозитория.
type Service struct {
	repo BookRepository
}

// New собирает сервис. Зависимость передаётся явно: никаких глобальных переменных и DI-контейнеров.
func New(repo BookRepository) *Service {
	return &Service{repo: repo}
}

// Create проверяет и нормализует данные, затем сохраняет книгу через репозиторий.
// Проверки идут по порядку, на первой ошибке вернуть *library.ValidationError и в репозиторий не ходить:
//
//	title   пробелы по краям обрезать; пусто: "обязательное поле"; длиннее 200 символов (рун): "максимальная длина 200"
//	author  пробелы по краям обрезать; пусто: "обязательное поле"
//	isbn    убрать дефисы и пробелы; должно остаться ровно 13 цифр, иначе "ISBN это 13 цифр"
//	copies  от 0 до 1000, иначе "от 0 до 1000"
//
// Ошибки репозитория вернуть как есть или обёрнутыми через %w: вызывающий проверяет errors.Is.
// ctx передать в репозиторий без изменений.
func (s *Service) Create(ctx context.Context, title, author, isbn string, copies int) (library.Book, error) {
	panic("TODO: реализуй Service.Create")
}

// Get возвращает книгу по ID. id < 1 означает library.ErrNotFound без похода в репозиторий.
func (s *Service) Get(ctx context.Context, id int64) (library.Book, error) {
	panic("TODO: реализуй Service.Get")
}

// Take выдаёт один экземпляр книги (repo.TakeCopy). id < 1: library.ErrNotFound без похода в репозиторий.
func (s *Service) Take(ctx context.Context, id int64) (library.Book, error) {
	panic("TODO: реализуй Service.Take")
}
