package dummy

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/whitxowl/pvz.git/internal/domain"
	srvErr "github.com/whitxowl/pvz.git/internal/service/errors"
	"github.com/whitxowl/pvz.git/pkg/jwt"
)

type Service struct {
	log          *slog.Logger
	tokenManager *jwt.TokenManager
}

func New(log *slog.Logger, tokenManager *jwt.TokenManager) *Service {
	return &Service{
		log:          log,
		tokenManager: tokenManager,
	}
}

func (s *Service) Login(ctx context.Context, role domain.Role) (string, error) {
	const op = "service.dummy.Login"

	log := s.log.With(slog.String("op", op))

	if !role.IsValid() {
		log.DebugContext(ctx, "invalid role")
		return "", srvErr.ErrInvalidRole
	}

	accessToken, err := s.tokenManager.GenerateToken(domain.TokenClaims{
		Role: role,
	})
	if err != nil {
		log.ErrorContext(ctx, "failed to generate token", slog.String("error", err.Error()))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.InfoContext(ctx, "successful dummy login")

	return accessToken, nil
}
