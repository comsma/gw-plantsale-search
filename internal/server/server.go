package server

import (
	"database/sql"
	"fmt"
	"io/fs"

	"github.com/comsma/gw-plantsale-search/internal/indexer"
	"github.com/comsma/gw-plantsale-search/internal/models"
	"github.com/comsma/gw-plantsale-search/ui"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func Start(db *sql.DB, syncer *indexer.Syncer) error {
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
	e.StaticFS("/static", staticFS)

	h := &Handler{queries: models.New(db), syncer: syncer}
	e.GET("/", h.Home)
	e.GET("/plants", h.PlantList)
	e.GET("/plants/:taxon", h.PlantDetail)
	e.POST("/admin/inat/resync", h.TriggerInatResync)

	syncer.Trigger()

	return e.Start(":8080")
}
