package reception

import (
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
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

func (s *Storage) DeleteLastAddedProduct(ctx context.Context, pvzID string) (bool, error) {
	const op = "storage.reception.DeleteLastAddedProduct"

	const query = `
		WITH deleted AS (
			DELETE FROM products
			WHERE id = (
				SELECT p.id FROM products p
				JOIN reception r ON r.id = p.reception_id
				WHERE r.pvz_id = $1
				  AND r.status = 'in_progress'
				ORDER BY p.date_time DESC
				LIMIT 1
			)
			RETURNING id
		)
		SELECT EXISTS(SELECT 1 FROM deleted)
	`

	var deleted bool
	err := s.Db.QueryRow(ctx, query, pvzID).Scan(&deleted)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return deleted, nil
}

func (s *Storage) GetReceptionsByPVZIDs(
	ctx context.Context,
	pvzIDs []string,
	startTime *time.Time,
	endTime *time.Time,
) (map[string][]*domain.Reception, error) {
	const op = "storage.reception.GetReceptionsByPVZIs"

	if len(pvzIDs) == 0 {
		return map[string][]*domain.Reception{}, nil
	}

	builder := sq.Select("id, date_time, pvz_id, status").
		From("reception").
		Where(sq.Eq{"pvz_id": pvzIDs}).
		OrderBy("date_time DESC").
		PlaceholderFormat(sq.Dollar)

	if startTime != nil {
		builder = builder.Where(sq.GtOrEq{"date_time": startTime})
	}
	if endTime != nil {
		builder = builder.Where(sq.LtOrEq{"date_time": endTime})
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	rows, err := s.Db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	result := map[string][]*domain.Reception{}
	var receptionIDs []string

	for rows.Next() {
		var r domain.Reception
		if err := rows.Scan(&r.ID, &r.Date, &r.PvzID, &r.Status); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		result[r.PvzID] = append(result[r.PvzID], &r)
		receptionIDs = append(receptionIDs, r.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if len(receptionIDs) == 0 {
		return result, nil
	}

	productQuery, productArgs, err := sq.Select("id", "date_time", "product_type", "reception_id").
		From("products").
		Where(sq.Eq{"reception_id": receptionIDs}).
		OrderBy("date_time DESC").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	productRows, err := s.Db.Query(ctx, productQuery, productArgs...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer productRows.Close()

	receptionByID := map[string]*domain.Reception{}
	for _, receptions := range result {
		for _, r := range receptions {
			receptionByID[r.ID] = r
		}
	}

	for productRows.Next() {
		var p domain.Product
		if err := productRows.Scan(&p.ID, &p.Date, &p.Type, &p.ReceptionID); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		if r, ok := receptionByID[p.ReceptionID]; ok {
			r.Products = append(r.Products, p)
		}
	}
	if err := productRows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return result, nil
}
