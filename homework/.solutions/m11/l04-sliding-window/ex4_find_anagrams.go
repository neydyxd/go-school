package l04slidingwindow

// FindAnagrams возвращает позиции (в рунах, по возрастанию) всех подстрок s,
// которые являются анаграммами p: те же руны в тех же количествах.
// Если p пустая или длиннее s, верни пустой результат (или nil).
//
// Ожидаемая сложность: O(len(s)) при небольшом алфавите. Сравнивать каждое
// окно с p заново (O(len(s)*len(p))) слишком медленно.
func FindAnagrams(s, p string) []int {
	sr, pr := []rune(s), []rune(p)
	m := len(pr)
	if m == 0 || m > len(sr) {
		return nil
	}
	// diff[r] = (сколько r в окне) - (сколько r в p); bad: сколько рун с diff != 0.
	diff := make(map[rune]int)
	bad := 0
	add := func(r rune, d int) {
		if diff[r] == 0 {
			bad++
		}
		diff[r] += d
		if diff[r] == 0 {
			bad--
		}
	}
	for _, r := range pr {
		add(r, -1)
	}
	var res []int
	for i, r := range sr {
		add(r, +1)
		if i >= m {
			add(sr[i-m], -1) // руна выходит из окна
		}
		if i >= m-1 && bad == 0 {
			res = append(res, i-m+1)
		}
	}
	return res
}
