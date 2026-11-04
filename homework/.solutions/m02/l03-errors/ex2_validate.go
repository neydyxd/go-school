package l03errors

import (
	"errors"
	"strings"
)

type FieldError struct {
	Field string
	Msg   string
}

func (e *FieldError) Error() string { return e.Field + ": " + e.Msg }

type User struct {
	Name  string
	Email string
	Age   int
}

func ValidateUser(u User) error {
	var errs []error
	if strings.TrimSpace(u.Name) == "" {
		errs = append(errs, &FieldError{"name", "обязательно"})
	}
	if !strings.Contains(u.Email, "@") {
		errs = append(errs, &FieldError{"email", "некорректный адрес"})
	}
	if u.Age < 18 {
		errs = append(errs, &FieldError{"age", "должно быть не меньше 18"})
	}
	return errors.Join(errs...)
}
