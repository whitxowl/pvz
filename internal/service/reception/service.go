package reception

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/whitxowl/pvz.git/internal/domain"
	srvErr "github.com/whitxowl/pvz.git/internal/service/errors"
	storageErr "github.com/whitxowl/pvz.git/internal/storage/errors"
	"github.com/whitxowl/pvz.git/pkg/metrics"
)

type ReceptionStorage interface {
	CreateReception(ctx context.Context, pvzID string, status domain.Status) (*domain.Reception, error)
	GetReceptionInProgressID(ctx context.Context, pvzID string) (string, error)
	CreateProduct(ctx context.Context, productType domain.Type, receptionId string) (*domain.Product, error)
}

type TxManager interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type Service struct {
	log              *slog.Logger
	receptionStorage ReceptionStorage
	txManager        TxManager
}

// New returns reception service instance
func New(
	log *slog.Logger,
	receptionStorage ReceptionStorage,
	txManager TxManager,
) *Service {
	return &Service{
		log:              log,
		receptionStorage: receptionStorage,
		txManager:        txManager,
	}
}

func (s *Service) CreateReception(ctx context.Context, pvzID string) (*domain.Reception, error) {
	const op = "storage.pvz.CreateReception"

	log := s.log.With(slog.String("op", op))

	reception, err := s.receptionStorage.CreateReception(ctx, pvzID, domain.StatusInProgress)
	if err != nil {
		if errors.Is(err, storageErr.ErrInProgressReceptionExists) {
			return nil, srvErr.ErrInProgressReceptionExists
		}

		log.ErrorContext(ctx, "failed to create reception",
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.InfoContext(ctx, "created reception", slog.String("id", pvzID))

	metrics.ReceptionsCreatedTotal.Inc()
	return reception, nil
}

// Add adds product in in-progress reception if it exists
func (s *Service) Add(ctx context.Context, productType domain.Type, pvzID string) (*domain.Product, error) {
	const op = "service.product.Add"

	log := s.log.With(slog.String("op", op))

	var receptionID string
	var product *domain.Product

	err := s.txManager.WithTx(ctx, func(ctx context.Context) error {
		id, err := s.receptionStorage.GetReceptionInProgressID(ctx, pvzID)
		if errors.Is(err, storageErr.ErrNoInProgressReception) {
			log.DebugContext(ctx, "no in-progress reception exists", slog.String("id", pvzID))
			return srvErr.ErrNoInProgressReception
		}
		if err != nil {
			log.ErrorContext(ctx, "error getting reception ID from storage", slog.String("error", err.Error()))
			return fmt.Errorf("%s: %w", op, err)
		}

		receptionID = id

		product, err = s.receptionStorage.CreateProduct(ctx, productType, id)
		if err != nil {
			log.ErrorContext(ctx, "error creating product", slog.String("error", err.Error()))
			return fmt.Errorf("%s: %w", op, err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	log.InfoContext(ctx, "added product",
		slog.String("reception_id", receptionID),
		slog.String("pvzID", pvzID))

	metrics.ProductsAddedTotal.Inc()
	return product, nil
}
