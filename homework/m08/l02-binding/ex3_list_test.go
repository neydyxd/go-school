package l02binding

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"testing"
)

func catalog() *Store {
	return NewStore(
		Product{Name: "Дюна", Slug: "dune", Price: 900, Category: "books"},
		Product{Name: "Ведьмак", Slug: "witcher", Price: 2500, Category: "games"},
		Product{Name: "Алиса", Slug: "alice", Price: 500, Category: "books"},
		Product{Name: "Кино", Slug: "kino", Price: 900, Category: "music"},
		Product{Name: "Бесы", Slug: "besy", Price: 900, Category: "books"},
		Product{Name: "Арканум", Slug: "arcanum", Price: 1500, Category: "games"},
		Product{Name: "Война и мир", Slug: "war-and-peace", Price: 1200, Category: "books"},
	)
}

func ids(t *testing.T, raw json.RawMessage) []int64 {
	t.Helper()
	if string(raw) == "null" || raw == nil {
		t.Fatalf("data = null, want массив (пусть даже пустой [])")
	}
	var ps []Product
	if err := json.Unmarshal(raw, &ps); err != nil {
		t.Fatalf("data не массив товаров: %s", raw)
	}
	out := []int64{}
	for _, p := range ps {
		out = append(out, p.ID)
	}
	return out
}

func TestListProducts(t *testing.T) {
	r := ginRouter(t)
	r.GET("/products", ListProducts(catalog()))

	tests := []struct {
		name    string
		query   string
		wantIDs []int64
		want    Meta
	}{
		{"по умолчанию", "", []int64{1, 2, 3, 4, 5, 6, 7}, Meta{Page: 1, PerPage: 20, Total: 7, TotalPages: 1}},
		{"первая страница", "?per_page=3", []int64{1, 2, 3}, Meta{1, 3, 7, 3}},
		{"последняя неполная", "?per_page=3&page=3", []int64{7}, Meta{3, 3, 7, 3}},
		{"за концом списка", "?per_page=3&page=4", []int64{}, Meta{4, 3, 7, 3}},
		{"ровно делится", "?per_page=7", []int64{1, 2, 3, 4, 5, 6, 7}, Meta{1, 7, 7, 1}},
		{"фильтр", "?category=books", []int64{1, 3, 5, 7}, Meta{1, 20, 4, 1}},
		{"фильтр и страница", "?category=books&per_page=3&page=2", []int64{7}, Meta{2, 3, 4, 2}},
		{"сортировка по цене устойчивая", "?sort=price", []int64{3, 1, 4, 5, 7, 6, 2}, Meta{1, 20, 7, 1}},
		{"по убыванию цены устойчивая", "?sort=-price", []int64{2, 6, 7, 1, 4, 5, 3}, Meta{1, 20, 7, 1}},
		{"по имени", "?sort=name&category=games", []int64{6, 2}, Meta{1, 20, 2, 1}},
		{"по имени по убыванию", "?sort=-name&per_page=2", []int64{4, 1}, Meta{1, 2, 7, 4}},
		{"фильтр, сортировка, страница", "?category=books&sort=-price&per_page=2&page=2", []int64{5, 3}, Meta{2, 2, 4, 2}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := serve(r, http.MethodGet, "/products"+tt.query, "")
			if w.Code != http.StatusOK {
				t.Fatalf("GET /products%s: код %d, want 200 (тело %q)", tt.query, w.Code, w.Body)
			}
			resp := decode(t, w)
			if got := ids(t, resp.Data); !slices.Equal(got, tt.wantIDs) {
				t.Errorf("GET /products%s: id = %v, want %v", tt.query, got, tt.wantIDs)
			}
			if resp.Meta == nil || *resp.Meta != tt.want {
				t.Errorf("GET /products%s: meta = %+v, want %+v", tt.query, resp.Meta, tt.want)
			}
		})
	}
}

func TestListProductsEmptyStore(t *testing.T) {
	r := ginRouter(t)
	r.GET("/products", ListProducts(NewStore()))
	resp := decode(t, serve(r, http.MethodGet, "/products", ""))
	if got := ids(t, resp.Data); len(got) != 0 {
		t.Fatalf("пустой каталог: id = %v, want []", got)
	}
	if want := (Meta{1, 20, 0, 0}); resp.Meta == nil || *resp.Meta != want {
		t.Fatalf("пустой каталог: meta = %+v, want %+v", resp.Meta, want)
	}
}

func TestListProductsErrors(t *testing.T) {
	r := ginRouter(t)
	r.GET("/products", ListProducts(catalog()))
	tests := []struct {
		name   string
		query  string
		status int
		code   string
		fields map[string]string
	}{
		{"page не число", "?page=abc", 400, CodeBadRequest, nil},
		{"page ноль", "?page=0", 422, CodeValidation, map[string]string{"page": "не меньше 1"}},
		{"per_page велик", "?per_page=500", 422, CodeValidation, map[string]string{"per_page": "не больше 100"}},
		{"per_page ноль", "?per_page=0", 422, CodeValidation, map[string]string{"per_page": "не меньше 1"}},
		{"чужая категория", "?category=food", 422, CodeValidation, map[string]string{"category": "одно из: books, games, music"}},
		{"инъекция в sort", "?sort=price%3B%20DROP%20TABLE%20products", 422, CodeValidation, map[string]string{"sort": "одно из: price, -price, name, -name"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectError(t, serve(r, http.MethodGet, "/products"+tt.query, ""), tt.status, tt.code, tt.fields)
		})
	}
}

func TestGetProduct(t *testing.T) {
	r := ginRouter(t)
	r.GET("/products/:id", GetProduct(catalog()))

	w := serve(r, http.MethodGet, "/products/3", "")
	if w.Code != http.StatusOK {
		t.Fatalf("GET /products/3: код %d, want 200 (тело %q)", w.Code, w.Body)
	}
	var got Product
	if err := json.Unmarshal(decode(t, w).Data, &got); err != nil || got.Slug != "alice" {
		t.Fatalf("GET /products/3: data = %+v (%v), want товар alice", got, err)
	}

	tests := []struct {
		url    string
		status int
		code   string
		fields map[string]string
	}{
		{"/products/abc", 400, CodeBadRequest, nil},
		{"/products/0", 422, CodeValidation, map[string]string{"id": "не меньше 1"}},
		{"/products/-1", 422, CodeValidation, map[string]string{"id": "не меньше 1"}},
		{"/products/100", 404, CodeNotFound, nil},
	}
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			expectError(t, serve(r, http.MethodGet, tt.url, ""), tt.status, tt.code, tt.fields)
		})
	}
}

func TestListProductsStableSortLarge(t *testing.T) {
	// На 12 элементах и меньше slices.SortFunc случайно устойчив (сортировка вставками),
	// поэтому здесь каталог побольше и много одинаковых цен.
	var items []Product
	for i := range 40 {
		items = append(items, Product{Name: fmt.Sprintf("p%02d", i), Slug: fmt.Sprintf("p%02d", i), Price: int64(i%3) * 100, Category: "books"})
	}
	r := ginRouter(t)
	r.GET("/products", ListProducts(NewStore(items...)))
	for _, sort := range []string{"price", "-price"} {
		resp := decode(t, serve(r, http.MethodGet, "/products?per_page=100&sort="+sort, ""))
		got := ids(t, resp.Data)
		if len(got) != 40 {
			t.Fatalf("sort=%s: %d товаров, want 40", sort, len(got))
		}
		var want []int64
		for _, rem := range []int64{0, 1, 2} {
			if sort == "-price" {
				rem = 2 - rem
			}
			for id := int64(1); id <= 40; id++ {
				if (id-1)%3 == rem {
					want = append(want, id)
				}
			}
		}
		if !slices.Equal(got, want) {
			t.Errorf("sort=%s: id = %v, want %v (равные цены должны идти в порядке ID)", sort, got, want)
		}
	}
}
