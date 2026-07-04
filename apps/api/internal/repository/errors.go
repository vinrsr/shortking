package repository

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

var ErrConflict = errors.New("repository: unique constraint violation")

const pgUniqueViolationCode = "23505"

func translateError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolationCode {
		return ErrConflict
	}
	return err
}
