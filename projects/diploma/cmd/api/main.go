// Команда api: точка сборки сервиса. Здесь только wiring (DI руками) и жизненный цикл:
// конфиг, логгер, пул, Redis, метрики, роутер, graceful shutdown.
//
//	api               запустить HTTP-сервер
//	api -migrate      применить миграции и выйти (шаг деплоя, сервис compose «migrate»)
//	api -healthcheck  проверить свой /healthz и выйти с кодом 0 или 1 (HEALTHCHECK в distroless)
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
	"strconv"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/redis/go-redis/v9"

	"goschool/diploma/internal/config"
	api "goschool/diploma/internal/http"
	"goschool/diploma/migrations"
)

func main() {
	migrate := flag.Bool("migrate", false, "применить миграции и выйти")
	healthcheck := flag.Bool("healthcheck", false, "проверить /healthz и выйти")
	flag.Parse()

	if *healthcheck {
		os.Exit(probe())
	}
	if err := run(*migrate); err != nil {
		slog.Error("сервис остановлен с ошибкой", "err", err.Error())
		os.Exit(1)
	}
}

func run(migrateOnly bool) error {
	// JSON с первой строки: даже ошибка конфига уходит в лог структурно.
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	cfg, err := config.Load(os.Getenv)
	if err != nil {
		return fmt.Errorf("конфиг: %w", err)
	}
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel}))
	slog.SetDefault(log)
	log.Info("старт", "config", cfg) // пароли скрывает Config.LogValue

	// ctx отменится по Ctrl+C или SIGTERM (так останавливают контейнеры).
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("пул базы: %w", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("база недоступна: %w", err)
	}

	if migrateOnly {
		n, err := migrations.Up(ctx, stdlib.OpenDBFromPool(pool))
		if err != nil {
			return err
		}
		log.Info("миграции применены", "new", n)
		return nil
	}

	checks := map[string]api.Check{"postgres": pool.Ping}
	if cfg.RedisURL != "" {
		opts, err := redis.ParseURL(cfg.RedisURL)
		if err != nil {
			return fmt.Errorf("REDIS_URL: %w", err)
		}
		rdb := redis.NewClient(opts)
		defer func() { _ = rdb.Close() }()
		checks["redis"] = func(ctx context.Context) error { return rdb.Ping(ctx).Err() }
	}

	reg := prometheus.NewRegistry()
	reg.MustRegister(collectors.NewGoCollector(), collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	// Сюда собираешь свою тему, слой за слоем:
	//   txm := repository.NewTxManager(pool)
	//   svc := service.NewBookingService(repository.NewBookings(pool), txm, time.Now)
	//   riverClient (воркеры River стартуют здесь же и останавливаются до закрытия пула)
	// и передаёшь сервисы в api.Deps.
	var draining atomic.Bool
	router := api.NewRouter(api.Deps{Log: log, Registry: reg, Checks: checks, Draining: &draining})

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	errCh := make(chan error, 1)
	go func() {
		log.Info("слушаю", "addr", srv.Addr)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		return err // например, порт занят
	case <-ctx.Done():
		stop() // второй Ctrl+C или SIGTERM снова убивает процесс сразу
	}

	// Порядок остановки из урока про наблюдаемость (m10/l02): сначала /readyz = 503,
	// пауза, пока балансировщик уберёт нас из ротации, потом Shutdown.
	draining.Store(true)
	log.Info("останавливаюсь", "drain", cfg.DrainDelay.String(), "timeout", cfg.ShutdownTimeout.String())
	time.Sleep(cfg.DrainDelay)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	if err := <-errCh; !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	log.Info("остановлен")
	return nil
}

// probe: маленький HTTP-клиент для HEALTHCHECK. В distroless нет curl, проверяет сам бинарник.
func probe() int {
	port := 8080
	if p, err := strconv.Atoi(os.Getenv("PORT")); err == nil && p > 0 && p < 65536 {
		port = p // только число: адрес всегда наш 127.0.0.1
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d/healthz", port), nil) //nolint:gosec // G704: хост фиксирован, порт только числом
	if err != nil {
		return 1
	}
	resp, err := http.DefaultClient.Do(req) //nolint:gosec // G704: см. выше
	if err != nil {
		return 1
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 1
	}
	return 0
}
