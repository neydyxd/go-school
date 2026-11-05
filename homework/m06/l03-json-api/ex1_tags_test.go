package l03jsonapi

import (
	"encoding/json"
	"testing"
	"time"
)

func TestProductMarshal(t *testing.T) {
	zero := 0
	ten := 10
	ts := time.Date(2026, 11, 5, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name string
		in   Product
		want string
	}{
		{"минимальный: ноль в цене остаётся", Product{ID: 1, Name: "Чай"},
			`{"id":"1","name":"Чай","price_kopecks":0}`},
		{"большой ID строкой", Product{ID: 9007199254740993, Name: "x", Price: 100},
			`{"id":"9007199254740993","name":"x","price_kopecks":100}`},
		{"пустой срез тегов не выводится", Product{ID: 2, Name: "x", Tags: []string{}},
			`{"id":"2","name":"x","price_kopecks":0}`},
		{"теги", Product{ID: 3, Name: "x", Tags: []string{"a", "b"}},
			`{"id":"3","name":"x","price_kopecks":0,"tags":["a","b"]}`},
		{"скидка 0% это не отсутствие скидки", Product{ID: 4, Name: "x", Discount: &zero},
			`{"id":"4","name":"x","price_kopecks":0,"discount":0}`},
		{"скидка 10%", Product{ID: 5, Name: "x", Discount: &ten},
			`{"id":"5","name":"x","price_kopecks":0,"discount":10}`},
		{"время создания", Product{ID: 6, Name: "x", CreatedAt: ts},
			`{"id":"6","name":"x","price_kopecks":0,"created_at":"2026-11-05T12:00:00Z"}`},
		{"секрет не утекает", Product{ID: 7, Name: "x", Secret: "s3cr3t"},
			`{"id":"7","name":"x","price_kopecks":0}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := json.Marshal(tt.in)
			if err != nil {
				t.Fatalf("json.Marshal(%+v): %v", tt.in, err)
			}
			if string(b) != tt.want {
				t.Errorf("json.Marshal(%+v) =\n %s\nwant\n %s", tt.in, b, tt.want)
			}
		})
	}
}

func TestProductUnmarshal(t *testing.T) {
	in := `{"id":"42","name":"Кофе","price_kopecks":25000,"Secret":"hacked","secret":"hacked","-":"hacked"}`
	var p Product
	if err := json.Unmarshal([]byte(in), &p); err != nil {
		t.Fatalf("json.Unmarshal(%s): %v", in, err)
	}
	want := Product{ID: 42, Name: "Кофе", Price: 25000}
	if p.ID != want.ID || p.Name != want.Name || p.Price != want.Price || p.Secret != "" {
		t.Errorf("json.Unmarshal(%s) = %+v, want %+v", in, p, want)
	}
}
