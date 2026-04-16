package server

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/alexedwards/scs/v2"
	"github.com/comsma/gw-plantsale-search/internal/indexer"
	"github.com/comsma/gw-plantsale-search/internal/models"
	"github.com/comsma/gw-plantsale-search/internal/search"
	"github.com/labstack/echo/v5"
)

type Handler struct {
	session *scs.SessionManager
	queries *models.Queries
	syncer  *indexer.Syncer
	search  *search.Index
}

// PlantView maps models.Plant to template-compatible field names.
type PlantView struct {
	Taxon      string
	Common     string
	Scientific string
	Section    string
	Color      string
	Bloom      string
	Height     string
	Sun        string
	Soil       string
	Price      string
	Available  bool
	InatURL    string
	ImageURL   string
}

func toPlantViewFromDoc(d search.PlantDoc) PlantView {
	return PlantView{
		Taxon:      d.ID,
		Common:     d.Common,
		Scientific: d.Scientific,
		Section:    d.Section,
		Color:      d.Color,
		Bloom:      d.Bloom,
		Height:     d.Height,
		Sun:        d.Sun,
		Soil:       d.Water,
		Price:      formatPrice(d.Price),
		InatURL:    inatURLStr(d.ID),
		ImageURL:   d.ImageURL,
		Available:  true, // only available plants are indexed
	}
}

func toPlantViewFromRow(r models.GetPlantWithInatrualistRow) PlantView {
	return PlantView{
		Taxon:      r.ID,
		Common:     r.Common,
		Scientific: r.Scientific.String,
		Section:    r.Section.String,
		Color:      r.Color.String,
		Bloom:      r.Bloom.String,
		Height:     r.Height.String,
		Sun:        r.Sun.String,
		Soil:       r.Water.String,
		Price:      formatPrice(r.Price),
		InatURL:    inatURL(r.InatrualistTaxonID),
		Available:  r.Available,
	}
}

func inatURL(taxonID sql.NullString) string {
	return inatURLStr(taxonID.String)
}

func inatURLStr(taxonID string) string {
	if taxonID != "" {
		return "https://www.inaturalist.org/taxa/" + taxonID
	}
	return ""
}

func formatPrice(s string) string {
	if s == "" {
		return ""
	}
	return fmt.Sprintf("$%s", strings.TrimRight(strings.TrimRight(s, "0"), "."))
}

func unwrapNullStrings(items []sql.NullString) []string {
	out := make([]string, 0, len(items))
	for _, v := range items {
		if v.Valid {
			out = append(out, v.String)
		}
	}
	return out
}

type HomeData struct {
	Sections []string
	Suns     []string
	Soils    []string
}

const pageSize = 20

type PlantListData struct {
	Sections    []string
	Suns        []string
	Soils       []string
	Sun         string
	Soil        string
	Section     string
	Search      string
	Sort        string
	Plants      []PlantView
	HasMore     bool
	NextPageURL string
}

type PlantDetailData struct {
	Plant            PlantView
	Summary          string
	ImageURL         string
	ImageAttribution string
}

func (h *Handler) Home(c *echo.Context) error {
	ctx := context.Background()
	sections, _ := h.queries.GetDistinctSections(ctx)
	suns, _ := h.queries.GetDistinctSuns(ctx)
	soils, _ := h.queries.GetDistinctWaters(ctx)

	return c.Render(http.StatusOK, "pages/index.gohtml", HomeData{
		Sections: unwrapNullStrings(sections),
		Suns:     unwrapNullStrings(suns),
		Soils:    unwrapNullStrings(soils),
	})
}

func (h *Handler) PlantList(c *echo.Context) error {
	ctx := context.Background()
	query := c.QueryParam("search")
	section := c.QueryParam("section")
	color := c.QueryParam("color")
	sun := c.QueryParam("sun")
	water := c.QueryParam("soil")
	sort := c.QueryParam("sort")
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	if offset < 0 {
		offset = 0
	}

	docs, err := h.search.Search(search.SearchParams{
		Query:   query,
		Section: section,
		Color:   color,
		Sun:     sun,
		Water:   water,
		Sort:    sort,
	})
	if err != nil {
		log.Printf("search error: %v", err)
		docs = nil
	}

	all := make([]PlantView, len(docs))
	for i, d := range docs {
		all[i] = toPlantViewFromDoc(d)
	}

	end := offset + pageSize
	hasMore := end < len(all)
	if end > len(all) {
		end = len(all)
	}
	page := all[offset:end]

	nextPageURL := ""
	if hasMore {
		q := url.Values{}
		if query != "" {
			q.Set("search", query)
		}
		if section != "" {
			q.Set("section", section)
		}
		if sun != "" {
			q.Set("sun", sun)
		}
		if water != "" {
			q.Set("soil", water)
		}
		if sort != "" {
			q.Set("sort", sort)
		}
		q.Set("offset", strconv.Itoa(end))
		nextPageURL = "/plants?" + q.Encode()
	}

	sections, _ := h.queries.GetDistinctSections(ctx)
	suns, _ := h.queries.GetDistinctSuns(ctx)
	soils, _ := h.queries.GetDistinctWaters(ctx)

	data := PlantListData{
		Sections:    unwrapNullStrings(sections),
		Suns:        unwrapNullStrings(suns),
		Soils:       unwrapNullStrings(soils),
		Sun:         sun,
		Soil:        water,
		Section:     section,
		Search:      query,
		Sort:        sort,
		Plants:      page,
		HasMore:     hasMore,
		NextPageURL: nextPageURL,
	}

	isHX := c.Request().Header.Get("HX-Request") == "true"
	if isHX && offset > 0 {
		return c.Render(http.StatusOK, "partials/plant_page.gohtml", data)
	}
	if isHX {
		return c.Render(http.StatusOK, "partials/plant_list.gohtml", data)
	}
	return c.Render(http.StatusOK, "pages/results.gohtml", data)
}

func (h *Handler) TriggerInatResync(c *echo.Context) error {
	h.syncer.Trigger()
	return c.String(http.StatusAccepted, "resync started")
}

func (h *Handler) PlantDetail(c *echo.Context) error {
	ctx := context.Background()
	id := c.Param("taxon")

	row, err := h.queries.GetPlantWithInatrualist(ctx, id)
	if err != nil {
		return echo.ErrNotFound
	}

	return c.Render(http.StatusOK, "partials/plant_detail.gohtml", PlantDetailData{
		Plant:            toPlantViewFromRow(row),
		Summary:          row.Summary.String,
		ImageURL:         row.ImageUrl.String,
		ImageAttribution: row.Attribution.String,
	})
}
