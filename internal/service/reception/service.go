package reception

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/whitxowl/pvz.git/internal/domain"
	srvErr "github.com/whitxowl/pvz.git/internal/service/errors"
)

type ReceptionStorage interface {
	ReceptionInProgressExists(ctx context.Context, pvzID string) (bool, error)
	CreateReception(ctx context.Context, pvzID string, status domain.Status) (*domain.Reception, error)
}

type Service struct {
	log              *slog.Logger
	receptionStorage ReceptionStorage
}

// New returns reception service instance
func New(
	log *slog.Logger,
	receptionStorage ReceptionStorage,
) *Service {
	return &Service{
		log:              log,
		receptionStorage: receptionStorage,
	}
}

// CreateReception checks whether there is any in_progress reception
// for pvz with pvzID and creates reception with in_progress status
func (s *Service) CreateReception(ctx context.Context, pvzID string) (*domain.Reception, error) {
	const op = "storage.pvz.CreateReception"

	log := s.log.With(slog.String("op", op))

	exists, err := s.receptionStorage.ReceptionInProgressExists(ctx, pvzID)
	if err != nil {
		log.ErrorContext(ctx, "failed to check if reception exists", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if exists {
		log.DebugContext(ctx, "in-progress reception exists", slog.String("id", pvzID))
		return nil, srvErr.ErrInProgressReceptionExists
	}

	status := domain.StatusInProgress
	reception, err := s.receptionStorage.CreateReception(ctx, pvzID, status)
	if err != nil {
		log.ErrorContext(ctx, "failed to create reception", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.InfoContext(ctx, "created reception", slog.String("id", pvzID))

	return reception, nil
}
