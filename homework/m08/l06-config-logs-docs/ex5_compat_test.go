package l06configlogsdocs

import (
	"slices"
	"testing"
)

func TestBreakingChanges(t *testing.T) {
	task := Schema{
		"id":     {"integer", true},
		"title":  {"string", true},
		"status": {"string", true},
		"due":    {"string", false},
	}
	with := func(base Schema, edit func(Schema)) Schema {
		s := Schema{}
		for k, v := range base {
			s[k] = v
		}
		edit(s)
		return s
	}
	tests := []struct {
		name string
		dir  Direction
		new  Schema
		want []string
	}{
		{"ответ без изменений", Response, task, nil},
		{"ответ: новое поле совместимо", Response, with(task, func(s Schema) { s["priority"] = Field{"integer", true} }), nil},
		{"ответ: удалили поле", Response, with(task, func(s Schema) { delete(s, "due") }), []string{"due: removed"}},
		{"ответ: id стал строкой", Response, with(task, func(s Schema) { s["id"] = Field{"string", true} }),
			[]string{"id: type changed integer -> string"}},
		{"ответ: обязательное стало необязательным", Response, with(task, func(s Schema) { s["status"] = Field{"string", false} }),
			[]string{"status: became optional"}},
		{"ответ: необязательное стало обязательным совместимо", Response, with(task, func(s Schema) { s["due"] = Field{"string", true} }), nil},
		{"ответ: несколько сразу, отсортировано", Response, with(task, func(s Schema) {
			delete(s, "title")
			s["due"] = Field{"integer", false}
			s["id"] = Field{"integer", false}
		}), []string{"due: type changed string -> integer", "id: became optional", "title: removed"}},

		{"запрос: новое необязательное поле совместимо", Request, with(task, func(s Schema) { s["priority"] = Field{"integer", false} }), nil},
		{"запрос: новое обязательное поле", Request, with(task, func(s Schema) { s["project_id"] = Field{"integer", true} }),
			[]string{"project_id: new required field"}},
		{"запрос: удалили поле совместимо", Request, with(task, func(s Schema) { delete(s, "due") }), nil},
		{"запрос: необязательное стало обязательным", Request, with(task, func(s Schema) { s["due"] = Field{"string", true} }),
			[]string{"due: became required"}},
		{"запрос: обязательное стало необязательным совместимо", Request, with(task, func(s Schema) { s["title"] = Field{"string", false} }), nil},
		{"запрос: смена типа", Request, with(task, func(s Schema) { s["status"] = Field{"integer", true} }),
			[]string{"status: type changed string -> integer"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BreakingChanges(tt.dir, task, tt.new)
			if len(got) == 0 && len(tt.want) == 0 {
				return
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("BreakingChanges(%s) = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}
