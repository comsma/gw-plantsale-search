package server

import (
	"database/sql"
	"fmt"
	"io/fs"

	"github.com/comsma/gw-plantsale-search/internal/models"
	"github.com/comsma/gw-plantsale-search/ui"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func Start(db *sql.DB) error {
	e := echo.New()

	tmpl, err := NewTemplateCache()
	if err != nil {
		return fmt.Errorf("template cache: %w", err)
	}

	staticFS, err := fs.Sub(ui.Static, "static")
	if err != nil {
		return fmt.Errorf("static fs: %w", err)
	}

	e.Renderer = tmpl
	e.Use(middleware.Gzip())
	e.Use(middleware.RequestLogger())
	e.StaticFS("/static", staticFS)

	h := &Handler{queries: models.New(db)}
	e.GET("/", h.Home)
	e.GET("/plants", h.PlantList)
	e.GET("/plants/:taxon", h.PlantDetail)

	return e.Start(":8080")
}
