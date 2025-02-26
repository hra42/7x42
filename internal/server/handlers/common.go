package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/hra42/7x42/internal/repository"
	"github.com/hra42/7x42/internal/server/responses"
)

// ErrorHandler is the custom error handler for the server
func ErrorHandler(c *fiber.Ctx, err error) error {
	// Default status code
	code := fiber.StatusInternalServerError

	// Check if it's a Fiber error
	var e *fiber.Error
	if errors.As(err, &e) {
		code = e.Code
	} else if repository.IsNotFound(err) {
		code = fiber.StatusNotFound
	} else if repository.IsAlreadyExists(err) {
		code = fiber.StatusConflict
	}

	// Return error response
	return responses.Error(c, code, err.Error())
}

// ParseUint64Param parses a uint64 parameter from the request
func ParseUint64Param(c *fiber.Ctx, param string) (uint64, error) {
	id, err := strconv.ParseUint(c.Params(param), 10, 64)
	if err != nil {
		return 0, fiber.NewError(http.StatusBadRequest, "Invalid "+param+" parameter")
	}
	return id, nil
}

// ParsePagination parses pagination parameters from the request
func ParsePagination(c *fiber.Ctx) (page, pageSize int) {
	page, _ = strconv.Atoi(c.Query("page", "1"))
	pageSize, _ = strconv.Atoi(c.Query("pageSize", "20"))

	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	return page, pageSize
}

// GetUserID gets the user ID from the request
func GetUserID(c *fiber.Ctx) string {
	// In a real application, this would come from authentication
	return c.Query("userId", "default-user")
}
