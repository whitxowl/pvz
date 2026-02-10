package postgres

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	ErrUniqueViolationCode     = "23505"
	ErrForeignKeyViolationCode = "23503"
)

func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == ErrUniqueViolationCode {
			return true
		}
	}

	return false
}

func IsForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == ErrForeignKeyViolationCode {
			return true
		}
	}

	return false
}

func IsNoRowsError(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}
