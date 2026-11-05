package l04ci

import (
	"cmp"
	"io"
	"slices"
)

// FileCoverage: покрытие одного файла в процентах.
type FileCoverage struct {
	File    string
	Percent float64
}

// WorstFiles возвращает файлы с покрытием строго ниже threshold, от худшего к
// лучшему, при равенстве по имени файла. Правила разбора и слияния блоков
// как в Coverage. Файлы без операторов не попадают в ответ. Ничего ниже
// порога: пустой срез (не nil-ошибка). Ошибка разбора: как в Coverage.
//
// Такой список CI печатает рядом с «coverage 71% < 80%», чтобы было видно,
// где именно не хватает тестов.
func WorstFiles(r io.Reader, threshold float64) ([]FileCoverage, error) {
	blocks, err := parseProfile(r)
	if err != nil {
		return nil, err
	}
	type acc struct{ total, covered int }
	files := map[string]*acc{}
	for _, b := range blocks {
		a := files[b.file]
		if a == nil {
			a = &acc{}
			files[b.file] = a
		}
		a.total += b.stmts
		if b.covered {
			a.covered += b.stmts
		}
	}
	out := []FileCoverage{}
	for f, a := range files {
		if a.total == 0 {
			continue
		}
		p := 100 * float64(a.covered) / float64(a.total)
		if p < threshold {
			out = append(out, FileCoverage{f, p})
		}
	}
	slices.SortFunc(out, func(x, y FileCoverage) int {
		return cmp.Or(cmp.Compare(x.Percent, y.Percent), cmp.Compare(x.File, y.File))
	})
	return out, nil
}
