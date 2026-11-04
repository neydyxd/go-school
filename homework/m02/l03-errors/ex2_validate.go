package l03errors

// FieldError: нарушение правила в одном поле.
type FieldError struct {
	Field string
	Msg   string
}

func (e *FieldError) Error() string { return e.Field + ": " + e.Msg }

// User: входные данные регистрации.
type User struct {
	Name  string
	Email string
	Age   int
}

// ValidateUser проверяет все поля и возвращает ВСЕ нарушения сразу,
// объединённые через errors.Join. Каждое нарушение: *FieldError.
// Порядок проверок и тексты:
//
//	Name пустое после strings.TrimSpace -> {"name", "обязательно"}
//	Email без символа "@"               -> {"email", "некорректный адрес"}
//	Age меньше 18                       -> {"age", "должно быть не меньше 18"}
//
// Нет нарушений: настоящий nil.
func ValidateUser(u User) error {
	panic("TODO: реализуй ValidateUser")
}
