package l01interfaces

// Shape: фигура на плоскости.
type Shape interface {
	Area() float64
	Perimeter() float64
}

// Rect: прямоугольник со сторонами W и H.
type Rect struct{ W, H float64 }

// Circle: круг радиуса R.
type Circle struct{ R float64 }

// Area возвращает площадь прямоугольника.
func (r Rect) Area() float64 { panic("TODO: реализуй Rect.Area") }

// Perimeter возвращает периметр прямоугольника.
func (r Rect) Perimeter() float64 { panic("TODO: реализуй Rect.Perimeter") }

// Area возвращает площадь круга (используй math.Pi).
func (c Circle) Area() float64 { panic("TODO: реализуй Circle.Area") }

// Perimeter возвращает длину окружности (используй math.Pi).
func (c Circle) Perimeter() float64 { panic("TODO: реализуй Circle.Perimeter") }

// TotalArea возвращает сумму площадей фигур.
// Пустой или nil слайс: 0. Элементы, равные nil, пропусти (не паникуй).
func TotalArea(shapes []Shape) float64 {
	panic("TODO: реализуй TotalArea")
}

// Largest возвращает фигуру с наибольшей площадью и true.
// Если фигур нет (или все nil), верни nil, false.
// При равных площадях побеждает та, что встретилась раньше.
func Largest(shapes []Shape) (Shape, bool) {
	panic("TODO: реализуй Largest")
}
