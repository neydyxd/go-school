package l08scheduler

import (
	"slices"
	"testing"
)

func drain(m *Model, p int) []int {
	var out []int
	for {
		g, ok := m.Next(p)
		if !ok {
			return out
		}
		out = append(out, g)
	}
}

func TestNewModel(t *testing.T) {
	m := NewModel(0)
	if len(m.Ps) != 1 || m.Global == nil || m.Ps[0].Local == nil {
		t.Fatalf("NewModel(0): %d P, Global=%v, Local=%v; want 1 P и пустые не-nil очереди", len(m.Ps), m.Global, m.Ps[0].Local)
	}
	if g, ok := m.Next(0); ok || g != 0 || m.Ps[0].SchedTick != 0 {
		t.Errorf("Next в пустой модели = %d, %v (SchedTick=%d), want 0, false и счётчик не меняется", g, ok, m.Ps[0].SchedTick)
	}
}

func TestSpawnRunNext(t *testing.T) {
	m := NewModel(1)
	for g := 1; g <= 3; g++ {
		m.Spawn(0, g)
	}
	p := m.Ps[0]
	if p.RunNext != 3 || !slices.Equal(p.Local, []int{1, 2}) {
		t.Fatalf("после go 1, 2, 3: RunNext=%d, Local=%v, want 3, [1 2]", p.RunNext, p.Local)
	}
	if got := drain(m, 0); !slices.Equal(got, []int{3, 1, 2}) {
		t.Errorf("порядок запуска %v, want [3 1 2]: свежая горутина из runnext первой", got)
	}
}

func TestSpawnOverflow(t *testing.T) {
	m := NewModel(1)
	for g := 1; g <= 6; g++ {
		m.Spawn(0, g)
	}
	p := m.Ps[0]
	if p.RunNext != 6 || !slices.Equal(p.Local, []int{3, 4}) || !slices.Equal(m.Global, []int{1, 2, 5}) {
		t.Errorf("после go 1..6: RunNext=%d, Local=%v, Global=%v; want 6, [3 4], [1 2 5]", p.RunNext, p.Local, m.Global)
	}
}

func TestNextGlobalFairness(t *testing.T) {
	m := NewModel(1)
	m.Global = []int{100}
	m.Spawn(0, 1)
	if g, _ := m.Next(0); g != 100 {
		t.Fatalf("первый Next (SchedTick=0) = %d, want 100 из Global", g)
	}
	// Две горутины постоянно порождают друг друга через runnext. Без проверки
	// раз в 61 тик горутина из Global голодала бы вечно.
	m.Global = []int{999}
	next := 2
	for call := 2; call <= 70; call++ {
		g, ok := m.Next(0)
		if !ok {
			t.Fatalf("Next №%d: нет работы, want горутину", call)
		}
		if g == 999 {
			if call != 62 {
				t.Fatalf("горутина из Global запущена на вызове №%d, want №62 (SchedTick=61)", call)
			}
			return
		}
		m.Spawn(0, next)
		next++
	}
	t.Fatal("горутина из Global так и не запустилась")
}

func TestNextGlobalBatch(t *testing.T) {
	m := NewModel(2)
	m.Global = []int{1, 2, 3, 4, 5, 6, 7}
	m.Ps[0].SchedTick = 1
	g, ok := m.Next(0)
	if !ok || g != 1 || !slices.Equal(m.Ps[0].Local, []int{2}) || !slices.Equal(m.Global, []int{3, 4, 5, 6, 7}) {
		t.Errorf("Next = %d, Local=%v, Global=%v; want 1, [2], [3 4 5 6 7]", g, m.Ps[0].Local, m.Global)
	}

	m = NewModel(4)
	m.Global = []int{1, 2, 3}
	m.Ps[0].SchedTick = 1
	g, _ = m.Next(0)
	if g != 1 || len(m.Ps[0].Local) != 0 || !slices.Equal(m.Global, []int{2, 3}) {
		t.Errorf("4 P и 3 в Global: Next = %d, Local=%v, Global=%v; want 1, [], [2 3] (3/4+1 = 1)", g, m.Ps[0].Local, m.Global)
	}
}

func TestNextOrder(t *testing.T) {
	m := NewModel(1)
	m.Ps[0].SchedTick = 1
	m.Ps[0].RunNext = 5
	m.Ps[0].Local = []int{6}
	m.Global = []int{7}
	if got := drain(m, 0); !slices.Equal(got, []int{5, 6, 7}) {
		t.Errorf("порядок %v, want [5 6 7]: runnext, затем Local, затем Global", got)
	}
	if m.Ps[0].SchedTick != 4 {
		t.Errorf("SchedTick = %d, want 4", m.Ps[0].SchedTick)
	}
}

func TestNextSteal(t *testing.T) {
	m := NewModel(3)
	m.Ps[0].SchedTick = 1
	m.Ps[1].Local = []int{1, 2, 3}
	m.Ps[2].Local = []int{4, 5}
	g, ok := m.Next(0)
	if !ok || g != 1 || !slices.Equal(m.Ps[0].Local, []int{2}) || !slices.Equal(m.Ps[1].Local, []int{3}) {
		t.Errorf("кража у P1: Next = %d, свой Local=%v, у P1 осталось %v; want 1, [2], [3]", g, m.Ps[0].Local, m.Ps[1].Local)
	}

	m = NewModel(3)
	m.Ps[2].SchedTick = 1
	m.Ps[0].Local = []int{9}
	m.Ps[1].Local = []int{8}
	if g, _ := m.Next(2); g != 9 || len(m.Ps[0].Local) != 0 || len(m.Ps[1].Local) != 1 {
		t.Errorf("P2 крадёт по кругу с P0: Next = %d, want 9 (P0.Local=%v, P1.Local=%v)", g, m.Ps[0].Local, m.Ps[1].Local)
	}
}

func TestNextStealRunNext(t *testing.T) {
	m := NewModel(3)
	m.Ps[0].SchedTick = 1
	m.Ps[1].RunNext = 9
	m.Ps[2].Local = []int{4}
	if g, _ := m.Next(0); g != 4 || m.Ps[1].RunNext != 9 {
		t.Errorf("Next = %d, P1.RunNext=%d; want 4 и 9 на месте: runnext чужого P крадут в последнюю очередь", g, m.Ps[1].RunNext)
	}
	if g, _ := m.Next(0); g != 9 || m.Ps[1].RunNext != 0 {
		t.Errorf("Next = %d, P1.RunNext=%d; want 9 и 0: больше красть нечего, кроме runnext", g, m.Ps[1].RunNext)
	}
}
