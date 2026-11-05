//go:build !race

package l10graphs

// raceScale: без детектора гонок таймаут на большой вход не растягивается.
const raceScale = 1
