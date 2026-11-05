//go:build !race

package l08sorting

// raceScale: без детектора гонок таймаут на большой вход не растягивается.
const raceScale = 1
