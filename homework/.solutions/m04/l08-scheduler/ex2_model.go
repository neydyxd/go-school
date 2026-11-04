package l08scheduler

// Игрушечная модель планировщика Go: только очереди, без потоков и времени.
// Горутины здесь просто номера (>= 1), 0 значит «пусто».

// LocalCap: ёмкость локальной очереди P. В рантайме 256, здесь 4, чтобы
// переполнение было видно в коротких тестах.
const LocalCap = 4

// P: логический процессор со своей очередью.
type P struct {
	RunNext   int   // слот «следующей» горутины, 0: пусто
	Local     []int // локальная очередь FIFO, не длиннее LocalCap
	SchedTick int   // сколько раз этот P выбрал горутину (успешные Next)
}

// Model: глобальная очередь и набор P.
type Model struct {
	Global []int // глобальная очередь FIFO
	Ps     []*P
}

// NewModel создаёт модель с nprocs процессорами (nprocs < 1 считать 1).
// Очереди пустые (не nil).
func NewModel(nprocs int) *Model {
	m := &Model{Global: []int{}}
	for range max(nprocs, 1) {
		m.Ps = append(m.Ps, &P{Local: []int{}})
	}
	return m
}

// Spawn моделирует оператор go на процессоре p (runqput с next=true):
//  1. новая горутина g встаёт в RunNext;
//  2. прежняя RunNext (если была) уходит в хвост Local;
//  3. если Local уже полна (LocalCap), то первая половина Local
//     (LocalCap/2 горутин из головы) и вытесненная горутина в этом порядке
//     переезжают в хвост Global, а вытесненная в Local не попадает.
func (m *Model) Spawn(p, g int) {
	pp := m.Ps[p]
	old := pp.RunNext
	pp.RunNext = g
	if old == 0 {
		return
	}
	if len(pp.Local) < LocalCap {
		pp.Local = append(pp.Local, old)
		return
	}
	half := LocalCap / 2
	m.Global = append(m.Global, pp.Local[:half]...)
	m.Global = append(m.Global, old)
	pp.Local = append([]int{}, pp.Local[half:]...)
}

// Next моделирует findRunnable для процессора p: возвращает горутину,
// которую P запустит следующей, и true; если работы нигде нет: 0, false.
// Порядок поиска:
//  1. Если SchedTick%61 == 0 и Global не пуста: голова Global (честность:
//     иначе глобальная очередь могла бы голодать).
//  2. RunNext.
//  3. Голова Local.
//  4. Global пачкой: n = min(LocalCap/2, len(Global), len(Global)/len(Ps)+1)
//     горутин из головы; первую вернуть, остальные n-1 в хвост Local.
//  5. Кража: обойти остальные P по кругу, начиная с p+1. У первого P с
//     непустой Local забрать из головы половину с округлением вверх
//     (k = len - len/2): первую вернуть, остальные в хвост своей Local.
//  6. Если ни у кого нет Local, второй круг (тот же порядок): украсть
//     RunNext у первого P, у которого он не пуст.
//
// Каждый успешный Next увеличивает SchedTick процессора p на 1.
// (В рантайме горутина из runnext наследует квант и счётчик не трогает,
// здесь упрощаем.)
func (m *Model) Next(p int) (int, bool) {
	g := m.find(p)
	if g == 0 {
		return 0, false
	}
	m.Ps[p].SchedTick++
	return g, true
}

func (m *Model) find(p int) int {
	pp := m.Ps[p]
	if pp.SchedTick%61 == 0 && len(m.Global) > 0 {
		g := m.Global[0]
		m.Global = m.Global[1:]
		return g
	}
	if g := pp.RunNext; g != 0 {
		pp.RunNext = 0
		return g
	}
	if len(pp.Local) > 0 {
		g := pp.Local[0]
		pp.Local = pp.Local[1:]
		return g
	}
	if len(m.Global) > 0 {
		n := min(LocalCap/2, len(m.Global), len(m.Global)/len(m.Ps)+1)
		batch := m.Global[:n]
		m.Global = m.Global[n:]
		pp.Local = append(pp.Local, batch[1:]...)
		return batch[0]
	}
	for i := 1; i < len(m.Ps); i++ {
		v := m.Ps[(p+i)%len(m.Ps)]
		if l := len(v.Local); l > 0 {
			k := l - l/2
			stolen := v.Local[:k]
			v.Local = v.Local[k:]
			pp.Local = append(pp.Local, stolen[1:]...)
			return stolen[0]
		}
	}
	for i := 1; i < len(m.Ps); i++ {
		v := m.Ps[(p+i)%len(m.Ps)]
		if g := v.RunNext; g != 0 {
			v.RunNext = 0
			return g
		}
	}
	return 0
}
