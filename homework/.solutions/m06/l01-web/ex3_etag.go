package l01web

import "strings"

// NotModified решает, можно ли ответить 304 Not Modified на условный GET.
//
// ifNoneMatch: значение заголовка If-None-Match из запроса. Бывает пустым,
// "*" или списком тегов через запятую с необязательными пробелами:
// `"v1", W/"v2"`. etag: текущий ETag ресурса, например `"v2"` или `W/"v2"`;
// пустая строка значит, что у ресурса нет ETag.
//
// Правила (RFC 9110, раздел 13.1.2):
//   - пустой ifNoneMatch или пустой etag: false;
//   - "*" совпадает с любым ресурсом, у которого есть etag;
//   - для If-None-Match сравнение слабое: префикс W/ отбрасывается с обеих
//     сторон, а остальное (вместе с кавычками) сравнивается целиком.
//     `"v1"` не совпадает с `"v10"`.
func NotModified(ifNoneMatch, etag string) bool {
	if ifNoneMatch == "" || etag == "" {
		return false
	}
	if strings.TrimSpace(ifNoneMatch) == "*" {
		return true
	}
	want := strings.TrimPrefix(etag, "W/")
	for tag := range strings.SplitSeq(ifNoneMatch, ",") {
		if strings.TrimPrefix(strings.TrimSpace(tag), "W/") == want {
			return true
		}
	}
	return false
}
