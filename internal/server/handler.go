package server

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/alexedwards/scs/v2"
	"github.com/comsma/gw-plantsale-search/internal/indexer"
	"github.com/comsma/gw-plantsale-search/internal/models"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v5"
)

type Handler struct {
	session *scs.SessionManager
	queries *models.Queries
	syncer  *indexer.Syncer
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

func toPlantViewFromSearchRow(r models.SearchPlantsRow) PlantView {
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
		ImageURL:   r.ImageUrl.String,
		Price:      formatPrice(r.Price),
		InatURL:    inatURL(r.InatrualistTaxonID.String),
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
		InatURL:    inatURL(r.InatrualistTaxonID.String),
		Available:  r.Available,
	}
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
		Sections: unwrapTextSlice(sections),
		Suns:     unwrapTextSlice(suns),
		Soils:    unwrapTextSlice(soils),
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

	plants, err := h.queries.SearchPlants(ctx, models.SearchPlantsParams{
		Query:      query,
		Section:    section,
		Color:      color,
		Sun:        sun,
		Water:      water,
		SortColumn: sort,
	})
	if err != nil {
		return err
	}

	all := make([]PlantView, len(plants))
	for i, p := range plants {
		all[i] = toPlantViewFromSearchRow(p)
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
		Sections:    unwrapTextSlice(sections),
		Suns:        unwrapTextSlice(suns),
		Soils:       unwrapTextSlice(soils),
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

func unwrapTextSlice(ts []pgtype.Text) []string {
	out := make([]string, 0, len(ts))
	for _, t := range ts {
		if t.Valid {
			out = append(out, t.String)
		}
	}
	return out
}

func inatURL(taxonID string) string {
	if taxonID != "" {
		return "https://www.inaturalist.org/taxa/" + taxonID
	}
	return ""
}

func formatPrice(n pgtype.Numeric) string {
	if !n.Valid {
		return ""
	}

	floatVal, err := n.Float64Value()
	if err != nil {
		return ""
	}
	return fmt.Sprintf("$%.2f", floatVal.Float64)
}
