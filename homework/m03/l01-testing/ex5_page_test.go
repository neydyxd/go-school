package l01testing

import "testing"

func TestPageCases(t *testing.T) {
	if len(pageCases) == 0 {
		t.Fatal("таблица pageCases пуста: заполни её в ex5_page.go")
	}
	for _, c := range pageCases {
		t.Run(c.Name, func(t *testing.T) {
			s, e, ok := Page(c.N, c.Page, c.Size)
			if s != c.Start || e != c.End || ok != c.OK {
				t.Errorf("кейс неверный: Page(%d, %d, %d) = %d, %d, %v, а в таблице want %d, %d, %v",
					c.N, c.Page, c.Size, s, e, ok, c.Start, c.End, c.OK)
			}
		})
	}
}

func TestPageMutants(t *testing.T) {
	for _, m := range pageMutants {
		t.Run(m.name, func(t *testing.T) {
			if survives(pageCases, func(c PageCase) bool {
				s, e, ok := m.fn(c.N, c.Page, c.Size)
				return s == c.Start && e == c.End && ok == c.OK
			}) {
				t.Errorf("мутант выжил: ни один кейс не отличает его от Page")
			}
		})
	}
}

type pageFunc func(n, page, size int) (int, int, bool)

var pageMutants = []struct {
	name string
	fn   pageFunc
}{
	{"последняя страница не обрезается по n", func(n, page, size int) (int, int, bool) {
		s, _, ok := Page(n, page, size)
		if !ok {
			return 0, 0, false
		}
		return s, s + size, true
	}},
	{"страница сразу за концом существует", func(n, page, size int) (int, int, bool) {
		if n > 0 && size > 0 && page > 1 && (page-1)*size == n {
			return n, n, true
		}
		return Page(n, page, size)
	}},
	{"нумерация страниц с нуля", func(n, page, size int) (int, int, bool) {
		return Page(n, page+1, size)
	}},
	{"page 0 считается первой", func(n, page, size int) (int, int, bool) {
		if page == 0 {
			page = 1
		}
		return Page(n, page, size)
	}},
	{"size 0 не проверяется", func(n, page, size int) (int, int, bool) {
		if size == 0 {
			_ = n / size
		}
		return Page(n, page, size)
	}},
	{"у пустого списка нет страниц", func(n, page, size int) (int, int, bool) {
		if n == 0 {
			return 0, 0, false
		}
		return Page(n, page, size)
	}},
	{"отрицательный n не проверяется", func(n, page, size int) (int, int, bool) {
		if n < 0 && page == 1 && size > 0 {
			return 0, 0, true
		}
		return Page(n, page, size)
	}},
	{"переполнение при огромной странице", func(n, page, size int) (int, int, bool) {
		if n < 0 || page < 1 || size < 1 {
			return 0, 0, false
		}
		start := (page - 1) * size
		if emptyFirst := n == 0 && page == 1; start >= n && !emptyFirst {
			return 0, 0, false
		}
		return start, min(start+size, n), true
	}},
}
