package server

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/comsma/gw-plantsale-search/ui"
	"github.com/labstack/echo/v5"
)

type Renderer struct {
	templates map[string]*template.Template
}

func (r *Renderer) Render(c *echo.Context, w io.Writer, name string, data any) error {
	t, ok := r.templates[name]
	if !ok {
		return fmt.Errorf("template %q not found", name)
	}
	if strings.HasPrefix(name, "partials/") {
		stem := strings.TrimSuffix(filepath.Base(name), ".gohtml")
		return t.ExecuteTemplate(w, stem, data)
	}
	return t.Execute(w, data)
}

func NewTemplateCache() (*Renderer, error) {
	cache := map[string]*template.Template{}

	pages, err := fs.Glob(ui.Views, "views/pages/*.gohtml")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		fileName := filepath.Base(page)
		mapKey := "pages/" + fileName

		patterns := []string{
			"views/layouts/base.gohtml",
			"views/partials/*.gohtml",
			page,
		}

		ts, err := template.New(fileName).ParseFS(ui.Views, patterns...)
		if err != nil {
			return nil, err
		}

		cache[mapKey] = ts
	}

	partials, err := fs.Glob(ui.Views, "views/partials/*.gohtml")
	if err != nil {
		return nil, err
	}

	for _, partial := range partials {
		fileName := filepath.Base(partial)
		mapKey := "partials/" + fileName
		ts, err := template.New(fileName).ParseFS(ui.Views, "views/partials/*.gohtml")
		if err != nil {
			return nil, err
		}
		cache[mapKey] = ts
	}

	return &Renderer{templates: cache}, nil
}
