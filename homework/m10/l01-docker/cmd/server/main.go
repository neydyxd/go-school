// Команда server: маленький сервис для упражнений с Docker.
// Конфиг из окружения (LoadConfig), /healthz и /readyz, graceful shutdown
// по SIGTERM и флаг -healthcheck для HEALTHCHECK в образе без curl.
//
// Код готовый, трогать его не нужно: заработает, когда решишь ex1 и ex2.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	l01docker "goschool/homework/m10/l01-docker"
)

func main() {
	os.Exit(run())
}

// run возвращает код выхода: так defer-ы внутри успевают отработать
// (os.Exit их не выполняет).
func run() int {
	healthcheck := flag.Bool("healthcheck", false, "проверить /healthz своего же процесса и выйти с кодом 0 или 1")
	flag.Parse()

	cfg, err := l01docker.LoadConfig(os.Getenv)
	if *healthcheck {
		// DATABASE_URL для проверки не нужен, порт берём какой получился.
		url := fmt.Sprintf("http://127.0.0.1:%d/healthz", cfg.Port)
		if err := l01docker.Probe(context.Background(), url, 2*time.Second); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		return 0
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "конфиг:", err)
		return 2
	}

	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel}))
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error("pgxpool", "err", err)
		return 1
	}
	defer pool.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second)
		defer cancel()
		if err := pool.Ping(ctx); err != nil {
			http.Error(w, "db: "+err.Error(), http.StatusServiceUnavailable)
			return
		}
		fmt.Fprintln(w, "ready")
	})

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	errc := make(chan error, 1)
	go func() { errc <- srv.ListenAndServe() }()
	log.Info("сервер запущен", "addr", srv.Addr)

	select {
	case err := <-errc:
		log.Error("сервер упал", "err", err)
		return 1
	case <-ctx.Done():
	}
	log.Info("получен сигнал, останавливаюсь")
	sctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(sctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error("shutdown", "err", err)
		return 1
	}
	log.Info("остановлен")
	return 0
}
