package reception

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
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &reception, nil
}

func (s *Storage) ReceptionInProgressExists(ctx context.Context, pvzID string) (bool, error) {
	const op = "storage.reception.ReceptionInProgress"

	const query = "SELECT EXISTS(SELECT 1 FROM reception WHERE pvz_id = $1 AND status = $2)"
	status := domain.StatusInProgress

	var exists bool
	err := s.Db.QueryRow(ctx, query, pvzID, status).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return exists, nil
}
