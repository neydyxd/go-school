package l06configlogsdocs

// Задание ex6: допиши контракт API задач в ex6_openapi.json (OpenAPI 3.x, JSON).
// Тест проверяет не стиль, а то, что обязан знать клиент:
//   - версия openapi "3.x";
//   - схема безопасности http bearer с bearerFormat JWT, применённая ко всем операциям
//     (глобальный security или security у каждой операции);
//   - операции GET и POST /v1/tasks, GET, PATCH и DELETE /v1/tasks/{id};
//   - у каждой операции есть ответ 401; у операций с {id} ещё 404 и параметр id
//     (in: path, required: true) на уровне пути или операции;
//   - GET /v1/tasks принимает query-параметры status (enum todo, in_progress, done)
//     и limit (с maximum);
//   - POST: обязательное тело application/json и ответ 201 со схемой Task;
//   - DELETE: ответ 204;
//   - components.schemas.Task: свойства id, title, status, все три в required;
//   - тело ответа 401 описано схемой с полем error (единый формат ошибок).
// Ссылки $ref на "#/components/..." разрешены и приветствуются.

import (
	"encoding/json"
	"os"
	"slices"
	"strings"
	"testing"
)

type node = map[string]any

func loadSpec(t *testing.T) node {
	t.Helper()
	b, err := os.ReadFile("ex6_openapi.json")
	if err != nil {
		t.Fatalf("не прочитать ex6_openapi.json: %v", err)
	}
	var doc node
	if err := json.Unmarshal(b, &doc); err != nil {
		t.Fatalf("ex6_openapi.json не валидный JSON: %v", err)
	}
	return doc
}

// get идёт по пути ключей, по дороге разрешая $ref.
func get(doc node, v any, path ...string) any {
	v = deref(doc, v)
	for _, k := range path {
		m, ok := v.(node)
		if !ok {
			return nil
		}
		v = deref(doc, m[k])
	}
	return v
}

func deref(doc node, v any) any {
	for range 10 { // защита от циклических ссылок
		m, ok := v.(node)
		if !ok {
			return v
		}
		ref, ok := m["$ref"].(string)
		if !ok || !strings.HasPrefix(ref, "#/") {
			return v
		}
		v = get(doc, doc, strings.Split(strings.TrimPrefix(ref, "#/"), "/")...)
	}
	return v
}

func params(doc node, pathItem, op any) []node {
	var out []node
	for _, src := range []any{get(doc, pathItem, "parameters"), get(doc, op, "parameters")} {
		list, _ := src.([]any)
		for _, p := range list {
			if m, ok := deref(doc, p).(node); ok {
				out = append(out, m)
			}
		}
	}
	return out
}

func findParam(ps []node, name, in string) node {
	for _, p := range ps {
		if p["name"] == name && p["in"] == in {
			return p
		}
	}
	return nil
}

func TestOpenAPISpec(t *testing.T) {
	doc := loadSpec(t)

	if v, _ := doc["openapi"].(string); !strings.HasPrefix(v, "3.") {
		t.Errorf("openapi = %q, want 3.x", v)
	}

	bearer := ""
	if schemes, ok := get(doc, doc, "components", "securitySchemes").(node); ok {
		for name, s := range schemes {
			s, _ := deref(doc, s).(node)
			sc, _ := s["scheme"].(string)
			if s["type"] == "http" && strings.EqualFold(sc, "bearer") && s["bearerFormat"] == "JWT" {
				bearer = name
			}
		}
	}
	if bearer == "" {
		t.Fatalf("в components.securitySchemes нет схемы {type: http, scheme: bearer, bearerFormat: JWT}")
	}
	usesBearer := func(sec any) bool {
		list, _ := sec.([]any)
		for _, req := range list {
			if m, ok := req.(node); ok {
				if _, ok := m[bearer]; ok {
					return true
				}
			}
		}
		return false
	}
	globalSec := usesBearer(doc["security"])

	ops := []struct{ path, method string }{
		{"/v1/tasks", "get"}, {"/v1/tasks", "post"},
		{"/v1/tasks/{id}", "get"}, {"/v1/tasks/{id}", "patch"}, {"/v1/tasks/{id}", "delete"},
	}
	for _, o := range ops {
		name := strings.ToUpper(o.method) + " " + o.path
		pathItem := get(doc, doc, "paths", o.path)
		op, ok := get(doc, pathItem, o.method).(node)
		if !ok {
			t.Errorf("%s: операция не описана", name)
			continue
		}
		if !globalSec && !usesBearer(op["security"]) {
			t.Errorf("%s: не защищена схемой %s (нет ни глобального security, ни своего)", name, bearer)
		}
		if get(doc, op, "responses", "401") == nil {
			t.Errorf("%s: нет ответа 401", name)
		} else if get(doc, op, "responses", "401", "content", "application/json", "schema", "properties", "error") == nil {
			t.Errorf("%s: ответ 401 не описан схемой с полем error", name)
		}
		if strings.Contains(o.path, "{id}") {
			if get(doc, op, "responses", "404") == nil {
				t.Errorf("%s: нет ответа 404", name)
			}
			p := findParam(params(doc, pathItem, op), "id", "path")
			if p == nil || p["required"] != true {
				t.Errorf("%s: нет параметра id (in: path, required: true)", name)
			}
		}
	}

	list := get(doc, doc, "paths", "/v1/tasks", "get")
	ps := params(doc, get(doc, doc, "paths", "/v1/tasks"), list)
	if p := findParam(ps, "status", "query"); p == nil {
		t.Errorf("GET /v1/tasks: нет query-параметра status")
	} else {
		enum, _ := get(doc, p, "schema", "enum").([]any)
		var got []string
		for _, e := range enum {
			s, _ := e.(string)
			got = append(got, s)
		}
		slices.Sort(got)
		if !slices.Equal(got, []string{"done", "in_progress", "todo"}) {
			t.Errorf("status enum = %v, want todo, in_progress, done", got)
		}
	}
	if p := findParam(ps, "limit", "query"); p == nil || get(doc, p, "schema", "maximum") == nil {
		t.Errorf("GET /v1/tasks: нет query-параметра limit с maximum (защита от limit=1000000)")
	}

	post := get(doc, doc, "paths", "/v1/tasks", "post")
	if get(doc, post, "requestBody", "required") != true || get(doc, post, "requestBody", "content", "application/json", "schema") == nil {
		t.Errorf("POST /v1/tasks: нужно обязательное тело application/json со схемой")
	}
	task, _ := get(doc, doc, "components", "schemas", "Task").(node)
	if task == nil {
		t.Fatalf("нет components.schemas.Task")
	}
	created, _ := get(doc, post, "responses", "201", "content", "application/json", "schema").(node)
	if created == nil || created["properties"] == nil || get(doc, created, "properties", "title") == nil {
		t.Errorf("POST /v1/tasks: ответ 201 должен отдавать Task")
	}
	if get(doc, doc, "paths", "/v1/tasks/{id}", "delete", "responses", "204") == nil {
		t.Errorf("DELETE /v1/tasks/{id}: нет ответа 204")
	}
	req, _ := task["required"].([]any)
	for _, f := range []string{"id", "title", "status"} {
		if get(doc, task, "properties", f) == nil {
			t.Errorf("Task: нет свойства %s", f)
		}
		if !slices.Contains(req, any(f)) {
			t.Errorf("Task: %s не в required (клиент не может на него рассчитывать)", f)
		}
	}
}
