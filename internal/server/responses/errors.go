package responses

import (
	"github.com/gofiber/fiber/v2"
)

// Error sends an error response
func Error(c *fiber.Ctx, status int, message string) error {
	return JSON(c, status, fiber.Map{
		"success": false,
		"error":   message,
	})
}

// BadRequest sends a bad request error response
func BadRequest(c *fiber.Ctx, message string) error {
	return Error(c, fiber.StatusBadRequest, message)
}

// NotFound sends a not found error response
func NotFound(c *fiber.Ctx, message string) error {
	return Error(c, fiber.StatusNotFound, message)
}

// Unauthorized sends an unauthorized error response
func Unauthorized(c *fiber.Ctx, message string) error {
	return Error(c, fiber.StatusUnauthorized, message)
}

// Forbidden sends a forbidden error response
func Forbidden(c *fiber.Ctx, message string) error {
	return Error(c, fiber.StatusForbidden, message)
}

// InternalError sends an internal server error response
func InternalError(c *fiber.Ctx, message string) error {
	return Error(c, fiber.StatusInternalServerError, message)
}
