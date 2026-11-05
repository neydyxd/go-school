// Команда library запускает API библиотеки. Это точка сборки: конфиг, пул, слои, сервер.
// Запуск (из папки homework, база поднята, схема загружена из testdata/schema.sql):
//
//	DATABASE_URL=postgres://goschool:goschool@localhost:55432/goschool go run ./m08/l04-architecture/cmd/library
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	app "goschool/homework/m08/l04-architecture"
)

func main() {
	if err := run(); err != nil {
		slog.Error("сервис остановлен с ошибкой", "err", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		return err
	}
	defer pool.Close()

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           app.NewRouter(pool),
		ReadHeaderTimeout: 5 * time.Second,
	}
	errc := make(chan error, 1)
	go func() { errc <- srv.ListenAndServe() }()
	slog.Info("слушаю", "addr", srv.Addr)

	select {
	case err := <-errc:
		return err
	case <-ctx.Done():
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
