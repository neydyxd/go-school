package l07testingservice

import "testing"

func TestJSONDiff(t *testing.T) {
	tests := []struct {
		name      string
		got, want string
		ignore    []string
		wantDiff  string
	}{
		{"порядок ключей и пробелы не важны", `{"b":1,"a":"x"}`, "{\n  \"a\": \"x\",\n  \"b\": 1\n}", nil, ""},
		{"1 и 1.0 равны", `{"n":1}`, `{"n":1.0}`, nil, ""},
		{"разные скаляры", `{"title":"a"}`, `{"title":"b"}`, nil, `$.title: got "a", want "b"`},
		{"число против строки", `{"id":1}`, `{"id":"1"}`, nil, `$.id: got 1, want "1"`},
		{"нет ключа", `{"a":1}`, `{"a":1,"b":2}`, nil, "$.b: missing"},
		{"лишний ключ", `{"a":1,"debug":true}`, `{"a":1}`, nil, "$.debug: unexpected"},
		{"первое различие по алфавиту", `{"z":1,"b":1}`, `{"z":2,"b":2}`, nil, "$.b: got 1, want 2"},
		{"missing раньше по алфавиту, чем отличие", `{"m":1}`, `{"a":1,"m":2}`, nil, "$.a: missing"},
		{"длина массива", `{"items":[1,2]}`, `{"items":[1,2,3]}`, nil, "$.items: length 2, want 3"},
		{"элемент массива", `{"items":[{"title":"a"},{"title":"a"}]}`, `{"items":[{"title":"a"},{"title":"b"}]}`, nil,
			`$.items[1].title: got "a", want "b"`},
		{"порядок массива важен", `[1,2]`, `[2,1]`, nil, "$[0]: got 1, want 2"},
		{"null против объекта", `{"user":null}`, `{"user":{"id":1}}`, nil, `$.user: got null, want {"id":1}`},
		{"ignore на любой глубине", `{"id":5,"items":[{"id":9,"t":"x","created_at":"2026-11-05T10:00:00Z"}]}`,
			`{"id":1,"items":[{"id":2,"t":"x","created_at":"2000-01-01T00:00:00Z"}]}`, []string{"id", "created_at"}, ""},
		{"ignore не скрывает настоящих различий", `{"id":5,"t":"x"}`, `{"id":1,"t":"y"}`, []string{"id"}, `$.t: got "x", want "y"`},
		{"вложенный объект целиком", `{"a":{"b":{"c":true}}}`, `{"a":{"b":{"c":false}}}`, nil, "$.a.b.c: got true, want false"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := JSONDiff([]byte(tt.got), []byte(tt.want), tt.ignore...)
			if err != nil {
				t.Fatalf("JSONDiff err = %v", err)
			}
			if got != tt.wantDiff {
				t.Errorf("JSONDiff(%s, %s, %v) = %q, want %q", tt.got, tt.want, tt.ignore, got, tt.wantDiff)
			}
		})
	}

	for _, bad := range [][2]string{{`{"a":`, `{}`}, {`{}`, `nope`}} {
		if _, err := JSONDiff([]byte(bad[0]), []byte(bad[1])); err == nil {
			t.Errorf("JSONDiff(%q, %q) err = nil, want ошибку разбора", bad[0], bad[1])
		}
	}
}
