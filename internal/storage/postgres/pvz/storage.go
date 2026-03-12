package pvz

import (
	"context"
	"fmt"

	"github.com/whitxowl/pvz.git/internal/domain"
	storageErr "github.com/whitxowl/pvz.git/internal/storage/errors"
	"github.com/whitxowl/pvz.git/pkg/postgres"
)

type Storage struct {
	Db postgres.DB
}

// New returns pvz storage instance
func New(db postgres.DB) *Storage {
	return &Storage{
		Db: db,
	}
}

func (s *Storage) CreatePVZ(ctx context.Context, pvz *domain.PVZ) error {
	const op = "storage.pvz.CreatePVZ"

	const query = "INSERT INTO pvz(id, registration_date, city) VALUES ($1, $2, $3)"

	_, err := s.Db.Exec(ctx, query, pvz.ID, pvz.RegistrationDate, pvz.City)
	if postgres.IsUniqueViolation(err) {
		return fmt.Errorf("%s: %w", op, storageErr.ErrPVZExists)
	}
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) GetPVZList(ctx context.Context, page int, limit int) ([]*domain.PVZ, error) {
	const op = "storage.pvz.GetPVZList"

	const query = `
        SELECT id, registration_date, city FROM pvz
        ORDER BY registration_date DESC
        LIMIT $1 OFFSET $2
    `

	rows, err := s.Db.Query(ctx, query, limit, (page-1)*limit)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var pvzList []*domain.PVZ
	for rows.Next() {
		var pvz domain.PVZ
		if err := rows.Scan(&pvz.ID, &pvz.RegistrationDate, &pvz.City); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		pvzList = append(pvzList, &pvz)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return pvzList, nil
}

func (s *Storage) GetAll(ctx context.Context) ([]*domain.PVZ, error) {
	const op = "storage.pvz.GetAll"

	const query = "SELECT id, registration_date, city FROM pvz"

	rows, err := s.Db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var pvzs []*domain.PVZ
	for rows.Next() {
		var pvz domain.PVZ
		if err := rows.Scan(&pvz.ID, &pvz.RegistrationDate, &pvz.City); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		pvzs = append(pvzs, &pvz)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return pvzs, nil
}
