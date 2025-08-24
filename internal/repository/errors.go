package repository

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrTaskNotFound      = errors.New("task not found")
	ErrTaskAlreadyExists = errors.New("task already exists")
	ErrDatabaseConnection = errors.New("database connection error")
	ErrInvalidData       = errors.New("invalid data provided")
	ErrConstraintViolation = errors.New("database constraint violation")
	ErrTransactionFailed = errors.New("transaction failed")
)

type RepositoryError struct {
	Op  string
	Err error
}

func (e *RepositoryError) Error() string {
	if e.Op == "" {
		return e.Err.Error()
	}
	return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

func (e *RepositoryError) Unwrap() error {
	return e.Err
}
 
func WrapError(op string, err error) error {
	if err == nil {
		return nil
	}
	return &RepositoryError{Op: op, Err: err}
}

func HandlePgxError(op string, err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return WrapError(op, ErrTaskNotFound)
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return WrapError(op, ErrTaskAlreadyExists)
		case "23503":
		case "23502":
		case "23514":
			return WrapError(op, ErrConstraintViolation)
		case "08000", "08003", "08006":
			return WrapError(op, ErrDatabaseConnection)
		case "22P02":
			return WrapError(op, ErrInvalidData)
		default:
			return WrapError(op, fmt.Errorf("database error [%s]: %s", pgErr.Code, pgErr.Message))
		}
	}

	return WrapError(op, err)
}

func IsNotFoundError(err error) bool {
	var repoErr *RepositoryError
	if errors.As(err, &repoErr) {
		return errors.Is(repoErr.Err, ErrTaskNotFound)
	}
	return errors.Is(err, ErrTaskNotFound)
}

func IsConstraintError(err error) bool {
	var repoErr *RepositoryError
	if errors.As(err, &repoErr) {
		return errors.Is(repoErr.Err, ErrConstraintViolation) || 
			   errors.Is(repoErr.Err, ErrTaskAlreadyExists)
	}
	return errors.Is(err, ErrConstraintViolation) || 
		   errors.Is(err, ErrTaskAlreadyExists)
}

func IsConnectionError(err error) bool {
	var repoErr *RepositoryError
	if errors.As(err, &repoErr) {
		return errors.Is(repoErr.Err, ErrDatabaseConnection)
	}
	return errors.Is(err, ErrDatabaseConnection)
}