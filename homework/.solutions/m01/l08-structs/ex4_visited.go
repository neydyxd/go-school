package l08structs

// Point: клетка на клетчатом поле.
type Point struct {
	X, Y int
}

// Visited считает, сколько разных клеток посетил робот.
// Робот стартует в (0, 0) (стартовая клетка тоже считается) и выполняет
// команды из path: 'U' (y+1), 'D' (y-1), 'L' (x-1), 'R' (x+1).
// Другие символы игнорируются.
//
// Point сравнимый (все поля сравнимые), поэтому его можно сделать ключом мапы.
// Пустой path: 1.
func Visited(path string) int {
	var p Point
	seen := map[Point]struct{}{p: {}}
	for _, c := range path {
		switch c {
		case 'U':
			p.Y++
		case 'D':
			p.Y--
		case 'L':
			p.X--
		case 'R':
			p.X++
		default:
			continue
		}
		seen[p] = struct{}{}
	}
	return len(seen)
}
