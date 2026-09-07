package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"

	"avito-kitchen/internal/domain"
)

// https://www.postgresql.org/docs/current/errcodes-appendix.html
const (
	pgUniqueViolation     = "23505"
	pgForeignKeyViolation = "23503"
	pgCheckViolation      = "23514"
)

func mapError(err error, notFoundErr error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, sql.ErrNoRows) {
		if notFoundErr != nil {
			return notFoundErr
		}
		return domain.ErrNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgUniqueViolation:
			return fmt.Errorf("%w: %s", domain.ErrAlreadyExists, pgErr.ConstraintName)
		case pgForeignKeyViolation, pgCheckViolation:
			return fmt.Errorf("%w: %s", domain.ErrInvalidInput, pgErr.ConstraintName)
		}
	}

	return fmt.Errorf("%w: %v", domain.ErrInternal, err)
}
