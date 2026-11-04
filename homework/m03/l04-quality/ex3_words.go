package l04quality

import (
	"regexp"
	"strings"
)

// CountWords считает слова в тексте без учёта регистра. Слово это
// непрерывная серия букв и цифр любого алфавита; всё остальное разделители.
// CountWords("Go, go GO! Гофер") = map[go:3 гофер:1].
//
// Функция работает правильно, но медленно. Задание: найди узкое место
// профилировщиком и перепиши так, чтобы результат не изменился, а аллокаций
// стало в разы меньше (тест TestCountWordsAllocs ставит потолок 1000).
//
//	go test -run=^$ -bench=CountWords -benchmem -cpuprofile=cpu.out -memprofile=mem.out ./m03/l04-quality/
//	go tool pprof -top -cum cpu.out
//	go tool pprof -sample_index=alloc_space -list CountWords mem.out
//
// Потолок отсекает компиляцию регулярки на каждой строке (около 6000
// аллокаций) и разбор текста по строкам с FindAllString на каждой (около 1800).
// Регулярка в переменной пакета, применённая ко всему тексту сразу, уже
// уложится. Ещё быстрее strings.FieldsFunc с функцией «не буква и не число»
// (unicode.IsLetter, unicode.IsNumber: \p{L} и \p{N} в регулярке это ровно
// эти категории). Сравни варианты бенчмарком и benchstat.
//
// Под -race проверка аллокаций пропускается: детектор сбрасывает sync.Pool,
// на котором regexp держит свои машины, и счётчик завышается.
func CountWords(text string) map[string]int {
	counts := map[string]int{}
	for _, line := range strings.Split(text, "\n") {
		re := regexp.MustCompile(`[\p{L}\p{N}]+`)
		for _, w := range re.FindAllString(line, -1) {
			counts[strings.ToLower(w)]++
		}
	}
	return counts
}
