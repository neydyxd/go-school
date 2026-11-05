package l02binding

import (
	"errors"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

var slugRe = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// SetupValidator настраивает валидатор (в Gin это binding.Validator.Engine()):
//
//  1. Имена полей в ошибках берутся из json-тега (RegisterTagNameFunc): "email", а не "Email".
//     Часть после запятой (",omitempty") отбрасывается.
//  2. Кастомный тег slug (RegisterValidation): строка из строчных латинских букв и цифр,
//     группы разделены одиночными дефисами: "go-school", "a1", но не "Go", "-a", "a--b", "".
//
// Вернуть ошибку регистрации, если она случилась.
func SetupValidator(v *validator.Validate) error {
	v.RegisterTagNameFunc(func(f reflect.StructField) string {
		name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
		if name == "-" {
			return ""
		}
		return name
	})
	return v.RegisterValidation("slug", func(fl validator.FieldLevel) bool {
		return slugRe.MatchString(fl.Field().String())
	})
}

// FieldErrors переводит ошибки валидации в карту "поле: сообщение по-русски".
//
// Ключ это Namespace() ошибки без первого сегмента (имени структуры):
// "In.address.city" превращается в "address.city", "In.tags[1]" в "tags[1]".
// У анонимной структуры (var in struct{...}) имени типа в Namespace() нет: "name"
// так и остаётся "name", не теряй такой ключ.
//
// Сообщения по тегу (N это Param()):
//
//	required        обязательное поле
//	email           некорректный email
//	oneof           одно из: a, b, c         (значения Param() через запятую с пробелом)
//	slug            только строчные латинские буквы, цифры и дефисы
//	min             строка: минимальная длина N; слайс, массив, map: минимум элементов: N;
//	                число: не меньше N
//	max             строка: максимальная длина N; слайс, массив, map: максимум элементов: N;
//	                число: не больше N
//	gt, gte, lt, lte  больше N, не меньше N, меньше N, не больше N
//	любой другой    не прошло проверку <тег>
//
// Если err не содержит validator.ValidationErrors (в том числе обёрнутые через %w),
// вернуть nil.
func FieldErrors(err error) map[string]string {
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return nil
	}
	out := make(map[string]string, len(ve))
	for _, fe := range ve {
		key := fe.Namespace()
		if _, rest, ok := strings.Cut(key, "."); ok {
			key = rest
		}
		out[key] = message(fe)
	}
	return out
}

func message(fe validator.FieldError) string {
	p := fe.Param()
	switch fe.Tag() {
	case "required":
		return "обязательное поле"
	case "email":
		return "некорректный email"
	case "oneof":
		return "одно из: " + strings.Join(strings.Fields(p), ", ")
	case "slug":
		return "только строчные латинские буквы, цифры и дефисы"
	case "min", "max":
		minTag := fe.Tag() == "min"
		switch fe.Kind() {
		case reflect.String:
			if minTag {
				return "минимальная длина " + p
			}
			return "максимальная длина " + p
		case reflect.Slice, reflect.Array, reflect.Map:
			if minTag {
				return "минимум элементов: " + p
			}
			return "максимум элементов: " + p
		}
		if minTag {
			return "не меньше " + p
		}
		return "не больше " + p
	case "gt":
		return "больше " + p
	case "gte":
		return "не меньше " + p
	case "lt":
		return "меньше " + p
	case "lte":
		return "не больше " + p
	}
	return "не прошло проверку " + fe.Tag()
}
