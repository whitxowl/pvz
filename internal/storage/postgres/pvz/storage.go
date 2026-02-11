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
