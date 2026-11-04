package l08structs

import (
	"encoding/json"
	"testing"
)

func TestProductMarshal(t *testing.T) {
	tests := []struct {
		name string
		in   Product
		want string
	}{
		{
			"со скидкой",
			Product{ID: 1, Title: "Чай", Price: 15000, Discount: 500, SupplierCode: "SUP-7", cost: 9000},
			`{"id":1,"title":"Чай","price_kop":15000,"discount_kop":500}`,
		},
		{
			"без скидки",
			Product{ID: 2, Title: "Кофе", Price: 30000, SupplierCode: "SUP-1"},
			`{"id":2,"title":"Кофе","price_kop":30000}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := json.Marshal(tt.in)
			if err != nil {
				t.Fatalf("json.Marshal: %v", err)
			}
			if string(b) != tt.want {
				t.Errorf("json.Marshal(%+v) = %s, want %s", tt.in, b, tt.want)
			}
		})
	}
}

func TestProductUnmarshal(t *testing.T) {
	in := `{"id":3,"title":"Мёд","price_kop":70000,"discount_kop":0,"SupplierCode":"SUP-9","supplier_code":"SUP-9"}`
	var p Product
	if err := json.Unmarshal([]byte(in), &p); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	want := Product{ID: 3, Title: "Мёд", Price: 70000}
	if p != want {
		t.Errorf("json.Unmarshal(%s) = %+v, want %+v", in, p, want)
	}
}
