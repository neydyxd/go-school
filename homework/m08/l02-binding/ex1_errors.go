package l02binding

import "github.com/go-playground/validator/v10"

// SetupValidator настраивает валидатор (в Gin это binding.Validator.Engine()):
//
//  1. Имена полей в ошибках берутся из json-тега (RegisterTagNameFunc): "email", а не "Email".
//     Часть после запятой (",omitempty") отбрасывается.
//  2. Кастомный тег slug (RegisterValidation): строка из строчных латинских букв и цифр,
//     группы разделены одиночными дефисами: "go-school", "a1", но не "Go", "-a", "a--b", "".
//
// Вернуть ошибку регистрации, если она случилась.
func SetupValidator(v *validator.Validate) error {
	panic("TODO: реализуй SetupValidator")
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
	panic("TODO: реализуй FieldErrors")
}
