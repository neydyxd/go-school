package l03jsonapi

import (
	"errors"
	"maps"
	"strings"
	"testing"
)

func intp(n int) *int { return &n }

func TestValidate(t *testing.T) {
	tests := []struct {
		name string
		in   CreateUser
		want FieldErrors // nil: ошибок нет
	}{
		{"всё верно", CreateUser{Email: "a@b.ru", Name: "Степан", Age: intp(30)}, nil},
		{"возраст не указан", CreateUser{Email: "a@b.ru", Name: "Ира"}, nil},
		{"возраст 0 допустим", CreateUser{Email: "a@b.ru", Name: "Ира", Age: intp(0)}, nil},
		{"возраст 150 допустим", CreateUser{Email: "a@b.ru", Name: "Ира", Age: intp(150)}, nil},
		{"пробелы вокруг допустимы", CreateUser{Email: "  a@b.ru ", Name: " Ира "}, nil},
		{"имя ровно 50 рун кириллицей", CreateUser{Email: "a@b.ru", Name: strings.Repeat("ё", 50)}, nil},
		{"поддомен", CreateUser{Email: "a.b+c@mail.example.com", Name: "x"}, nil},
		{"всё пустое", CreateUser{}, FieldErrors{"email": "required", "name": "required"}},
		{"только пробелы", CreateUser{Email: "   ", Name: "\t"}, FieldErrors{"email": "required", "name": "required"}},
		{"без собаки", CreateUser{Email: "ab.ru", Name: "x"}, FieldErrors{"email": "invalid"}},
		{"две собаки", CreateUser{Email: "a@b@c.ru", Name: "x"}, FieldErrors{"email": "invalid"}},
		{"пусто слева", CreateUser{Email: "@b.ru", Name: "x"}, FieldErrors{"email": "invalid"}},
		{"пусто справа", CreateUser{Email: "a@", Name: "x"}, FieldErrors{"email": "invalid"}},
		{"домен без точки", CreateUser{Email: "a@localhost", Name: "x"}, FieldErrors{"email": "invalid"}},
		{"точка первой в домене", CreateUser{Email: "a@.ru", Name: "x"}, FieldErrors{"email": "invalid"}},
		{"точка последней в домене", CreateUser{Email: "a@b.", Name: "x"}, FieldErrors{"email": "invalid"}},
		{"email ровно 254 байта", CreateUser{Email: strings.Repeat("a", 249) + "@b.ru", Name: "x"}, nil},
		{"email 255 байт", CreateUser{Email: strings.Repeat("a", 250) + "@b.ru", Name: "x"}, FieldErrors{"email": "too_long"}},
		{"имя 51 руна", CreateUser{Email: "a@b.ru", Name: strings.Repeat("ё", 51)}, FieldErrors{"name": "too_long"}},
		{"возраст -1", CreateUser{Email: "a@b.ru", Name: "x", Age: intp(-1)}, FieldErrors{"age": "out_of_range"}},
		{"возраст 151", CreateUser{Email: "a@b.ru", Name: "x", Age: intp(151)}, FieldErrors{"age": "out_of_range"}},
		{"все ошибки сразу", CreateUser{Email: "nope", Name: "", Age: intp(200)},
			FieldErrors{"email": "invalid", "name": "required", "age": "out_of_range"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.in.Validate()
			if tt.want == nil {
				if err != nil {
					t.Errorf("Validate(%+v) = %v (%#v), want nil", tt.in, err, err)
				}
				return
			}
			var fe FieldErrors
			if !errors.As(err, &fe) {
				t.Fatalf("Validate(%+v) = %v, want FieldErrors %v", tt.in, err, tt.want)
			}
			if !maps.Equal(fe, tt.want) {
				t.Errorf("Validate(%+v) = %v, want %v", tt.in, fe, tt.want)
			}
		})
	}
}
