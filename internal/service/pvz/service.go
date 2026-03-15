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
	"github.com/whitxowl/pvz.git/pkg/metrics"
)

type PVZStorage interface {
	CreatePVZ(ctx context.Context, pvz *domain.PVZ) error
	GetPVZList(ctx context.Context, page int, limit int) ([]*domain.PVZ, error)
	GetAll(ctx context.Context) ([]*domain.PVZ, error)
}

type ReceptionStorage interface {
	SetStatusClosed(ctx context.Context, pvzID string) (*domain.Reception, error)
	DeleteLastAddedProduct(ctx context.Context, pvzID string) (bool, error)
	GetReceptionsByPVZIDs(
		ctx context.Context,
		pvzIDs []string,
		startTime *time.Time,
		endTime *time.Time,
	) (map[string][]*domain.Reception, error)
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

	metrics.PVZCreatedTotal.Inc()
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

func (s *Service) GetPVZList(
	ctx context.Context,
	page int,
	limit int,
	startTime *time.Time,
	endTime *time.Time,
) ([]*domain.PVZ, error) {
	const op = "service.pvz.GetPVZList"

	pvzList, err := s.pvzStorage.GetPVZList(ctx, page, limit)
	if err != nil {
		s.log.ErrorContext(ctx, "failed to get pvz list", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	pvzIDs := make([]string, len(pvzList))
	for i, pvz := range pvzList {
		pvzIDs[i] = pvz.ID
	}

	receptionsByPVZ, err := s.recStorage.GetReceptionsByPVZIDs(ctx, pvzIDs, startTime, endTime)
	if err != nil {
		s.log.ErrorContext(ctx, "failed to get receptions by PVZ IDs", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	for _, pvz := range pvzList {
		if receptions, ok := receptionsByPVZ[pvz.ID]; ok {
			for _, r := range receptions {
				pvz.Receptions = append(pvz.Receptions, *r)
			}
		}
	}

	return pvzList, nil
}

func (s *Service) GetAll(ctx context.Context) ([]*domain.PVZ, error) {
	const op = "service.pvz.GetAll"

	pvzs, err := s.pvzStorage.GetAll(ctx)
	if err != nil {
		s.log.ErrorContext(ctx, "failed to get pvzs", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return pvzs, nil
}
