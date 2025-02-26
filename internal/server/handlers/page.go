package handlers

import (
	"github.com/gofiber/fiber/v2"
)

// PageHandler handles page rendering requests
type PageHandler struct{}

// NewPageHandler creates a new page handler
func NewPageHandler() *PageHandler {
	return &PageHandler{}
}

// Index handles the index page
func (h *PageHandler) Index(c *fiber.Ctx) error {
	return c.Render("base", fiber.Map{
		"title": "7x42 - Home",
	})
}

// Chat handles the chat page
func (h *PageHandler) Chat(c *fiber.Ctx) error {
	return c.Render("chat", fiber.Map{
		"title": "7x42 - Chat",
	})
}

// Settings handles the settings page
func (h *PageHandler) Settings(c *fiber.Ctx) error {
	return c.SendString("Settings page")
}
