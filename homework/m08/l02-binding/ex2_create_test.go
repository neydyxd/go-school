package l02binding

import (
	"bytes"
	"encoding/json"
	"io"
	"maps"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// ginRouter настраивает валидатор Gin через SetupValidator и возвращает пустой роутер.
func ginRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		t.Fatal("binding.Validator.Engine() не *validator.Validate")
	}
	if err := SetupValidator(v); err != nil {
		t.Fatalf("SetupValidator: %v", err)
	}
	return gin.New()
}

func serve(h http.Handler, method, url, body string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, url, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	h.ServeHTTP(w, req)
	return w
}

type response struct {
	Data  json.RawMessage `json:"data"`
	Meta  *Meta           `json:"meta"`
	Error *ErrorBody      `json:"error"`
}

// decode требует ровно один JSON-объект в теле.
func decode(t *testing.T, w *httptest.ResponseRecorder) response {
	t.Helper()
	dec := json.NewDecoder(bytes.NewReader(w.Body.Bytes()))
	dec.DisallowUnknownFields()
	var r response
	if err := dec.Decode(&r); err != nil {
		t.Fatalf("тело не в едином формате {data, meta, error}: %q (%v)", w.Body, err)
	}
	if _, err := dec.Token(); err != io.EOF {
		t.Fatalf("в теле больше одного JSON-значения: %q (забыт return?)", w.Body)
	}
	return r
}

// expectError проверяет код ответа, код ошибки и (если wantFields != nil) поля.
func expectError(t *testing.T, w *httptest.ResponseRecorder, status int, code string, wantFields map[string]string) {
	t.Helper()
	if w.Code != status {
		t.Fatalf("код ответа %d, want %d (тело %q)", w.Code, status, w.Body)
	}
	r := decode(t, w)
	if r.Error == nil || r.Data != nil {
		t.Fatalf("тело %q, want только {\"error\": {...}}", w.Body)
	}
	if r.Error.Code != code || r.Error.Message == "" {
		t.Fatalf("error = %+v, want code %q и непустой message", *r.Error, code)
	}
	if wantFields != nil && !maps.Equal(r.Error.Fields, wantFields) {
		t.Fatalf("error.fields = %v, want %v", r.Error.Fields, wantFields)
	}
}

func TestCreateProduct(t *testing.T) {
	r := ginRouter(t)
	store := NewStore(Product{Name: "Go", Slug: "taken", Price: 100, Category: "books"})
	r.POST("/products", CreateProduct(store))

	tests := []struct {
		name   string
		body   string
		status int
		code   string
		fields map[string]string
	}{
		{"не JSON", `{"name":`, 400, CodeBadRequest, nil},
		{"пустое тело", ``, 400, CodeBadRequest, nil},
		{"не тот тип", `{"name":"Go","slug":"go","price":"дорого","category":"books"}`, 400, CodeBadRequest, nil},
		{"пустой объект", `{}`, 422, CodeValidation, map[string]string{
			"name": "обязательное поле", "slug": "обязательное поле", "price": "обязательное поле", "category": "обязательное поле",
		}},
		{"короткое имя и плохой slug", `{"name":"G","slug":"Go!","price":1,"category":"books"}`, 422, CodeValidation, map[string]string{
			"name": "минимальная длина 2", "slug": "только строчные латинские буквы, цифры и дефисы",
		}},
		{"имя длиннее 100", `{"name":"` + strings.Repeat("я", 101) + `","slug":"long","price":1,"category":"books"}`, 422, CodeValidation, map[string]string{
			"name": "максимальная длина 100",
		}},
		{"отрицательная цена и чужая категория", `{"name":"Go","slug":"go","price":-5,"category":"food"}`, 422, CodeValidation, map[string]string{
			"price": "больше 0", "category": "одно из: books, games, music",
		}},
		{"плохие теги", `{"name":"Go","slug":"go","price":1,"category":"books","tags":["ok","x"]}`, 422, CodeValidation, map[string]string{
			"tags[1]": "минимальная длина 2",
		}},
		{"слишком много тегов", `{"name":"Go","slug":"go","price":1,"category":"books","tags":["aa","bb","cc","dd","ee","ff"]}`, 422, CodeValidation, map[string]string{
			"tags": "максимум элементов: 5",
		}},
		{"slug занят", `{"name":"Go","slug":"taken","price":1,"category":"books"}`, 409, CodeConflict, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectError(t, serve(r, http.MethodPost, "/products", tt.body), tt.status, tt.code, tt.fields)
		})
	}
	if n := len(store.All()); n != 1 {
		t.Fatalf("после ошибочных запросов в хранилище %d товаров, want 1", n)
	}
}

func TestCreateProductOK(t *testing.T) {
	r := ginRouter(t)
	store := NewStore()
	r.POST("/products", CreateProduct(store))

	tests := []struct {
		body string
		want Product
	}{
		{`{"name":"Книга про Go","slug":"go-book","price":159900,"category":"books","tags":["go","книги"]}`,
			Product{ID: 1, Name: "Книга про Go", Slug: "go-book", Price: 159900, Category: "books", Tags: []string{"go", "книги"}}},
		{`{"name":"Ян","slug":"yan","price":1,"category":"music"}`,
			Product{ID: 2, Name: "Ян", Slug: "yan", Price: 1, Category: "music", Tags: []string{}}},
	}
	for _, tt := range tests {
		w := serve(r, http.MethodPost, "/products", tt.body)
		if w.Code != http.StatusCreated {
			t.Fatalf("POST %s: код %d, want 201 (тело %q)", tt.body, w.Code, w.Body)
		}
		resp := decode(t, w)
		if resp.Error != nil {
			t.Fatalf("POST %s: в успешном ответе есть error: %q", tt.body, w.Body)
		}
		var got Product
		if err := json.Unmarshal(resp.Data, &got); err != nil {
			t.Fatalf("data не Product: %s", resp.Data)
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("POST %s: data = %+v, want %+v", tt.body, got, tt.want)
		}
		if strings.Contains(string(resp.Data), `"tags":null`) {
			t.Errorf("POST %s: tags = null, want []", tt.body)
		}
	}
}
