//go:build !race

package l06binarysearch

// raceScale: без детектора гонок таймаут на большой вход не растягивается.
const raceScale = 1
