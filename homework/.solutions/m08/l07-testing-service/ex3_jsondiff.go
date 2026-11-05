package l07testingservice

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
)

// JSONDiff сравнивает два JSON-документа по смыслу, а не побайтно: порядок ключей
// и пробелы не важны, числа сравниваются как числа (1 и 1.0 равны).
// Ключи объектов из ignore (на любой глубине, в обоих документах) выбрасываются
// до сравнения: так в ответах API не мешают id, created_at и прочие нестабильные поля.
//
// Невалидный JSON в got или want: ошибка. Документы равны: "".
// Иначе описание ПЕРВОГО различия при обходе в глубину, где ключи объекта
// перебираются по возрастанию (объединение ключей обеих сторон), а элементы
// массива по порядку. Путь начинается с "$", ключ добавляется как ".key", индекс как "[i]":
//
//	"$.user.name: missing"          ключ есть в want, нет в got
//	"$.debug: unexpected"           ключ есть в got, нет в want
//	"$.items: length 2, want 3"     массивы разной длины (проверяется до элементов)
//	"$.items[1].title: got \"a\", want \"b\""   прочее: значения через json.Marshal
func JSONDiff(got, want []byte, ignore ...string) (string, error) {
	var g, w any
	if err := json.Unmarshal(got, &g); err != nil {
		return "", fmt.Errorf("got: %w", err)
	}
	if err := json.Unmarshal(want, &w); err != nil {
		return "", fmt.Errorf("want: %w", err)
	}
	return diff("$", strip(g, ignore), strip(w, ignore)), nil
}

func strip(v any, ignore []string) any {
	switch x := v.(type) {
	case map[string]any:
		for k, child := range x {
			if slices.Contains(ignore, k) {
				delete(x, k)
			} else {
				x[k] = strip(child, ignore)
			}
		}
	case []any:
		for i := range x {
			x[i] = strip(x[i], ignore)
		}
	}
	return v
}

func diff(path string, g, w any) string {
	gm, gIsMap := g.(map[string]any)
	wm, wIsMap := w.(map[string]any)
	if gIsMap && wIsMap {
		keys := slices.Sorted(maps.Keys(gm))
		for k := range wm {
			if _, ok := gm[k]; !ok {
				keys = append(keys, k)
			}
		}
		slices.Sort(keys)
		for _, k := range keys {
			gv, inG := gm[k]
			wv, inW := wm[k]
			switch {
			case !inG:
				return path + "." + k + ": missing"
			case !inW:
				return path + "." + k + ": unexpected"
			}
			if d := diff(path+"."+k, gv, wv); d != "" {
				return d
			}
		}
		return ""
	}
	ga, gIsArr := g.([]any)
	wa, wIsArr := w.([]any)
	if gIsArr && wIsArr {
		if len(ga) != len(wa) {
			return fmt.Sprintf("%s: length %d, want %d", path, len(ga), len(wa))
		}
		for i := range ga {
			if d := diff(fmt.Sprintf("%s[%d]", path, i), ga[i], wa[i]); d != "" {
				return d
			}
		}
		return ""
	}
	gj, _ := json.Marshal(g)
	wj, _ := json.Marshal(w)
	if string(gj) != string(wj) {
		return fmt.Sprintf("%s: got %s, want %s", path, gj, wj)
	}
	return ""
}
