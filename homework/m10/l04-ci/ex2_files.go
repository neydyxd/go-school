package l04ci

import "io"

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
	panic("TODO: реализуй WorstFiles")
}
