package main

import (
	"log"

	"github.com/comsma/gw-plantsale-search/ui"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func main() {
	e := echo.New()
	tmpl, err := NewTemplateCache()

	if err != nil {
		panic(err)
	}

	e.Renderer = tmpl
	e.Use(middleware.Gzip())
	e.Use(middleware.RequestLogger())
	e.StaticFS("/static", ui.Static)

	h := &Handler{}
	e.GET("/", h.Home)
	e.GET("/plants", h.PlantList)
	e.GET("/plants/:taxon", h.PlantDetail)

	if err := e.Start(":8080"); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
