package l06configlogsdocs

import (
	"fmt"
	"slices"
)

// Field описывает поле JSON-объекта в контракте API.
type Field struct {
	Type     string // "string", "integer", "boolean", ...
	Required bool   // в запросе: клиент обязан прислать; в ответе: сервер всегда присылает
}

// Schema это плоская схема JSON-объекта: имя поля -> описание.
type Schema map[string]Field

// Direction говорит, кто пишет объект, а кто читает.
type Direction int

const (
	Request  Direction = iota // пишет клиент, читает сервер
	Response                  // пишет сервер, читает клиент
)

// BreakingChanges сравнивает старую и новую версии схемы и возвращает список
// несовместимых изменений, то есть таких, что сломают уже работающих клиентов
// версии API без смены /v1 на /v2. Правило: «будь строг к тому, что обещаешь, и
// не требуй больше, чем требовал».
//
// Для Response (клиент читает то, что мы прислали):
//   - поле удалено: "<поле>: removed";
//   - поле было Required, стало необязательным: "<поле>: became optional".
//
// Для Request (сервер читает то, что прислал клиент; неизвестные поля сервер игнорирует):
//   - появилось новое обязательное поле: "<поле>: new required field";
//   - существующее необязательное поле стало обязательным: "<поле>: became required".
//
// Для обоих направлений: тип поля изменился: "<поле>: type changed string -> integer".
//
// Всё остальное совместимо: новое поле в ответе, новое необязательное поле в запросе,
// удаление поля из запроса, ослабление обязательности в запросе.
// Результат отсортирован по строкам (slices.Sort); нет изменений: nil или пустой слайс.
func BreakingChanges(dir Direction, old, new Schema) []string {
	var out []string
	for name, o := range old {
		n, ok := new[name]
		switch {
		case !ok:
			if dir == Response {
				out = append(out, name+": removed")
			}
		case o.Type != n.Type:
			out = append(out, fmt.Sprintf("%s: type changed %s -> %s", name, o.Type, n.Type))
		case dir == Response && o.Required && !n.Required:
			out = append(out, name+": became optional")
		case dir == Request && !o.Required && n.Required:
			out = append(out, name+": became required")
		}
	}
	if dir == Request {
		for name, n := range new {
			if _, existed := old[name]; !existed && n.Required {
				out = append(out, name+": new required field")
			}
		}
	}
	slices.Sort(out)
	return out
}
