//go:build !race

package l02twopointers

// raceScale: без детектора гонок таймаут на большой вход не растягивается.
const raceScale = 1
