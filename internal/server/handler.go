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

	"github.com/comsma/gw-plantsale-search/internal/models"
	"github.com/labstack/echo/v5"
)

type Handler struct {
	queries *models.Queries
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

func toPlantView(p models.Plant) PlantView {
	return PlantView{
		Taxon:      p.ID,
		Common:     p.Common,
		Scientific: p.Scientific.String,
		Section:    p.Section.String,
		Color:      p.Color.String,
		Bloom:      p.Bloom.String,
		Height:     p.Height.String,
		Sun:        p.Sun.String,
		Soil:       p.Water.String,
		Price:      formatPrice(p.Price),
		InatURL:    inatURL(p.InatrualistTaxonID),
	}
}

func toPlantViewFromSearch(r models.SearchPlantsRow) PlantView {
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
		ImageURL:   r.ImageUrl.String,
		Available:  r.Available,
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
	if taxonID.Valid && taxonID.String != "" {
		return "https://www.inaturalist.org/taxa/" + taxonID.String
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

func ns(s string) sql.NullString {
	return sql.NullString{String: s, Valid: true}
}

// fuzzyQuery converts a user search string into a BOOLEAN MODE expression
// with prefix wildcards so partial words still match.
// "purple cone" → "purple* cone*"
func fuzzyQuery(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	words := strings.Fields(s)
	for i, w := range words {
		words[i] = w + "*"
	}
	return strings.Join(words, " ")
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

	fq := fuzzyQuery(query)
	rows, err := h.queries.SearchPlants(ctx, models.SearchPlantsParams{
		Query:      fq,
		Query_2:    fq,
		Query_3:    fq,
		Section:    ns(section),
		Color:      ns(color),
		Sun:        ns(sun),
		Water:      ns(water),
		Query_5:    fq,
		Query_6:    fq,
		SortColumn: sort,
	})
	if err != nil {
		log.Printf("SearchPlants error: %v", err)
		rows = nil
	}

	all := make([]PlantView, len(rows))
	for i, p := range rows {
		all[i] = toPlantViewFromSearch(p)
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
