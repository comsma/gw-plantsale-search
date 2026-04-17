package server

import (
	"fmt"
	"io/fs"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/comsma/gw-plantsale-search/internal/indexer"
	"github.com/comsma/gw-plantsale-search/internal/models"
	"github.com/comsma/gw-plantsale-search/ui"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func Start(db *pgx.Conn, syncer *indexer.Syncer) error {
	e := echo.New()

	sessionManager := scs.New()
	sessionManager.Lifetime = 24 * time.Hour

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
	e.Use(LoadAndSave(sessionManager))

	e.StaticFS("/static", staticFS)

	h := &Handler{queries: models.New(db), syncer: syncer}
	e.GET("/", h.Home)
	e.GET("/plants", h.PlantList)
	e.GET("/plants/:taxon", h.PlantDetail)
	e.POST("/admin/inat/resync", h.TriggerInatResync)

	syncer.Trigger()

	return e.Start(":8080")
}
