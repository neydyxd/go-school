package l05auth

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// MinCost это минимальная стоимость bcrypt, которую принимает HashPassword.
const MinCost = bcrypt.DefaultCost // 10

var (
	ErrEmptyPassword = errors.New("пустой пароль")
	ErrWeakCost      = errors.New("слишком маленький cost для bcrypt")
)

// HashPassword возвращает bcrypt-хеш пароля со стоимостью cost.
//
// Правила:
//   - пустой пароль: ErrEmptyPassword;
//   - cost меньше MinCost: ErrWeakCost;
//   - пароль длиннее 72 байт (байт, не символов!): ошибка, которая через errors.Is
//     совпадает с bcrypt.ErrPasswordTooLong. Молча обрезать пароль нельзя.
//
// Соль генерировать самому не нужно: bcrypt кладёт случайную соль прямо в хеш.
func HashPassword(password string, cost int) (string, error) {
	panic("TODO: реализуй HashPassword")
}

// CheckPassword сообщает, подходит ли пароль к хешу. Битый хеш, пустой хеш
// и неверный пароль дают false. Сравнивать строки через == нельзя: хеш
// каждый раз разный из-за соли, а сравнение должно идти за постоянное время.
func CheckPassword(hash, password string) bool {
	panic("TODO: реализуй CheckPassword")
}
