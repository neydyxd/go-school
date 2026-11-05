//go:build !race

package l12dpgreedy

// raceScale: без детектора гонок таймаут на большой вход не растягивается.
const raceScale = 1
