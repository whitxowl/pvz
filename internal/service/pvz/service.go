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
)

type PVZStorage interface {
	CreatePVZ(ctx context.Context, pvz *domain.PVZ) error
}

type ReceptionStorage interface {
	SetStatusClosed(ctx context.Context, pvzID string) (*domain.Reception, error)
	DeleteLastAddedProduct(ctx context.Context, pvzID string) (bool, error)
}

type Service struct {
	log        *slog.Logger
	pvzStorage PVZStorage
	recStorage ReceptionStorage
}

// New returns pvz service instance
func New(
	log *slog.Logger,
	pvzStorage PVZStorage,
	recStorage ReceptionStorage,
) *Service {
	return &Service{
		log:        log,
		pvzStorage: pvzStorage,
		recStorage: recStorage,
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

func (s *Service) CloseReception(ctx context.Context, pvzID string) (*domain.Reception, error) {
	const op = "storage.pvz.CloseReception"

	log := s.log.With(
		slog.String("op", op),
		slog.String("pvzID", pvzID),
	)

	reception, err := s.recStorage.SetStatusClosed(ctx, pvzID)
	if errors.Is(err, storageErr.ErrNoInProgressReception) {
		log.DebugContext(ctx, "in_progress reception not found", slog.String("pvzID", pvzID))
		return nil, srvErr.ErrNoInProgressReception
	}
	if err != nil {
		log.ErrorContext(ctx, "failed to close reception", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.InfoContext(ctx, "close reception", slog.String("id", reception.ID))

	reception.Status = domain.StatusInProgress

	return reception, nil
}

func (s *Service) DeleteLastProduct(ctx context.Context, pvzID string) (bool, error) {
	const op = "storage.pvz.DeleteLastProduct"

	log := s.log.With(
		slog.String("op", op),
		slog.String("pvzID", pvzID),
	)

	deleted, err := s.recStorage.DeleteLastAddedProduct(ctx, pvzID)
	if err != nil {
		log.ErrorContext(ctx, "failed to delete last product", slog.String("error", err.Error()))
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return deleted, nil
}
