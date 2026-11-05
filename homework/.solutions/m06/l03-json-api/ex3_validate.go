package l03jsonapi

import (
	"strings"
	"unicode/utf8"
)

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
	fe := FieldErrors{}

	email := strings.TrimSpace(c.Email)
	switch {
	case email == "":
		fe["email"] = "required"
	case len(email) > 254:
		fe["email"] = "too_long"
	case !validEmail(email):
		fe["email"] = "invalid"
	}

	name := strings.TrimSpace(c.Name)
	switch {
	case name == "":
		fe["name"] = "required"
	case utf8.RuneCountInString(name) > 50:
		fe["name"] = "too_long"
	}

	if c.Age != nil && (*c.Age < 0 || *c.Age > 150) {
		fe["age"] = "out_of_range"
	}

	if len(fe) == 0 {
		return nil
	}
	return fe
}

func validEmail(s string) bool {
	local, domain, ok := strings.Cut(s, "@")
	if !ok || local == "" || strings.Contains(domain, "@") {
		return false
	}
	dot := strings.Index(domain, ".")
	return dot > 0 && !strings.HasSuffix(domain, ".")
}
