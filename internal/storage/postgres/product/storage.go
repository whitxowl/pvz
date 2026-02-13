package product

import (
	"context"
	"fmt"

	"github.com/whitxowl/pvz.git/internal/domain"
	"github.com/whitxowl/pvz.git/pkg/postgres"
)

type Storage struct {
	Db postgres.DB
}

func New(db postgres.DB) *Storage {
	return &Storage{
		Db: db,
	}
}

func (s *Storage) CreateProduct(ctx context.Context, productType domain.Type, receptionId string) (*domain.Product, error) {
	const op = "storage.product.CreateProduct"

	const query = `
		INSERT INTO products(product_type, reception_id) 
		VALUES ($1, $2)
		RETURNING id, date_time, product_type, reception_id
	`

	var product domain.Product
	err := s.Db.QueryRow(ctx, query, productType, receptionId).Scan(
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
