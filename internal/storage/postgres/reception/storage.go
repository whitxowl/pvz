package reception

import (
	"context"
	"fmt"

	"github.com/whitxowl/pvz.git/internal/domain"
	storageErr "github.com/whitxowl/pvz.git/internal/storage/errors"
	"github.com/whitxowl/pvz.git/pkg/postgres"
)

const statusInProgress = domain.StatusInProgress

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

	var exists bool
	err := s.Db.QueryRow(ctx, query, pvzID, statusInProgress).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return exists, nil
}

func (s *Storage) GetReceptionInProgressID(ctx context.Context, pvzID string) (string, error) {
	const op = "storage.reception.GetReceptionInProgressID"

	const query = "SELECT id FROM reception WHERE pvz_id = $1 AND status = $2"

	var id string
	err := s.Db.QueryRow(ctx, query, pvzID, statusInProgress).Scan(&id)
	if postgres.IsNoRowsError(err) {
		return "", fmt.Errorf("%s: %w", op, storageErr.ErrNoInProgressReception)
	}
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}
