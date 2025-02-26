package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// Repository is the base repository interface that all repositories should implement
type Repository interface {
	DB() *gorm.DB
	WithContext(ctx context.Context) Repository
	WithTimeout(timeout time.Duration) (Repository, context.CancelFunc)
}

// BaseRepository implements the Repository interface
type BaseRepository struct {
	db *gorm.DB
}

// NewBaseRepository creates a new base repository
func NewBaseRepository(db *gorm.DB) *BaseRepository {
	return &BaseRepository{
		db: db,
	}
}

// DB returns the database connection
func (r *BaseRepository) DB() *gorm.DB {
	return r.db
}

// WithContext returns a new repository with the given context
func (r *BaseRepository) WithContext(ctx context.Context) Repository {
	return &BaseRepository{
		db: r.db.WithContext(ctx),
	}
}

// WithTimeout returns a new repository with a timeout context
func (r *BaseRepository) WithTimeout(timeout time.Duration) (Repository, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	return &BaseRepository{
		db: r.db.WithContext(ctx),
	}, cancel
}

// RunInTransaction runs the given function in a transaction
func RunInTransaction(db *gorm.DB, fn func(tx *gorm.DB) error) error {
	return db.Transaction(fn)
}
