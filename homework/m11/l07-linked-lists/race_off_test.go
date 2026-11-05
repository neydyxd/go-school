//go:build !race

package l07linkedlists

// raceScale: без детектора гонок таймаут на большой вход не растягивается.
const raceScale = 1
