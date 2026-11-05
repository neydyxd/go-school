package l01web

import (
	"errors"
	"reflect"
	"testing"
)

func TestParseRequestOK(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want Request
	}{
		{
			"GET без тела",
			"GET /api/links?limit=10 HTTP/1.1\r\nHost: example.com\r\nAccept: */*\r\n\r\n",
			Request{Method: "GET", Target: "/api/links?limit=10", Proto: "HTTP/1.1",
				Header: map[string][]string{"Host": {"example.com"}, "Accept": {"*/*"}}},
		},
		{
			"POST с телом",
			"POST /api/links HTTP/1.1\r\nHost: s.test\r\nContent-Type: application/json\r\nContent-Length: 13\r\n\r\n{\"url\":\"x\"}\r\n",
			Request{Method: "POST", Target: "/api/links", Proto: "HTTP/1.1",
				Header: map[string][]string{"Host": {"s.test"}, "Content-Type": {"application/json"}, "Content-Length": {"13"}},
				Body:   []byte("{\"url\":\"x\"}\r\n")},
		},
		{
			"имена в разном регистре и пробелы вокруг значения",
			"GET / HTTP/1.1\r\nhost:   a.test \t\r\ncontent-TYPE:text/plain\r\nx-request-id: 7\r\n\r\n",
			Request{Method: "GET", Target: "/", Proto: "HTTP/1.1",
				Header: map[string][]string{"Host": {"a.test"}, "Content-Type": {"text/plain"}, "X-Request-Id": {"7"}}},
		},
		{
			"повторный заголовок",
			"GET / HTTP/1.1\r\nHost: a.test\r\nAccept: text/html\r\naccept: application/json\r\n\r\n",
			Request{Method: "GET", Target: "/", Proto: "HTTP/1.1",
				Header: map[string][]string{"Host": {"a.test"}, "Accept": {"text/html", "application/json"}}},
		},
		{
			"Content-Length в байтах, а не в символах",
			"POST /n HTTP/1.1\r\nHost: a.test\r\nContent-Length: 12\r\n\r\nпривет",
			Request{Method: "POST", Target: "/n", Proto: "HTTP/1.1",
				Header: map[string][]string{"Host": {"a.test"}, "Content-Length": {"12"}},
				Body:   []byte("привет")},
		},
		{
			"в теле тоже есть пустая строка",
			"PUT /f HTTP/1.1\r\nHost: a.test\r\nContent-Length: 8\r\n\r\na\r\n\r\nbcd",
			Request{Method: "PUT", Target: "/f", Proto: "HTTP/1.1",
				Header: map[string][]string{"Host": {"a.test"}, "Content-Length": {"8"}},
				Body:   []byte("a\r\n\r\nbcd")},
		},
		{
			"лишние байты после тела не трогаем",
			"POST /a HTTP/1.1\r\nHost: a.test\r\nContent-Length: 2\r\n\r\nokGET /b HTTP/1.1\r\n",
			Request{Method: "POST", Target: "/a", Proto: "HTTP/1.1",
				Header: map[string][]string{"Host": {"a.test"}, "Content-Length": {"2"}},
				Body:   []byte("ok")},
		},
		{
			"HTTP/1.0 без Host",
			"GET /old HTTP/1.0\r\n\r\n",
			Request{Method: "GET", Target: "/old", Proto: "HTTP/1.0", Header: map[string][]string{}},
		},
		{
			"одинаковые Content-Length допустимы",
			"POST / HTTP/1.1\r\nHost: a\r\nContent-Length: 1\r\nContent-Length: 1\r\n\r\nz",
			Request{Method: "POST", Target: "/", Proto: "HTTP/1.1",
				Header: map[string][]string{"Host": {"a"}, "Content-Length": {"1", "1"}},
				Body:   []byte("z")},
		},
		{
			"значение с двоеточием",
			"GET / HTTP/1.1\r\nHost: localhost:8080\r\n\r\n",
			Request{Method: "GET", Target: "/", Proto: "HTTP/1.1",
				Header: map[string][]string{"Host": {"localhost:8080"}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRequest(tt.raw)
			if err != nil {
				t.Fatalf("ParseRequest(%q) error = %v, want nil", tt.raw, err)
			}
			if got.Header == nil {
				t.Fatalf("ParseRequest(%q).Header = nil, want непустую или пустую, но не nil мапу", tt.raw)
			}
			if len(got.Body) == 0 && len(tt.want.Body) == 0 {
				got.Body, tt.want.Body = nil, nil
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseRequest(%q) =\n %+v\nwant\n %+v", tt.raw, got, tt.want)
			}
		})
	}
}

func TestParseRequestErrors(t *testing.T) {
	tests := []struct {
		name string
		raw  string
	}{
		{"пустой ввод", ""},
		{"нет пустой строки в конце заголовков", "GET / HTTP/1.1\r\nHost: a\r\n"},
		{"голые LF вместо CRLF", "GET / HTTP/1.1\nHost: a\n\n"},
		{"две части в стартовой строке", "GET /\r\nHost: a\r\n\r\n"},
		{"двойной пробел в стартовой строке", "GET  / HTTP/1.1\r\nHost: a\r\n\r\n"},
		{"метод строчными", "get / HTTP/1.1\r\nHost: a\r\n\r\n"},
		{"метод с цифрой", "G3T / HTTP/1.1\r\nHost: a\r\n\r\n"},
		{"цель без слеша", "GET api HTTP/1.1\r\nHost: a\r\n\r\n"},
		{"неизвестный протокол", "GET / HTTP/2.0\r\nHost: a\r\n\r\n"},
		{"нет Host в HTTP/1.1", "GET / HTTP/1.1\r\nAccept: */*\r\n\r\n"},
		{"два Host", "GET / HTTP/1.1\r\nHost: a\r\nHost: b\r\n\r\n"},
		{"два Host в HTTP/1.0", "GET / HTTP/1.0\r\nHost: a\r\nhost: b\r\n\r\n"},
		{"пробел перед двоеточием", "GET / HTTP/1.1\r\nHost : a\r\n\r\n"},
		{"строка заголовка без двоеточия", "GET / HTTP/1.1\r\nHost: a\r\nbroken\r\n\r\n"},
		{"пустое имя заголовка", "GET / HTTP/1.1\r\nHost: a\r\n: x\r\n\r\n"},
		{"Transfer-Encoding", "POST / HTTP/1.1\r\nHost: a\r\nTransfer-Encoding: chunked\r\n\r\n0\r\n\r\n"},
		{"transfer-encoding строчными", "POST / HTTP/1.1\r\nHost: a\r\ntransfer-encoding: chunked\r\n\r\n0\r\n\r\n"},
		{"таб перед двоеточием", "GET / HTTP/1.1\r\nHost: a\r\nX-Api\t: b\r\n\r\n"},
		{"Content-Length не число", "POST / HTTP/1.1\r\nHost: a\r\nContent-Length: abc\r\n\r\n"},
		{"Content-Length отрицательный", "POST / HTTP/1.1\r\nHost: a\r\nContent-Length: -1\r\n\r\n"},
		{"Content-Length с плюсом", "POST / HTTP/1.1\r\nHost: a\r\nContent-Length: +1\r\n\r\nx"},
		{"разные Content-Length", "POST / HTTP/1.1\r\nHost: a\r\nContent-Length: 1\r\nContent-Length: 2\r\n\r\nxy"},
		{"тело короче Content-Length", "POST / HTTP/1.1\r\nHost: a\r\nContent-Length: 10\r\n\r\nshort"},
		{"кириллица: символов 6, а байтов меньше 13", "POST / HTTP/1.1\r\nHost: a\r\nContent-Length: 13\r\n\r\nпривет"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRequest(tt.raw)
			if !errors.Is(err, ErrMalformed) {
				t.Errorf("ParseRequest(%q) = %+v, %v; want ошибку, оборачивающую ErrMalformed", tt.raw, got, err)
			}
		})
	}
}
