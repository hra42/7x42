package database

import (
	"github.com/hra42/7x42/internal/models"
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	// Enable PostgreSQL extensions
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		return err
	}
	// Run migrations
	return db.AutoMigrate(
		&models.Chat{},
		&models.Message{},
	)
}
