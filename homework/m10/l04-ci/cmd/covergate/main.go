// Команда covergate: шлагбаум покрытия для CI.
//
//	go test -coverprofile=coverage.out ./...
//	go run ./m10/l04-ci/cmd/covergate -min 80 coverage.out
//
// Печатает общее покрытие и худшие файлы, выходит с кодом 1, если покрытие
// ниже порога. Код готовый: заработает, когда решишь ex1 и ex2.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	l04ci "goschool/homework/m10/l04-ci"
)

func main() {
	minPct := flag.Float64("min", 80, "минимальное покрытие в процентах")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "использование: covergate -min 80 coverage.out")
		os.Exit(2)
	}
	path := flag.Arg(0)

	total, err := withFile(path, l04ci.Coverage)
	if err != nil {
		fmt.Fprintln(os.Stderr, "covergate:", err)
		os.Exit(2)
	}
	fmt.Printf("покрытие %.1f%%, порог %.1f%%\n", total, *minPct)
	if total >= *minPct {
		return
	}
	worst, err := withFile(path, func(r io.Reader) ([]l04ci.FileCoverage, error) {
		return l04ci.WorstFiles(r, *minPct)
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "covergate:", err)
		os.Exit(2)
	}
	for _, fc := range worst {
		fmt.Printf("  %5.1f%%  %s\n", fc.Percent, fc.File)
	}
	os.Exit(1)
}

func withFile[T any](path string, fn func(io.Reader) (T, error)) (T, error) {
	f, err := os.Open(path) //nolint:gosec // путь к профилю передаёт сам пользователь CLI
	if err != nil {
		var zero T
		return zero, err
	}
	defer f.Close()
	return fn(f)
}
