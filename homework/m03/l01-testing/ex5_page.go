package l01testing

// Page считает границы страницы для пагинации: items[start:end] это элементы
// страницы page (нумерация с 1) размера size в списке из n элементов.
// Это готовая и правильная функция, менять её не нужно.
//
// Правила:
//   - ok=false и нули, если n < 0, page < 1, size < 1 или страница целиком за концом списка;
//   - последняя страница может быть неполной: end не больше n;
//   - у пустого списка есть одна пустая первая страница: Page(0, 1, 10) = 0, 0, true;
//   - огромные page и size не должны приводить к переполнению int.
func Page(n, page, size int) (start, end int, ok bool) {
	if n < 0 || page < 1 || size < 1 {
		return 0, 0, false
	}
	if page-1 > n/size { // проверка до умножения: (page-1)*size может переполниться
		return 0, 0, false
	}
	start = (page - 1) * size
	emptyFirst := n == 0 && page == 1 // единственная страница пустого списка
	if start >= n && !emptyFirst {
		return 0, 0, false
	}
	return start, start + min(size, n-start), true
}

// PageCase это один кейс таблицы для Page.
type PageCase struct {
	Name          string
	N, Page, Size int
	Start, End    int
	OK            bool
}

// pageCases это твоя таблица тестов для Page.
//
// Задание то же, что в ex2: таблица должна быть верной для Page и ловить
// всех мутантов из ex5_page_test.go. Здесь мутанты хитрее: почти все ошибки
// живут на границах. Ищи их методом граничных значений: первая и последняя
// страница, страница сразу за концом, неполная страница, ноль, отрицательные
// числа, пустой список, очень большие значения (math.MaxInt).
var pageCases = []PageCase{
	// TODO: например {Name: "первая страница", N: 10, Page: 1, Size: 3, Start: 0, End: 3, OK: true},
}
