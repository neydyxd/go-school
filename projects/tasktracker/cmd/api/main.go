// Команда api: точка сборки сервиса. Здесь только wiring (DI руками) и жизненный цикл.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"golang.org/x/crypto/bcrypt"

	"goschool/tasktracker/internal/auth"
	"goschool/tasktracker/internal/config"
	api "goschool/tasktracker/internal/http"
	"goschool/tasktracker/internal/jobs"
	"goschool/tasktracker/internal/repository"
	"goschool/tasktracker/internal/service"
	"goschool/tasktracker/migrations"
)

func main() {
	if err := run(); err != nil {
		slog.Error("сервис остановлен с ошибкой", "err", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load(os.Getenv)
	if err != nil {
		return fmt.Errorf("конфиг:\n%w", err)
	}
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel}))
	slog.SetDefault(log)
	log.Info("старт", "config", cfg) // секреты скрывает Config.LogValue

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
	// ponytail: миграции на старте удобны в учебном проекте; в проде их гоняют отдельным шагом деплоя
	if err := migrations.Up(ctx, stdlib.OpenDBFromPool(pool)); err != nil {
		return err
	}
	if err := migrations.RiverUp(ctx, pool); err != nil {
		return err
	}

	// Модуль 09: фоновые задачи. Клиент работает в этом же процессе, что и API;
	// в проде воркеры часто выносят в отдельный процесс (cmd/worker) с тем же NewClient.
	riverClient, err := jobs.NewClient(pool, log, jobs.LogMailer{Log: log})
	if err != nil {
		return fmt.Errorf("river: %w", err)
	}
	// Отдельный контекст: останавливаем River сами, ПОСЛЕ HTTP-сервера, чтобы
	// запросы, которые ещё дорабатывают, могли поставить задачи.
	if err := riverClient.Start(context.WithoutCancel(ctx)); err != nil {
		return fmt.Errorf("river start: %w", err)
	}
	// Мягкая остановка River на любом выходе из run: defer выполнится после
	// остановки HTTP и ДО pool.Close (defer-ы идут в обратном порядке), иначе
	// Close ждал бы соединение, которое держит River. Новые задачи не берём,
	// текущие дорабатывают до SoftStopTimeout, потом их ctx отменяется.
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout+15*time.Second)
		defer cancel()
		if err := riverClient.Stop(stopCtx); err != nil {
			log.Error("остановка River", "err", err)
			return
		}
		log.Info("остановлен")
	}()

	tokens := auth.NewTokens([]byte(cfg.JWTSecret), cfg.AccessTTL)
	users, refresh := repository.NewUsers(pool), repository.NewRefreshTokens(pool)
	projects, tasks := repository.NewProjects(pool), repository.NewTasks(pool)

	router := api.NewRouter(api.Deps{
		Log:      log,
		DB:       pool,
		Tokens:   tokens,
		Auth:     service.NewAuthService(users, refresh, auth.Hasher{Cost: bcrypt.DefaultCost}, tokens, cfg.AccessTTL, cfg.RefreshTTL),
		Projects: service.NewProjectService(projects),
		Tasks:    service.NewTaskService(projects, tasks, repository.NewTxManager(pool).WithRiver(riverClient)),
	})

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
	}

	log.Info("останавливаюсь, дожидаюсь текущих запросов", "timeout", cfg.ShutdownTimeout)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	if err := <-errCh; !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	log.Info("HTTP остановлен")
	return nil
}
