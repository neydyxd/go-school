package l02gc

import (
	"math/rand/v2"
	"slices"
	"testing"
)

const N = Nil

func TestCollectorNoMutation(t *testing.T) {
	// 0 -> 1 -> 2, 3 <-> 4 (цикл без корней), 5 одинокий.
	fields := [][]int{{1}, {2}, {}, {4}, {3}, {}}
	orig := [][]int{{1}, {2}, {}, {4}, {3}, {}}
	stack := []int{N}
	c := NewCollector(fields, []int{0}, stack)
	got := c.Finish()
	if want := []int{3, 4, 5}; !slices.Equal(got, want) {
		t.Errorf("Finish() = %v, want %v (недостижимые объекты, включая цикл)", got, want)
	}
	for i := range fields {
		if !slices.Equal(fields[i], orig[i]) {
			t.Fatalf("NewCollector изменил исходный fields[%d]: %v, want %v (скопируй поля)", i, fields[i], orig[i])
		}
	}
}

func TestCollectorSteps(t *testing.T) {
	// 0 -> {1, 2}, 1 -> {3}, 2 -> {nil}, 3 -> {}.
	c := NewCollector([][]int{{1, 2}, {3}, {N}, {}, {}}, []int{0, 0}, nil)
	check := func(step string, want []Color) {
		t.Helper()
		for i, w := range want {
			if got := c.Color(i); got != w {
				t.Errorf("%s: Color(%d) = %v, want %v (0 белый, 1 серый, 2 чёрный)", step, i, got, w)
			}
		}
	}
	check("после NewCollector", []Color{Grey, White, White, White, White})
	if !c.Step() {
		t.Fatal("Step() = false при сером корне, want true")
	}
	check("после шага 1", []Color{Black, Grey, Grey, White, White})
	c.Step()
	check("после шага 2 (FIFO: следующим сканируется 1)", []Color{Black, Black, Grey, Grey, White})
	c.Step()
	c.Step()
	check("после шага 4", []Color{Black, Black, Black, Black, White})
	if c.Step() {
		t.Error("Step() = true без серых объектов, want false")
	}
}

func TestCollectorStackScan(t *testing.T) {
	// Объект 1 доступен только со стека, объект 2 ни откуда.
	stack := []int{1, N}
	c := NewCollector([][]int{{}, {}, {}}, []int{0}, stack)
	if got := c.Color(1); got != White {
		t.Errorf("до ScanStack Color(1) = %v, want White: стек сканируется отдельным шагом, а не в NewCollector", got)
	}
	c.ScanStack()
	if got := c.Color(1); got != Grey {
		t.Errorf("после ScanStack Color(1) = %v, want Grey", got)
	}
	if got := c.Finish(); !slices.Equal(got, []int{2}) {
		t.Errorf("Finish() = %v, want [2]", got)
	}
	st := []int{1}
	c2 := NewCollector([][]int{{}, {}, {}}, []int{0}, st)
	st[0] = 2 // NewCollector должен был скопировать стек
	if got := c2.Finish(); !slices.Equal(got, []int{2}) {
		t.Errorf("Finish() без явного ScanStack = %v, want [2]: Finish сам сканирует стек (и стек скопирован в NewCollector)", got)
	}
}

func TestCollectorLostObject(t *testing.T) {
	// Классика: A (корень) -> B -> C. A уже чёрный, B серый.
	// Мутатор переносит единственный указатель на C из B в чёрный A.
	const A, B, C = 0, 1, 2
	c := NewCollector([][]int{{B, N}, {C}, {}}, []int{A}, nil)
	c.Step() // A чёрный, B серый
	c.Write(A, 1, C)
	c.Write(B, 0, N)
	if got := c.Finish(); slices.Contains(got, C) {
		t.Errorf("Finish() = %v: живой объект C (на него указывает A) попал в мусор. Барьер записи не сработал", got)
	}
}

func TestCollectorHeapToBlackStack(t *testing.T) {
	// Пункт 1 из runtime/mbarrier.go: стек уже чёрный, мутатор переносит
	// единственный указатель на C из кучи в стек (чтение без барьера) и стирает
	// его в куче. Спасает только часть Юасы: shade(старое значение).
	const A, B, C = 0, 1, 2
	c := NewCollector([][]int{{B}, {C}, {}}, []int{A}, nil)
	c.ScanStack() // стек пуст и стал чёрным
	c.Step()      // A чёрный, B серый, C белый
	c.Load(B, 0)  // локальная переменная = C
	c.Write(B, 0, N)
	if got := c.Finish(); slices.Contains(got, C) {
		t.Errorf("Finish() = %v: C на стеке, но попал в мусор. Нужна часть Юасы: красить старое значение поля", got)
	}
}

func TestCollectorGreyStackToBlackHeap(t *testing.T) {
	// Пункт 2 из runtime/mbarrier.go: стек ещё серый (не просканирован),
	// на нём единственный указатель на C. Мутатор записывает его в чёрный A
	// и затирает на стеке (без барьера). Старое значение поля A было nil,
	// часть Юасы ничего не покрасит. Спасает только часть Дейкстры.
	const A, C = 0, 1
	c := NewCollector([][]int{{N}, {}}, []int{A}, []int{C})
	c.Step()         // A чёрный
	c.StoreTop(A, 0) // A.f0 = C
	c.Pop()          // C больше нет на стеке
	if got := c.Finish(); slices.Contains(got, C) {
		t.Errorf("Finish() = %v: живой объект C (на него указывает A) попал в мусор. Пока стек серый, нужно красить и новое значение", got)
	}
}

func TestCollectorBarrierShadesBoth(t *testing.T) {
	// 0 (корень) -> 1, объекты 2 и 3 пока никто не нашёл. Стек не просканирован.
	c := NewCollector([][]int{{1, 2}, {}, {}, {}}, []int{0}, nil)
	c.Write(0, 0, 3) // старое значение 1, новое 3
	if got := c.Color(1); got == White {
		t.Errorf("после Write старое значение поля (1) белое, want серое: часть Юасы красит удаляемый указатель")
	}
	if got := c.Color(3); got == White {
		t.Errorf("после Write при сером стеке новое значение поля (3) белое, want серое: часть Дейкстры")
	}
	c.Write(0, 1, N) // запись nil не должна паниковать, старое значение 2 красится
	if got := c.Color(2); got == White {
		t.Errorf("после Write(0, 1, Nil) старое значение (2) белое, want серое")
	}
}

func TestCollectorFloatingGarbage(t *testing.T) {
	// Объект 1 стал мусором во время пометки. Барьер удаления уже покрасил его,
	// поэтому в этом цикле он выживет («плавающий мусор») и умрёт в следующем.
	c := NewCollector([][]int{{1}, {}}, []int{0}, nil)
	c.Write(0, 0, N)
	if got := c.Finish(); len(got) != 0 {
		t.Errorf("Finish() = %v, want пусто: объект, отцепленный во время пометки, доживает до следующего цикла", got)
	}
	c2 := NewCollector([][]int{{N}, {}}, []int{0}, nil)
	if got := c2.Finish(); !slices.Equal(got, []int{1}) {
		t.Errorf("следующий цикл: Finish() = %v, want [1]", got)
	}
}

func TestCollectorAlloc(t *testing.T) {
	c := NewCollector([][]int{{N}, {}}, []int{0}, nil)
	id := c.Alloc(2)
	if id != 2 {
		t.Fatalf("Alloc(2) = %d, want 2 (следующий номер)", id)
	}
	if got := c.Color(id); got != Black {
		t.Errorf("Color(новый объект) = %v, want Black: во время пометки новые объекты сразу чёрные", got)
	}
	id2 := c.Alloc(0)
	c.Write(id, 1, id2)
	c.Write(0, 0, id)
	if got := c.Finish(); !slices.Equal(got, []int{1}) {
		t.Errorf("Finish() = %v, want [1]", got)
	}
}

// reachable считает достижимые от корней и стека объекты по текущим полям.
func reachable(fields [][]int, roots, stack []int) []bool {
	seen := make([]bool, len(fields))
	todo := append(append([]int(nil), roots...), stack...)
	for len(todo) > 0 {
		o := todo[len(todo)-1]
		todo = todo[:len(todo)-1]
		if o == Nil || seen[o] {
			continue
		}
		seen[o] = true
		todo = append(todo, fields[o]...)
	}
	return seen
}

func TestCollectorRandomMutator(t *testing.T) {
	// Случайная куча и случайный мутатор: шаги пометки, сканирование стека в
	// случайный момент, чтение указателей на стек, запись с вершины стека
	// в кучу, затирание полей и локальных переменных. Как в настоящей
	// программе, указатель можно взять только из достижимого объекта.
	for seed := range uint64(500) {
		r := rand.New(rand.NewPCG(seed, 42))
		const n, nf = 30, 2
		fields := make([][]int, n)
		for i := range fields {
			fields[i] = make([]int, nf)
			for j := range fields[i] {
				fields[i][j] = Nil
				if r.IntN(3) > 0 {
					fields[i][j] = r.IntN(n)
				}
			}
		}
		roots := []int{0}
		stack := []int{1, 2}
		c := NewCollector(fields, roots, stack)
		cur := make([][]int, n)
		for i := range fields {
			cur[i] = append([]int(nil), fields[i]...)
		}
		curStack := append([]int(nil), stack...)
		scanAt := r.IntN(80)
		for step := range 80 {
			if step == scanAt {
				c.ScanStack()
			}
			live := reachable(cur, roots, curStack)
			var cand []int
			for i, ok := range live {
				if ok {
					cand = append(cand, i)
				}
			}
			obj := cand[r.IntN(len(cand))]
			f := r.IntN(nf)
			switch r.IntN(5) {
			case 0:
				c.Step()
			case 1:
				c.Load(obj, f)
				curStack = append(curStack, cur[obj][f])
			case 2:
				if len(curStack) > 0 {
					c.Pop()
					curStack = curStack[:len(curStack)-1]
				}
			case 3:
				if len(curStack) > 0 {
					c.StoreTop(obj, f)
					cur[obj][f] = curStack[len(curStack)-1]
				}
			case 4:
				c.Write(obj, f, N)
				cur[obj][f] = N
			}
		}
		garbage := c.Finish()
		live := reachable(cur, roots, curStack)
		for _, g := range garbage {
			if live[g] {
				t.Fatalf("seed %d: Finish() отдал в мусор живой объект %d (достижим от корней или стека после всех операций)", seed, g)
			}
		}
	}
}
