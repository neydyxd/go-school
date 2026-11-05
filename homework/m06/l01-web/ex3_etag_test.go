package l01web

import "testing"

func TestNotModified(t *testing.T) {
	tests := []struct {
		name              string
		ifNoneMatch, etag string
		want              bool
	}{
		{"точное совпадение", `"v2"`, `"v2"`, true},
		{"другая версия", `"v1"`, `"v2"`, false},
		{"префикс не считается совпадением", `"v1"`, `"v10"`, false},
		{"пустой If-None-Match", "", `"v1"`, false},
		{"у ресурса нет ETag", `"v1"`, "", false},
		{"звёздочка и есть ETag", "*", `"abc"`, true},
		{"звёздочка без ETag", "*", "", false},
		{"список, совпадает второй", `"a", "b", "c"`, `"b"`, true},
		{"список без пробелов", `"a","b"`, `"b"`, true},
		{"список без совпадений", `"a", "b"`, `"z"`, false},
		{"слабый тег в запросе", `W/"v2"`, `"v2"`, true},
		{"слабый тег у ресурса", `"v2"`, `W/"v2"`, true},
		{"оба слабые", `"x", W/"v2"`, `W/"v2"`, true},
		{"кавычки входят в тег", `v2`, `"v2"`, false},
		{"регистр важен", `"ABC"`, `"abc"`, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NotModified(tt.ifNoneMatch, tt.etag); got != tt.want {
				t.Errorf("NotModified(%q, %q) = %v, want %v", tt.ifNoneMatch, tt.etag, got, tt.want)
			}
		})
	}
}
