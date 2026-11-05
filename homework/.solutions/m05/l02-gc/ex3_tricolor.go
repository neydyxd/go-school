package l02gc

// Color это цвет объекта в трёхцветной пометке.
type Color int

const (
	White Color = iota // ещё не найден: если останется белым, это мусор
	Grey               // найден, но его поля ещё не просмотрены
	Black              // найден, и все его поля просмотрены
)

// Nil это «нулевой указатель» в модели кучи.
const Nil = -1

// Collector: поля объектов, цвета, очередь серых и стек горутины. В рантайме
// Go очередь серых это рабочие буферы (gcWork), цвет хранится битами
// пометки, а «стек просканирован» это флаг gcscandone у горутины.
type Collector struct {
	fields     [][]int
	color      []Color
	grey       []int
	stack      []int
	stackBlack bool
}

// NewCollector копирует кучу и стек и красит корни.
func NewCollector(fields [][]int, roots []int, stack []int) *Collector {
	c := &Collector{
		fields: make([][]int, len(fields)),
		color:  make([]Color, len(fields)),
		stack:  append([]int(nil), stack...),
	}
	for i, f := range fields {
		c.fields[i] = append([]int(nil), f...)
	}
	for _, r := range roots {
		c.shade(r)
	}
	return c
}

func (c *Collector) shade(obj int) {
	if obj == Nil || c.color[obj] != White {
		return
	}
	c.color[obj] = Grey
	c.grey = append(c.grey, obj)
}

// Step сканирует один серый объект.
func (c *Collector) Step() bool {
	if len(c.grey) == 0 {
		return false
	}
	obj := c.grey[0]
	c.grey = c.grey[1:]
	for _, p := range c.fields[obj] {
		c.shade(p)
	}
	c.color[obj] = Black
	return true
}

// ScanStack красит всё, на что указывает стек, один раз за цикл.
func (c *Collector) ScanStack() {
	if c.stackBlack {
		return
	}
	for _, p := range c.stack {
		c.shade(p)
	}
	c.stackBlack = true
}

// Write: shade(*slot); если стек серый, shade(ptr); *slot = ptr.
func (c *Collector) Write(obj, field, target int) {
	c.shade(c.fields[obj][field])
	if !c.stackBlack {
		c.shade(target)
	}
	c.fields[obj][field] = target
}

// Load читает поле на стек без барьера.
func (c *Collector) Load(obj, field int) {
	c.stack = append(c.stack, c.fields[obj][field])
}

// Pop выбрасывает вершину стека без барьера.
func (c *Collector) Pop() {
	c.stack = c.stack[:len(c.stack)-1]
}

// StoreTop пишет вершину стека в кучу через барьер.
func (c *Collector) StoreTop(obj, field int) {
	c.Write(obj, field, c.stack[len(c.stack)-1])
}

// Alloc выделяет объект сразу чёрным.
func (c *Collector) Alloc(nfields int) int {
	f := make([]int, nfields)
	for i := range f {
		f[i] = Nil
	}
	c.fields = append(c.fields, f)
	c.color = append(c.color, Black)
	return len(c.fields) - 1
}

// Color возвращает текущий цвет объекта.
func (c *Collector) Color(obj int) Color {
	return c.color[obj]
}

// Finish сканирует стек, домечивает и собирает белые.
func (c *Collector) Finish() []int {
	c.ScanStack()
	for c.Step() {
	}
	var garbage []int
	for i, col := range c.color {
		if col == White {
			garbage = append(garbage, i)
		}
	}
	return garbage
}
