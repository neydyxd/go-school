//go:build !race

package l09trees

// raceScale: без детектора гонок таймаут на большой вход не растягивается.
const raceScale = 1
