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
	User(ctx context.Context, email string) (*domain.User, error)
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

// Login cheks whether user is registered
// Returns token
func (s *Service) Login(ctx context.Context, email, password string) (string, error) {
	const op = "service.auth.Login"

	log := s.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	user, err := s.authStorage.User(ctx, email)
	if errors.Is(err, storageErr.ErrUserNotFound) {
		log.DebugContext(ctx, "user does not exist")
		return "", srvErr.ErrInvalidCredentials
	}
	if err != nil {
		log.ErrorContext(ctx, "failed to get user", slog.String("error", err.Error()))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	if !s.passHasher.Verify(password, user.PassHash) {
		log.DebugContext(ctx, "invalid credentials")
		return "", srvErr.ErrInvalidCredentials
	}

	accessToken, err := s.tokenManager.GenerateToken(domain.TokenClaims{
		Role: user.Role,
	})

	if err != nil {
		log.ErrorContext(ctx, "failed to generate token", slog.String("error", err.Error()))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.InfoContext(ctx, "user logged in", slog.String("email", email))

	return accessToken, nil
}

// ValidateToken checks token and returns claims
func (s *Service) ValidateToken(ctx context.Context, token string) (*domain.TokenClaims, error) {
	return s.tokenManager.ValidateToken(token)
}
