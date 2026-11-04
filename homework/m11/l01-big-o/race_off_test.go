//go:build !race

package l01bigo

// raceScale: без детектора гонок таймаут на большой вход не растягивается.
const raceScale = 1
