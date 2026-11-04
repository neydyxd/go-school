//go:build !race

package l03hashing

// raceScale: без детектора гонок таймаут на большой вход не растягивается.
const raceScale = 1
