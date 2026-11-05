package l03jsonapi

import (
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
)

type signup struct {
	Email string `json:"email"`
	Age   int    `json:"age"`
}

func TestDecodeJSON(t *testing.T) {
	tests := []struct {
		name, ctype, body string
		wantStatus        int // 0: ошибки нет
		wantCode          string
		want              signup
	}{
		{"всё хорошо", "application/json", `{"email":"a@b.c","age":30}`, 0, "", signup{"a@b.c", 30}},
		{"charset в Content-Type", "application/json; charset=utf-8", `{"email":"x"}`, 0, "", signup{Email: "x"}},
		{"тип в другом регистре", "Application/JSON", `{"age":1}`, 0, "", signup{Age: 1}},
		{"пробелы и перевод строки после объекта", "application/json", "{\"age\":2}\n \n", 0, "", signup{Age: 2}},
		{"ровно на лимите", "application/json", `{"email":"` + strings.Repeat("a", 52) + `"}`, 0, "", signup{Email: strings.Repeat("a", 52)}},
		{"нет Content-Type", "", `{"age":1}`, 415, "unsupported_media_type", signup{}},
		{"text/plain", "text/plain", `{"age":1}`, 415, "unsupported_media_type", signup{}},
		{"json-подобный, но не тот тип", "application/jsonp", `{"age":1}`, 415, "unsupported_media_type", signup{}},
		{"тело больше лимита", "application/json", `{"email":"` + strings.Repeat("a", 53) + `"}`, 413, "body_too_large", signup{}},
		{"мусор за лимитом", "application/json", `{"age":1}` + strings.Repeat(" ", 100) + "x", 413, "body_too_large", signup{}},
		{"пустое тело", "application/json", "", 400, "empty_body", signup{}},
		{"битый JSON", "application/json", `{"email":`, 400, "invalid_json", signup{}},
		{"не тот тип поля", "application/json", `{"age":"сорок"}`, 400, "invalid_json", signup{}},
		{"дробное в int", "application/json", `{"age":1.5}`, 400, "invalid_json", signup{}},
		{"массив вместо объекта", "application/json", `[1,2]`, 400, "invalid_json", signup{}},
		{"неизвестное поле", "application/json", `{"emial":"a@b.c"}`, 400, "unknown_field", signup{}},
		{"два объекта подряд", "application/json", `{"age":1}{"age":2}`, 400, "invalid_json", signup{}},
		{"мусор после объекта", "application/json", `{"age":1}xyz`, 400, "invalid_json", signup{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("POST", "/signup", strings.NewReader(tt.body))
			if tt.ctype != "" {
				r.Header.Set("Content-Type", tt.ctype)
			}
			var got signup
			err := DecodeJSON(httptest.NewRecorder(), r, &got, 64)
			if tt.wantStatus == 0 {
				if err != nil {
					t.Fatalf("DecodeJSON(%q) = %v, want nil", tt.body, err)
				}
				if got != tt.want {
					t.Errorf("DecodeJSON(%q) записал %+v, want %+v", tt.body, got, tt.want)
				}
				return
			}
			var ae *APIError
			if !errors.As(err, &ae) {
				t.Fatalf("DecodeJSON(%q) = %v, want *APIError %d %s", tt.body, err, tt.wantStatus, tt.wantCode)
			}
			if ae.Status != tt.wantStatus || ae.Code != tt.wantCode {
				t.Errorf("DecodeJSON(%q) = %d %s, want %d %s", tt.body, ae.Status, ae.Code, tt.wantStatus, tt.wantCode)
			}
		})
	}
}
