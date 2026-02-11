package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	authHdlr "github.com/whitxowl/pvz.git/internal/api/v1/auth"
	dummyHdlr "github.com/whitxowl/pvz.git/internal/api/v1/dummy"
	"github.com/whitxowl/pvz.git/internal/config"
	authSrv "github.com/whitxowl/pvz.git/internal/service/auth"
	dummySrv "github.com/whitxowl/pvz.git/internal/service/dummy"
)

type Server struct {
	log          *slog.Logger
	dummyService *dummySrv.Service
	authService  *authSrv.Service
	cfg          *config.HTTPServer

	mu     sync.Mutex
	server *http.Server
}

func New(
	log *slog.Logger,
	dummyService *dummySrv.Service,
	authService *authSrv.Service,
	cfg config.HTTPServer,
) *Server {
	return &Server{
		log:          log,
		dummyService: dummyService,
		authService:  authService,
		cfg:          &cfg,
	}
}

func (s *Server) MustRun(ctx context.Context) {
	if err := s.Run(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		panic("failed to start server: " + err.Error())
	}
}

func (s *Server) Run(ctx context.Context) error {
	const op = "httpsrv.Run"

	log := s.log.With(
		slog.String("op", op),
		slog.String("address", s.cfg.Address),
	)

	log.InfoContext(ctx, "starting http server")

	dummyHdlr := dummyHdlr.New(s.dummyService)
	authHdlr := authHdlr.New(s.authService)

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(ginLogger(s.log))

	base := router.Group("/")

	dummyHdlr.RegisterRoutes(base)
	authHdlr.RegisterRoutes(base)

	srv := &http.Server{
		Addr:         s.cfg.Address,
		Handler:      router,
		ReadTimeout:  s.cfg.Timeout,
		WriteTimeout: s.cfg.Timeout,
		IdleTimeout:  s.cfg.IdleTimeout,
	}

	s.mu.Lock()
	s.server = srv
	s.mu.Unlock()

	log.InfoContext(ctx, "http server started")

	return srv.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	const op = "httpsrv.Stop"

	log := s.log.With(slog.String("op", op))

	log.InfoContext(ctx, "stopping http server")

	s.mu.Lock()
	server := s.server
	s.mu.Unlock()

	if server == nil {
		return nil
	}

	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.InfoContext(ctx, "http server stopped")

	return nil
}

func ginLogger(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		end := time.Now()
		latency := end.Sub(start)

		log.Info("HTTP request",
			slog.Int("status", c.Writer.Status()),
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.String("query", query),
			slog.String("ip", c.ClientIP()),
			slog.Duration("latency", latency),
			slog.String("user_agent", c.Request.UserAgent()),
		)
	}
}
