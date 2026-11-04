package l03errors

import (
	"errors"
	"slices"
	"testing"
)

func TestValidateUser(t *testing.T) {
	tests := []struct {
		name       string
		in         User
		wantFields []string
		wantMsg    string
	}{
		{"всё верно", User{"Аня", "anya@example.com", 30}, nil, ""},
		{"ровно 18", User{"Аня", "a@b", 18}, nil, ""},
		{"пустое имя", User{"", "a@b", 20}, []string{"name"}, "name: обязательно"},
		{"имя из пробелов", User{"   ", "a@b", 20}, []string{"name"}, "name: обязательно"},
		{"плохой email", User{"Боря", "boris.example.com", 20}, []string{"email"}, "email: некорректный адрес"},
		{"всё плохо", User{"", "", 17}, []string{"name", "email", "age"},
			"name: обязательно\nemail: некорректный адрес\nage: должно быть не меньше 18"},
		{"имя и возраст", User{"", "a@b", 0}, []string{"name", "age"},
			"name: обязательно\nage: должно быть не меньше 18"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUser(tt.in)
			if tt.wantFields == nil {
				if err != nil {
					t.Fatalf("ValidateUser(%+v) = %#v, want nil", tt.in, err)
				}
				return
			}
			if err == nil {
				t.Fatalf("ValidateUser(%+v) = nil, want ошибки %v", tt.in, tt.wantFields)
			}
			if err.Error() != tt.wantMsg {
				t.Errorf("ValidateUser(%+v).Error() = %q, want %q", tt.in, err.Error(), tt.wantMsg)
			}
			var fe *FieldError
			if !errors.As(err, &fe) || fe.Field != tt.wantFields[0] {
				t.Errorf("errors.As(err, *FieldError) дал %v, want первое поле %q", fe, tt.wantFields[0])
			}
			multi, ok := err.(interface{ Unwrap() []error })
			if !ok {
				t.Fatalf("ValidateUser(%+v) вернула %T, want результат errors.Join", tt.in, err)
			}
			var fields []string
			for _, e := range multi.Unwrap() {
				var fe *FieldError
				if !errors.As(e, &fe) {
					t.Fatalf("внутри Join лежит %T, want *FieldError", e)
				}
				fields = append(fields, fe.Field)
			}
			if !slices.Equal(fields, tt.wantFields) {
				t.Errorf("поля с ошибками = %v, want %v", fields, tt.wantFields)
			}
		})
	}
}
