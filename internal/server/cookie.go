package server

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

// GetVisitorID Get or create visitor ID

func GetVisitorID(c *echo.Context) string {
	cookie, err := c.Cookie("visitor_id")
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}

	// Create new ID
	id := uuid.New().String()
	c.SetCookie(&http.Cookie{
		Name:     "visitor_id",
		Value:    id,
		Path:     "/",
		MaxAge:   365 * 24 * 60 * 60, // 1 year
		HttpOnly: true,
	})
	return id
}
