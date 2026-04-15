package main

import (
	"net/http"
	"strconv"

	"github.com/comsma/gw-plantsale-search/internal/inatrualist"
	"github.com/comsma/gw-plantsale-search/internal/plants"

	"github.com/labstack/echo/v5"
)

type Handler struct{}

type HomeData struct {
	Sections []string
	Suns     []string
	Soils    []string
}

type PlantListData struct {
	Sections []string
	Suns     []string
	Soils    []string
	Sun      string
	Soil     string
	Section  string
	Search   string
	Sort     string
	Plants   []plants.Plant
}

type PlantDetailData struct {
	Plant            plants.Plant
	Summary          string
	ImageURL         string
	ImageAttribution string
}

func (h *Handler) Home(c *echo.Context) error {
	data := HomeData{
		Sections: plants.UniqueSections(),
		Suns:     plants.UniqueSuns(),
		Soils:    plants.UniqueSoils(),
	}
	return c.Render(http.StatusOK, "pages/index.gohtml", data)
}

func (h *Handler) PlantList(c *echo.Context) error {
	params := plants.FilterParams{
		Sun:     c.QueryParam("sun"),
		Soil:    c.QueryParam("soil"),
		Section: c.QueryParam("section"),
		Search:  c.QueryParam("search"),
		Sort:    c.QueryParam("sort"),
	}

	data := PlantListData{
		Sections: plants.UniqueSections(),
		Suns:     plants.UniqueSuns(),
		Soils:    plants.UniqueSoils(),
		Sun:      params.Sun,
		Soil:     params.Soil,
		Section:  params.Section,
		Search:   params.Search,
		Sort:     params.Sort,
		Plants:   plants.Filtered(params),
	}

	if c.Request().Header.Get("HX-Request") == "true" {
		return c.Render(http.StatusOK, "partials/plant_list.gohtml", data)
	}
	return c.Render(http.StatusOK, "pages/results.gohtml", data)
}

func (h *Handler) PlantDetail(c *echo.Context) error {
	taxon, err := strconv.Atoi(c.Param("taxon"))
	if err != nil {
		return echo.ErrBadRequest
	}

	plant := plants.FindByTaxon(taxon)
	if plant == nil {
		return echo.ErrNotFound
	}

	data := PlantDetailData{Plant: *plant}

	if details, err := inatrualist.GetPlantDetails(taxon); err == nil {
		data.Summary = details.Summary
		data.ImageURL = details.ImageUrl
		data.ImageAttribution = details.ImageAttribution
	}

	return c.Render(http.StatusOK, "partials/plant_detail.gohtml", data)
}
