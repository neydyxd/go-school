package l07godb

import (
	"reflect"
	"testing"

	"goschool/homework/internal/testdb"
)

func TestSearchUsers(t *testing.T) {
	t.Parallel()
	repo, pool := newRepo(t)
	tests := []struct {
		name  string
		query string
		limit int
		want  []int64 // id найденных
	}{
		{"без учёта регистра", "анна", 10, []int64{1, 5}},
		{"limit", "анна", 1, []int64{1}},
		{"процент ищется буквально", "50%", 10, []int64{3}},
		{"одиночный процент не значит «всё»", "%", 10, []int64{3}},
		{"подчёркивание ищется буквально", "t_", 10, []int64{4}},
		{"подчёркивание не значит «любой символ»", "o_", 10, nil},
		{"обратный слэш", `\`, 10, []int64{6}},
		{"обратный слэш перед процентом", `\%`, 10, nil},
		{"слэш и буква", `:\t`, 10, []int64{6}},
		{"инъекция", "' OR '1'='1", 10, nil},
		{"инъекция через комментарий", "x' OR true --", 10, nil},
		{"ничего не нашлось", "Зигфрид", 10, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			users, err := repo.SearchUsers(testCtx(t), tt.query, tt.limit)
			if err != nil {
				t.Fatalf("SearchUsers(%q, %d) error = %v", tt.query, tt.limit, err)
			}
			var got []int64
			for _, u := range users {
				got = append(got, u.ID)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SearchUsers(%q, %d) = id %v, want %v", tt.query, tt.limit, got, tt.want)
			}
		})
	}
	if n := testdb.Rows(t, pool, `SELECT count(*) FROM users`); n[0][0] != "6" {
		t.Errorf("после поиска в users %s строк, want 6", n[0][0])
	}
}
