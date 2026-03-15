package app

import (
	"context"
	"log/slog"

	grpcsrv "github.com/whitxowl/pvz.git/internal/app/grpc"
	"github.com/whitxowl/pvz.git/internal/app/metricserver"
	"github.com/whitxowl/pvz.git/internal/app/server"
	"github.com/whitxowl/pvz.git/internal/config"
	authService "github.com/whitxowl/pvz.git/internal/service/auth"
	dummyService "github.com/whitxowl/pvz.git/internal/service/dummy"
	pvzService "github.com/whitxowl/pvz.git/internal/service/pvz"
	rcptService "github.com/whitxowl/pvz.git/internal/service/reception"
	authStorage "github.com/whitxowl/pvz.git/internal/storage/postgres/auth"
	pvzStorage "github.com/whitxowl/pvz.git/internal/storage/postgres/pvz"
	rcptStorage "github.com/whitxowl/pvz.git/internal/storage/postgres/reception"
	"github.com/whitxowl/pvz.git/internal/storage/tx"
	"github.com/whitxowl/pvz.git/pkg/hash"
	"github.com/whitxowl/pvz.git/pkg/jwt"
	"github.com/whitxowl/pvz.git/pkg/metrics"
	"github.com/whitxowl/pvz.git/pkg/postgres"
)

type App struct {
	Srv        *server.Server
	GRPCSrv    *grpcsrv.Server
	MetricsSrv *metricserver.Server
}

func New(ctx context.Context, log *slog.Logger, cfg *config.Config) *App {
	pgPool, err := postgres.NewPool(ctx, cfg.StorageConfig.DSN(), postgres.WithMaxConnections(int32(cfg.StorageConfig.MaxConnections)))
	if err != nil {
		panic("failed to connect to database" + err.Error())
	}

	authStore := authStorage.New(pgPool)
	pvzStore := pvzStorage.New(pgPool)
	rcptStore := rcptStorage.New(pgPool)

	txManager := tx.NewTxManager(pgPool)

	tokenManager := jwt.NewTokenManager(
		cfg.JWTConfig.SecretKey,
		cfg.JWTConfig.AccessTokenDuration,
	)
	passwordHasher := hash.NewPasswordHasher()

	dummySrv := dummyService.New(log.WithGroup("service.dummy"), tokenManager)
	authSrv := authService.New(log.WithGroup("service.auth"), authStore, tokenManager, passwordHasher)
	pvzSrv := pvzService.New(log.WithGroup("service.pvz"), pvzStore, rcptStore)
	rcptSrv := rcptService.New(log.WithGroup("service.pvz"), rcptStore, txManager)

	srv := server.New(log, dummySrv, authSrv, pvzSrv, rcptSrv, cfg.HTTPServer)
	grpcSrv := grpcsrv.New(log.WithGroup("grpc"), pvzSrv, cfg.GRPCServer)

	metrics.MustRegister()
	metricsSrv := metricserver.New(log.WithGroup("metrics"), cfg.MetricsServer)

	return &App{
		Srv:        srv,
		GRPCSrv:    grpcSrv,
		MetricsSrv: metricsSrv,
	}
}
