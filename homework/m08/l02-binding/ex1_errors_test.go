package l02binding

import (
	"errors"
	"fmt"
	"maps"
	"testing"

	"github.com/go-playground/validator/v10"
)

// newValidator настраивает валидатор так же, как это делает Gin: теги читаются из binding.
func newValidator(t *testing.T) *validator.Validate {
	t.Helper()
	v := validator.New()
	v.SetTagName("binding")
	if err := SetupValidator(v); err != nil {
		t.Fatalf("SetupValidator: %v", err)
	}
	return v
}

type address struct {
	City string `json:"city" binding:"required"`
}

type signup struct {
	Name    string   `json:"name" binding:"required,min=2,max=10"`
	Email   string   `json:"email,omitempty" binding:"required,email"`
	Age     int      `json:"age" binding:"gte=18,lte=120"`
	Role    string   `json:"role" binding:"oneof=admin user guest"`
	Slug    string   `json:"slug" binding:"omitempty,slug"`
	Tags    []string `json:"tags" binding:"max=3,dive,min=2"`
	Score   float64  `json:"score" binding:"gt=0,lt=10"`
	Items   []int    `json:"items" binding:"min=1"`
	Count   int      `json:"count" binding:"min=1,max=5"`
	Code    string   `json:"code" binding:"omitempty,uuid"`
	Address address  `json:"address"`
}

func valid() signup {
	return signup{
		Name: "Ян", Email: "yan@example.com", Age: 30, Role: "user", Slug: "go-school",
		Tags: []string{"go"}, Score: 5, Items: []int{1}, Count: 3, Address: address{City: "Казань"},
	}
}

func TestFieldErrors(t *testing.T) {
	v := newValidator(t)
	tests := []struct {
		name   string
		modify func(*signup)
		want   map[string]string
	}{
		{"всё верно", func(*signup) {}, nil},
		{"пустое обязательное", func(s *signup) { s.Name = "" }, map[string]string{"name": "обязательное поле"}},
		{"json-имя без omitempty", func(s *signup) { s.Email = "" }, map[string]string{"email": "обязательное поле"}},
		{"email", func(s *signup) { s.Email = "yan@" }, map[string]string{"email": "некорректный email"}},
		{"min у строки", func(s *signup) { s.Name = "Я" }, map[string]string{"name": "минимальная длина 2"}},
		{"max у строки", func(s *signup) { s.Name = "Константинопольский" }, map[string]string{"name": "максимальная длина 10"}},
		{"min у слайса", func(s *signup) { s.Items = nil }, map[string]string{"items": "минимум элементов: 1"}},
		{"max у слайса", func(s *signup) { s.Tags = []string{"aa", "bb", "cc", "dd"} }, map[string]string{"tags": "максимум элементов: 3"}},
		{"min у числа", func(s *signup) { s.Count = 0 }, map[string]string{"count": "не меньше 1"}},
		{"max у числа", func(s *signup) { s.Count = 6 }, map[string]string{"count": "не больше 5"}},
		{"gte", func(s *signup) { s.Age = 17 }, map[string]string{"age": "не меньше 18"}},
		{"lte", func(s *signup) { s.Age = 121 }, map[string]string{"age": "не больше 120"}},
		{"gt", func(s *signup) { s.Score = 0 }, map[string]string{"score": "больше 0"}},
		{"lt", func(s *signup) { s.Score = 10 }, map[string]string{"score": "меньше 10"}},
		{"oneof", func(s *signup) { s.Role = "root" }, map[string]string{"role": "одно из: admin, user, guest"}},
		{"slug", func(s *signup) { s.Slug = "Go School" }, map[string]string{"slug": "только строчные латинские буквы, цифры и дефисы"}},
		{"неизвестный тег", func(s *signup) { s.Code = "123" }, map[string]string{"code": "не прошло проверку uuid"}},
		{"dive по элементам", func(s *signup) { s.Tags = []string{"go", "x", "db", "y"}[:3] }, map[string]string{"tags[1]": "минимальная длина 2"}},
		{"вложенная структура", func(s *signup) { s.Address.City = "" }, map[string]string{"address.city": "обязательное поле"}},
		{"несколько ошибок сразу", func(s *signup) { s.Name = ""; s.Age = 5; s.Role = "" }, map[string]string{
			"name": "обязательное поле", "age": "не меньше 18", "role": "одно из: admin, user, guest",
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := valid()
			tt.modify(&s)
			err := v.Struct(s)
			got := FieldErrors(err)
			if !maps.Equal(got, tt.want) {
				t.Errorf("FieldErrors(%v) = %v, want %v", err, got, tt.want)
			}
		})
	}
}

func TestFieldErrorsNotValidation(t *testing.T) {
	v := newValidator(t)
	s := valid()
	s.Name = ""
	wrapped := fmt.Errorf("создание пользователя: %w", v.Struct(s))
	if got, want := FieldErrors(wrapped), map[string]string{"name": "обязательное поле"}; !maps.Equal(got, want) {
		t.Errorf("FieldErrors(обёрнутая ошибка) = %v, want %v", got, want)
	}
	for _, err := range []error{nil, errors.New("unexpected EOF")} {
		if got := FieldErrors(err); got != nil {
			t.Errorf("FieldErrors(%v) = %v, want nil", err, got)
		}
	}
}

func TestSlug(t *testing.T) {
	v := newValidator(t)
	tests := []struct {
		slug string
		ok   bool
	}{
		{"go", true}, {"go-school", true}, {"a1-b2-c3", true}, {"2026", true},
		{"", false}, {"Go", false}, {"go_school", false}, {"-go", false}, {"go-", false},
		{"go--school", false}, {"го", false}, {"go school", false},
	}
	for _, tt := range tests {
		err := v.Var(tt.slug, "slug")
		if (err == nil) != tt.ok {
			t.Errorf("slug %q: ошибка %v, want ok=%v", tt.slug, err, tt.ok)
		}
	}
}

func TestFieldErrorsAnonymousStruct(t *testing.T) {
	// Анонимная структура прямо в обработчике: в Namespace() нет имени типа.
	v := newValidator(t)
	in := struct {
		Name string `json:"name" binding:"min=4"`
		Age  int    `json:"age" binding:"gte=18"`
	}{Name: "Ян", Age: 5}
	got := FieldErrors(v.Struct(in))
	if want := map[string]string{"name": "минимальная длина 4", "age": "не меньше 18"}; !maps.Equal(got, want) {
		t.Errorf("FieldErrors(анонимная структура) = %v, want %v", got, want)
	}
}
