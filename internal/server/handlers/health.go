package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/hra42/7x42/internal/server/responses"
	"gorm.io/gorm"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	db *gorm.DB
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(db *gorm.DB) *HealthHandler {
	return &HealthHandler{
		db: db,
	}
}

// Check handles the health check endpoint
func (h *HealthHandler) Check(c *fiber.Ctx) error {
	status := "ok"
	dbStatus := "ok"

	// Check database connection
	sqlDB, err := h.db.DB()
	if err != nil {
		dbStatus = "error: " + err.Error()
		status = "degraded"
	} else if err := sqlDB.Ping(); err != nil {
		dbStatus = "error: " + err.Error()
		status = "degraded"
	}

	return responses.JSON(c, fiber.StatusOK, fiber.Map{
		"status":    status,
		"timestamp": time.Now().Format(time.RFC3339),
		"services": fiber.Map{
			"database": dbStatus,
		},
	})
}
