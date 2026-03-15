package metricserver

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/whitxowl/pvz.git/internal/config"
)

type Server struct {
	log    *slog.Logger
	server *http.Server
}

func New(log *slog.Logger, cfg config.MetricsServer) *Server {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	return &Server{
		log: log,
		server: &http.Server{
			Addr:    cfg.Address(),
			Handler: mux,
		},
	}
}

func (s *Server) MustRun(ctx context.Context) {
	if err := s.Run(ctx); err != nil {
		panic("failed to start metrics server: " + err.Error())
	}
}

func (s *Server) Run(ctx context.Context) error {
	s.log.InfoContext(ctx, "starting metrics server", slog.String("address", s.server.Addr))

	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("metricserver.Run: %w", err)
	}

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	s.log.InfoContext(ctx, "stopping metrics server")
	return s.server.Shutdown(ctx)
}
