package responses

import (
	"github.com/gofiber/fiber/v2"
)

// JSON sends a JSON response
func JSON(c *fiber.Ctx, status int, data interface{}) error {
	return c.Status(status).JSON(data)
}

// Success sends a success response
func Success(c *fiber.Ctx, data interface{}) error {
	return JSON(c, fiber.StatusOK, fiber.Map{
		"success": true,
		"data":    data,
	})
}

// Created sends a created response
func Created(c *fiber.Ctx, data interface{}) error {
	return JSON(c, fiber.StatusCreated, fiber.Map{
		"success": true,
		"data":    data,
	})
}

// NoContent sends a no content response
func NoContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}
