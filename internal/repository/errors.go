package repository

import (
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// Common repository errors
var (
	ErrNotFound      = errors.New("resource not found")
	ErrAlreadyExists = errors.New("resource already exists")
	ErrInvalidInput  = errors.New("invalid input")
	ErrDatabase      = errors.New("database error")
)

// RepositoryError is a custom error type for repository errors
type RepositoryError struct {
	Op     string // Operation that failed
	Entity string // Entity that was being operated on
	Err    error  // Original error
}

// NewError creates a new repository error
func NewError(op, entity string, err error) error {
	return &RepositoryError{
		Op:     op,
		Entity: entity,
		Err:    err,
	}
}

// Error implements the error interface
func (e *RepositoryError) Error() string {
	if e.Entity == "" {
		return fmt.Sprintf("%s: %v", e.Op, e.Err)
	}
	return fmt.Sprintf("%s %s: %v", e.Op, e.Entity, e.Err)
}

// Unwrap returns the underlying error
func (e *RepositoryError) Unwrap() error {
	return e.Err
}

// IsNotFound returns true if the error is a not found error
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}

	var repoErr *RepositoryError
	if errors.As(err, &repoErr) {
		return errors.Is(repoErr.Err, ErrNotFound) || errors.Is(repoErr.Err, gorm.ErrRecordNotFound)
	}

	return errors.Is(err, ErrNotFound) || errors.Is(err, gorm.ErrRecordNotFound)
}

// IsAlreadyExists returns true if the error is an already exists error
func IsAlreadyExists(err error) bool {
	if err == nil {
		return false
	}

	var repoErr *RepositoryError
	if errors.As(err, &repoErr) {
		return errors.Is(repoErr.Err, ErrAlreadyExists)
	}

	return errors.Is(err, ErrAlreadyExists) ||
		strings.Contains(err.Error(), "duplicate key") ||
		strings.Contains(err.Error(), "Duplicate entry")
}
