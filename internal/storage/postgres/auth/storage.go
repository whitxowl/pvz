package auth

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

// New returns user storage instance
func New(db postgres.DB) *Storage {
	return &Storage{
		Db: db,
	}
}

// RegisterUser registers user in database
// Returns generated UUID
func (s *Storage) RegisterUser(ctx context.Context, user *domain.User) (string, error) {
	const op = "storage.auth.RegisterUser"

	const query = "INSERT INTO users (email, pass_hash, user_role) VALUES ($1, $2, $3) RETURNING id"

	var id string
	err := s.Db.QueryRow(ctx, query, user.Email, user.PassHash, user.Role).Scan(&id)
	if postgres.IsUniqueViolation(err) {
		return "", fmt.Errorf("%s: %w", op, storageErr.ErrUserExists)
	}
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

// User returns user
func (s *Storage) User(ctx context.Context, email string) (*domain.User, error) {
	const op = "storage.auth.User"

	const query = "SELECT * FROM users WHERE email = $1"

	user := &domain.User{}
	err := s.Db.QueryRow(ctx, query, email).Scan(&user.ID, &user.Email, &user.PassHash, &user.Role)
	if postgres.IsNoRowsError(err) {
		return nil, fmt.Errorf(`%s: %w`, op, storageErr.ErrUserNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}
