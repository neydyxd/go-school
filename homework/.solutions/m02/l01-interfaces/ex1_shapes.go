package l01interfaces

import "math"

// Shape: фигура на плоскости.
type Shape interface {
	Area() float64
	Perimeter() float64
}

// Rect: прямоугольник со сторонами W и H.
type Rect struct{ W, H float64 }

// Circle: круг радиуса R.
type Circle struct{ R float64 }

func (r Rect) Area() float64      { return r.W * r.H }
func (r Rect) Perimeter() float64 { return 2 * (r.W + r.H) }

func (c Circle) Area() float64      { return math.Pi * c.R * c.R }
func (c Circle) Perimeter() float64 { return 2 * math.Pi * c.R }

func TotalArea(shapes []Shape) float64 {
	var sum float64
	for _, s := range shapes {
		if s == nil {
			continue
		}
		sum += s.Area()
	}
	return sum
}

func Largest(shapes []Shape) (Shape, bool) {
	var best Shape
	for _, s := range shapes {
		if s == nil {
			continue
		}
		if best == nil || s.Area() > best.Area() {
			best = s
		}
	}
	return best, best != nil
}
