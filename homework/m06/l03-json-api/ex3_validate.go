package l03jsonapi

import "strings"

// FieldErrors: ошибки валидации по полям, ключ это имя поля в JSON.
// Значения: "required", "invalid", "too_long", "out_of_range".
type FieldErrors map[string]string

func (fe FieldErrors) Error() string {
	var b strings.Builder
	b.WriteString("validation failed:")
	for k, v := range fe {
		b.WriteString(" " + k + "=" + v)
	}
	return b.String()
}

// CreateUser: тело POST /users.
type CreateUser struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	Age   *int   `json:"age"` // необязательное: nil значит «не указал»
}

// Validate проверяет запрос и возвращает nil или FieldErrors со ВСЕМИ
// нарушениями сразу (клиенту удобнее исправить форму за один заход).
//
//   - email: пустой (после strings.TrimSpace) это "required";
//     длиннее 254 байт это "too_long";
//     иначе ровно один "@", непустые части слева и справа, в правой части
//     есть точка, но не первым и не последним символом. Нарушение: "invalid".
//   - name: после TrimSpace пустое это "required"; длиннее 50 СИМВОЛОВ
//     (рун, не байт: «Ёжик» это 4 символа) это "too_long".
//   - age: nil допустим; иначе 0..150 включительно, вне диапазона "out_of_range".
//
// Возвращай именно nil, когда ошибок нет: не пустую FieldErrors,
// иначе получишь ненулевой интерфейс error с пустой мапой внутри.
func (c CreateUser) Validate() error {
	panic("TODO: реализуй Validate")
}
