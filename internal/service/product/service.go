package product

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/whitxowl/pvz.git/internal/domain"
	srvErr "github.com/whitxowl/pvz.git/internal/service/errors"
	storageErr "github.com/whitxowl/pvz.git/internal/storage/errors"
)

type ProductStorage interface {
	CreateProduct(ctx context.Context, productType domain.Type, receptionId string) (*domain.Product, error)
}

type ReceptionStorage interface {
	GetReceptionInProgressID(ctx context.Context, pvzID string) (string, error)
}

type Service struct {
	log              *slog.Logger
	productStorage   ProductStorage
	receptionStorage ReceptionStorage
}

func New(log *slog.Logger, productStorage ProductStorage, receptionStorage ReceptionStorage) *Service {
	return &Service{
		log:              log,
		productStorage:   productStorage,
		receptionStorage: receptionStorage,
	}
}

// TODO: add transactional logic

// Add adds product in in-progress reception if it exists
func (s *Service) Add(ctx context.Context, productType domain.Type, pvzID string) (*domain.Product, error) {
	const op = "service.product.Add"

	log := s.log.With(slog.String("op", op))

	receptionId, err := s.receptionStorage.GetReceptionInProgressID(ctx, pvzID)
	if errors.Is(err, storageErr.ErrNoInProgressReception) {
		log.DebugContext(ctx, "no in-progress reception exists", slog.String("id", pvzID))
		return nil, srvErr.ErrNoInProgressReception
	}
	if err != nil {
		log.ErrorContext(ctx, "error getting reception ID from storage", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	product, err := s.productStorage.CreateProduct(ctx, productType, receptionId)
	if err != nil {
		log.ErrorContext(ctx, "error creating product", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.InfoContext(ctx, "added product", slog.String("reception_id", receptionId))

	return product, nil
}
