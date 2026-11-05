//go:build !race

package l11heap

// raceScale: без детектора гонок таймаут на большой вход не растягивается.
const raceScale = 1
