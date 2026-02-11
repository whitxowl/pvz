package pvz

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/whitxowl/pvz.git/internal/domain"
	srvErr "github.com/whitxowl/pvz.git/internal/service/errors"
	storageErr "github.com/whitxowl/pvz.git/internal/storage/errors"
	"github.com/whitxowl/pvz.git/pkg/hash"
	"github.com/whitxowl/pvz.git/pkg/jwt"
)

type PVZStorage interface {
	CreatePVZ(ctx context.Context, pvz *domain.PVZ) error
}

type Service struct {
	log          *slog.Logger
	pvzStorage   PVZStorage
	tokenManager *jwt.TokenManager
	passHasher   *hash.PasswordHasher
}

// New returns pvz service instance
func New(
	log *slog.Logger,
	pvzStorage PVZStorage,
	tokenManager *jwt.TokenManager,
	passHasher *hash.PasswordHasher,
) *Service {
	return &Service{
		log:          log,
		pvzStorage:   pvzStorage,
		tokenManager: tokenManager,
		passHasher:   passHasher,
	}
}

func (s *Service) CreatePVZ(
	ctx context.Context,
	id string,
	registrationDate *time.Time,
	city domain.City,
) (*domain.PVZ, error) {
	const op = "storage.pvz.CreatePVZ"

	log := s.log.With(
		slog.String("op", op),
		slog.String("id", id),
	)

	if !city.IsValid() {
		log.DebugContext(ctx, "invalid city")
		return nil, srvErr.ErrInvalidCity
	}

	pvz := &domain.PVZ{
		ID:               id,
		RegistrationDate: registrationDate,
		City:             city,
	}

	err := s.pvzStorage.CreatePVZ(ctx, pvz)
	if errors.Is(err, storageErr.ErrPVZExists) {
		log.DebugContext(ctx, "pvz already exists", slog.String("id", id))
		return nil, srvErr.ErrPVZExists
	}
	if err != nil {
		log.ErrorContext(ctx, "failed to create pvz", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.InfoContext(ctx, "created pvz", slog.String("id", id))

	return pvz, nil
}
