//go:build !race

package l05stackqueue

// raceScale: без детектора гонок таймаут на большой вход не растягивается.
const raceScale = 1
