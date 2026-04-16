package search

import (
	"fmt"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/analysis/lang/en"
	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/blevesearch/bleve/v2/search/query"
)

// PlantDoc is the unit stored and searched in the bleve index.
// Only available plants are indexed.
type PlantDoc struct {
	ID         string `json:"id"`
	Common     string `json:"common"`
	Scientific string `json:"scientific"`
	Section    string `json:"section"`
	Color      string `json:"color"`
	Bloom      string `json:"bloom"`
	Height     string `json:"height"`
	HeightSort string `json:"height_sort"`
	Sun        string `json:"sun"`
	Water      string `json:"water"`
	Price      string `json:"price"`
	Summary    string `json:"summary"`
	ImageURL   string `json:"image_url"`
}

// SearchParams holds query text and filter values for a search request.
type SearchParams struct {
	Query   string
	Section string
	Color   string
	Sun     string
	Water   string
	Sort    string
}

// Index wraps an in-memory bleve index.
type Index struct {
	idx bleve.Index
}

// New creates a new in-memory bleve index.
func New() (*Index, error) {
	idx, err := bleve.NewMemOnly(buildMapping())
	if err != nil {
		return nil, fmt.Errorf("bleve: %w", err)
	}
	return &Index{idx: idx}, nil
}

// IndexPlant adds or updates a single plant document.
func (i *Index) IndexPlant(doc PlantDoc) error {
	return i.idx.Index(doc.ID, doc)
}

// DeletePlant removes a plant from the index.
func (i *Index) DeletePlant(id string) error {
	return i.idx.Delete(id)
}

// IndexBatch replaces the entire index contents with docs.
func (i *Index) IndexBatch(docs []PlantDoc) error {
	b := i.idx.NewBatch()
	for _, doc := range docs {
		if err := b.Index(doc.ID, doc); err != nil {
			return fmt.Errorf("batch index %q: %w", doc.ID, err)
		}
	}
	return i.idx.Batch(b)
}

// Search returns documents matching the given params.
func (i *Index) Search(params SearchParams) ([]PlantDoc, error) {
	req := bleve.NewSearchRequest(buildQuery(params))
	req.Size = 10000
	req.Fields = []string{"*"}

	switch params.Sort {
	case "height":
		req.SortBy([]string{"height_sort", "common"})
	case "price":
		req.SortBy([]string{"price", "common"})
	default:
		if params.Query != "" {
			req.SortBy([]string{"-_score", "common"})
		} else {
			req.SortBy([]string{"common"})
		}
	}

	res, err := i.idx.Search(req)
	if err != nil {
		return nil, fmt.Errorf("bleve search: %w", err)
	}

	out := make([]PlantDoc, len(res.Hits))
	for j, hit := range res.Hits {
		out[j] = PlantDoc{
			ID:         hit.ID,
			Common:     fieldStr(hit.Fields, "common"),
			Scientific: fieldStr(hit.Fields, "scientific"),
			Section:    fieldStr(hit.Fields, "section"),
			Color:      fieldStr(hit.Fields, "color"),
			Bloom:      fieldStr(hit.Fields, "bloom"),
			Height:     fieldStr(hit.Fields, "height"),
			HeightSort: fieldStr(hit.Fields, "height_sort"),
			Sun:        fieldStr(hit.Fields, "sun"),
			Water:      fieldStr(hit.Fields, "water"),
			Price:      fieldStr(hit.Fields, "price"),
			Summary:    fieldStr(hit.Fields, "summary"),
			ImageURL:   fieldStr(hit.Fields, "image_url"),
		}
	}
	return out, nil
}

func fieldStr(fields map[string]interface{}, key string) string {
	if v, ok := fields[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func buildQuery(params SearchParams) query.Query {
	var must []query.Query

	if params.Query != "" {
		cq := bleve.NewMatchQuery(params.Query)
		cq.SetFuzziness(1)
		cq.SetField("common")
		sq := bleve.NewMatchQuery(params.Query)
		sq.SetField("scientific")
		sumq := bleve.NewMatchPhraseQuery(params.Query)
		sumq.SetField("summary")
		cq.SetFuzziness(1)
		must = append(must, bleve.NewDisjunctionQuery(cq, sq, sumq))
	}

	addTermFilter := func(val, field string) {
		if val != "" {
			tq := bleve.NewTermQuery(val)
			tq.SetField(field)
			must = append(must, tq)
		}
	}
	addTermFilter(params.Section, "section")
	addTermFilter(params.Color, "color")
	addTermFilter(params.Sun, "sun")
	addTermFilter(params.Water, "water")

	if len(must) == 0 {
		return bleve.NewMatchAllQuery()
	}
	if len(must) == 1 {
		return must[0]
	}
	return bleve.NewConjunctionQuery(must...)
}

func buildMapping() mapping.IndexMapping {
	im := bleve.NewIndexMapping()

	textField := bleve.NewTextFieldMapping()
	textField.Analyzer = en.AnalyzerName

	keywordField := bleve.NewTextFieldMapping()
	keywordField.Analyzer = "keyword"

	dm := bleve.NewDocumentMapping()
	dm.AddFieldMappingsAt("common", textField)
	dm.AddFieldMappingsAt("scientific", textField)
	dm.AddFieldMappingsAt("summary", textField)
	dm.AddFieldMappingsAt("section", keywordField)
	dm.AddFieldMappingsAt("color", keywordField)
	dm.AddFieldMappingsAt("sun", keywordField)
	dm.AddFieldMappingsAt("water", keywordField)
	dm.AddFieldMappingsAt("height_sort", keywordField)
	dm.AddFieldMappingsAt("price", keywordField)

	im.DefaultMapping = dm
	return im
}
