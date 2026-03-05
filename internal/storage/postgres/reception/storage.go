package reception

import (
	"context"
	"fmt"

	"github.com/whitxowl/pvz.git/internal/domain"
	storageErr "github.com/whitxowl/pvz.git/internal/storage/errors"
	"github.com/whitxowl/pvz.git/internal/storage/tx"
	"github.com/whitxowl/pvz.git/pkg/postgres"
)

const statusInProgress = domain.StatusInProgress
const statusClosed = domain.StatusClosed

type Storage struct {
	Db postgres.DB
}

func New(db postgres.DB) *Storage {
	return &Storage{
		Db: db,
	}
}

func (s *Storage) CreateReception(ctx context.Context, pvzID string, status domain.Status) (*domain.Reception, error) {
	const op = "storage.reception.CreateReception"

	const query = `
		INSERT INTO reception(pvz_id, status) 
		VALUES ($1, $2)
		RETURNING id, date_time, pvz_id, status
	`

	var reception domain.Reception
	err := s.Db.QueryRow(ctx, query, pvzID, status).Scan(
		&reception.ID,
		&reception.Date,
		&reception.PvzID,
		&reception.Status,
	)
	if postgres.IsUniqueViolation(err) {
		return nil, fmt.Errorf("%s: %w", op, storageErr.ErrInProgressReceptionExists)
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &reception, nil
}

func (s *Storage) GetReceptionInProgressID(ctx context.Context, pvzID string) (string, error) {
	const op = "storage.reception.GetReceptionInProgressID"

	db := tx.TxFromCtx(ctx, s.Db)

	const query = "SELECT id FROM reception WHERE pvz_id = $1 AND status = $2"

	var id string
	err := db.QueryRow(ctx, query, pvzID, statusInProgress).Scan(&id)
	if postgres.IsNoRowsError(err) {
		return "", fmt.Errorf("%s: %w", op, storageErr.ErrNoInProgressReception)
	}
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Storage) SetStatusClosed(ctx context.Context, pvzID string) (*domain.Reception, error) {
	const op = "storage.reception.SetStatusClosed"

	const query = `
		UPDATE reception
		SET status = $1
		WHERE pvz_id = $2 AND status = $3
		RETURNING id, date_time, pvz_id, status
	`

	var reception domain.Reception
	err := s.Db.QueryRow(ctx, query, statusClosed, pvzID, statusInProgress).Scan(
		&reception.ID,
		&reception.Date,
		&reception.PvzID,
		&reception.Status,
	)
	if postgres.IsNoRowsError(err) {
		return nil, fmt.Errorf("%s: %w", op, storageErr.ErrNoInProgressReception)
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &reception, nil
}

func (s *Storage) CreateProduct(
	ctx context.Context,
	productType domain.Type,
	receptionId string,
) (*domain.Product, error) {
	const op = "storage.reception.CreateProduct"

	db := tx.TxFromCtx(ctx, s.Db)

	const query = `
		INSERT INTO products(product_type, reception_id) 
		VALUES ($1, $2)
		RETURNING id, date_time, product_type, reception_id
	`

	var product domain.Product
	err := db.QueryRow(ctx, query, productType, receptionId).Scan(
		&product.ID,
		&product.Date,
		&product.Type,
		&product.ReceptionID,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &product, nil
}
