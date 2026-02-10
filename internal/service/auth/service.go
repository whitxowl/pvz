package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/whitxowl/pvz.git/internal/domain"
	srvErr "github.com/whitxowl/pvz.git/internal/service/errors"
	storageErr "github.com/whitxowl/pvz.git/internal/storage/errors"
	"github.com/whitxowl/pvz.git/pkg/hash"
	"github.com/whitxowl/pvz.git/pkg/jwt"
)

type AuthStorage interface {
	RegisterUser(ctx context.Context, user *domain.User) (string, error)
}

type Service struct {
	log          *slog.Logger
	authStorage  AuthStorage
	tokenManager *jwt.TokenManager
	passHasher   *hash.PasswordHasher
}

// New returns auth service instance
func New(
	log *slog.Logger,
	authStorage AuthStorage,
	tokenManager *jwt.TokenManager,
	passHasher *hash.PasswordHasher,
) *Service {
	return &Service{
		log:          log,
		authStorage:  authStorage,
		tokenManager: tokenManager,
		passHasher:   passHasher,
	}
}

// RegisterUser registers user
func (s *Service) RegisterUser(ctx context.Context, email, password string, role domain.Role) (*domain.User, error) {
	const op = "service.auth.RegisterUser"

	log := s.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	if !role.IsValid() {
		log.DebugContext(ctx, "invalid role")
		return nil, srvErr.ErrInvalidRole
	}

	passHash, err := s.passHasher.Hash(password)
	if err != nil {
		log.ErrorContext(ctx, "failed to hash password", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	user := &domain.User{
		Email:    email,
		PassHash: passHash,
		Role:     role,
	}

	id, err := s.authStorage.RegisterUser(ctx, user)
	if errors.Is(err, storageErr.ErrUserExists) {
		log.DebugContext(ctx, "user already exists")
		return nil, srvErr.ErrUserExists
	}
	if err != nil {
		log.ErrorContext(ctx, "failed to register user", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	user.ID = id

	log.InfoContext(ctx, "user registered")

	return user, nil
}
