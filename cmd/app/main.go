package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/whitxowl/pvz.git/internal/app"
	"github.com/whitxowl/pvz.git/internal/config"
)

const (
	shutdownTimeout = 10 * time.Second
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("application panic", "err", r)
		}
	}()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	cfg := config.MustLoadConfig()

	log := config.NewLogger(cfg.Env)
	slog.SetDefault(log)

	log.Info("starting application")

	application := app.New(ctx, log, cfg)

	go application.Srv.MustRun(ctx)
	go application.GRPCSrv.MustRun(ctx)
	go application.MetricsSrv.MustRun(ctx)

	<-ctx.Done()

	log.Info("received stop signal")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		if err := application.Srv.Stop(shutdownCtx); err != nil {
			log.Error("failed to stop http server gracefully", "err", err)
		}
	}()

	go func() {
		defer wg.Done()
		application.GRPCSrv.Stop(shutdownCtx)
	}()

	go func() {
		defer wg.Done()
		if err := application.MetricsSrv.Stop(shutdownCtx); err != nil {
			log.Error("failed to stop metrics server gracefully", "err", err)
		}
	}()

	wg.Wait()
	log.Info("application stopped gracefully")
}
