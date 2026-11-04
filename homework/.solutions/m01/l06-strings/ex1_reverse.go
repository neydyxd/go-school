package l06strings

// Reverse переворачивает строку посимвольно (по рунам, а не по байтам).
// Reverse("привет") == "тевирп", Reverse("Go🙂") == "🙂oG".
// Составные символы (буква плюс отдельный диакритический знак) не учитывай:
// одна руна = один символ.
func Reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}
