package app

import (
	"context"
	"log/slog"

	"github.com/whitxowl/pvz.git/internal/app/server"
	"github.com/whitxowl/pvz.git/internal/config"
	authService "github.com/whitxowl/pvz.git/internal/service/auth"
	dummyService "github.com/whitxowl/pvz.git/internal/service/dummy"
	pvzService "github.com/whitxowl/pvz.git/internal/service/pvz"
	authStorage "github.com/whitxowl/pvz.git/internal/storage/postgres/auth"
	pvzStorage "github.com/whitxowl/pvz.git/internal/storage/postgres/pvz"
	"github.com/whitxowl/pvz.git/pkg/hash"
	"github.com/whitxowl/pvz.git/pkg/jwt"
	"github.com/whitxowl/pvz.git/pkg/postgres"
)

type App struct {
	Srv *server.Server
}

func New(ctx context.Context, log *slog.Logger, cfg *config.Config) *App {
	pgPool, err := postgres.NewPool(ctx, cfg.StorageConfig.DSN(), postgres.WithMaxConnections(int32(cfg.StorageConfig.MaxConnections)))
	if err != nil {
		panic("failed to connect to database" + err.Error())
	}

	authStore := authStorage.New(pgPool)
	pvzStore := pvzStorage.New(pgPool)

	tokenManager := jwt.NewTokenManager(
		cfg.JWTConfig.SecretKey,
		cfg.JWTConfig.AccessTokenDuration,
	)
	passwordHasher := hash.NewPasswordHasher()

	dummySrv := dummyService.New(log.WithGroup("service.dummy"), tokenManager)
	authSrv := authService.New(log.WithGroup("service.auth"), authStore, tokenManager, passwordHasher)
	pvzSrv := pvzService.New(log.WithGroup("service.pvz"), pvzStore, tokenManager, passwordHasher)

	srv := server.New(log, dummySrv, authSrv, pvzSrv, cfg.HTTPServer)

	return &App{
		Srv: srv,
	}
}
