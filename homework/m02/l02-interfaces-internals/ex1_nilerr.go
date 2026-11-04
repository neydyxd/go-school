package l02interfacesinternals

// ValidationError: ошибка валидации поля.
type ValidationError struct {
	Field  string
	Reason string
}

func (e *ValidationError) Error() string { return e.Field + ": " + e.Reason }

// checkAge уже написана, менять её не нужно: для корректного возраста
// она возвращает nil-указатель типа *ValidationError.
func checkAge(age int) *ValidationError {
	if age < 0 || age > 150 {
		return &ValidationError{Field: "age", Reason: "вне диапазона 0-150"}
	}
	return nil
}

// ValidateAge проверяет возраст через checkAge и возвращает результат как error.
// Для корректного возраста вызывающий код должен получить настоящий nil:
// проверка err == nil обязана сработать. Подумай, почему return checkAge(age) не годится.
func ValidateAge(age int) error {
	panic("TODO: реализуй ValidateAge")
}
