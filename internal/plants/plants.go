package plants

import (
	_ "embed"
	"encoding/json"
	"sort"
	"strings"
)

//go:embed plant_sale_list.json
var rawJSON []byte

type Plant struct {
	Common     string  `json:"common"`
	Scientific string  `json:"scientific"`
	Taxon      int     `json:"taxon"`
	Section    string  `json:"section"`
	Color      string  `json:"color"`
	Bloom      string  `json:"bloom"`
	Height     string  `json:"height"`
	Sun        string  `json:"sun"`
	Soil       string  `json:"soil"`
	Price      string  `json:"price"`
	InatURL    string  `json:"inat"`
	HeightSort float64 `json:"heightSort"`
}

var All []Plant

func init() {
	if err := json.Unmarshal(rawJSON, &All); err != nil {
		panic("plants: failed to parse plant_sale_list.json: " + err.Error())
	}
}

type FilterParams struct {
	Sun     string
	Soil    string
	Section string
	Search  string
	Sort    string
}

func Filtered(p FilterParams) []Plant {
	var result []Plant
	search := strings.ToLower(strings.TrimSpace(p.Search))

	for _, plant := range All {
		if p.Sun != "" && !strings.Contains(strings.ToLower(plant.Sun), strings.ToLower(p.Sun)) {
			continue
		}
		if p.Soil != "" && !strings.Contains(strings.ToLower(plant.Soil), strings.ToLower(p.Soil)) {
			continue
		}
		if p.Section != "" && !strings.EqualFold(plant.Section, p.Section) {
			continue
		}
		if search != "" {
			if !strings.Contains(strings.ToLower(plant.Common), search) &&
				!strings.Contains(strings.ToLower(plant.Scientific), search) {
				continue
			}
		}
		result = append(result, plant)
	}

	sortPlants(result, p.Sort)
	return result
}

func sortPlants(plants []Plant, by string) {
	sort.SliceStable(plants, func(i, j int) bool {
		a, b := plants[i], plants[j]
		switch by {
		case "bloom":
			return a.Bloom < b.Bloom
		case "height":
			return a.HeightSort < b.HeightSort
		case "price":
			return a.Price < b.Price
		case "section":
			if a.Section != b.Section {
				return a.Section < b.Section
			}
			return strings.ToLower(a.Common) < strings.ToLower(b.Common)
		default:
			return strings.ToLower(a.Common) < strings.ToLower(b.Common)
		}
	})
}

func FindByTaxon(taxon int) *Plant {
	for i := range All {
		if All[i].Taxon == taxon {
			return &All[i]
		}
	}
	return nil
}

func UniqueSections() []string {
	return uniqueField(func(p Plant) string { return p.Section })
}

func UniqueSuns() []string {
	return uniqueField(func(p Plant) string { return p.Sun })
}

func UniqueSoils() []string {
	return uniqueField(func(p Plant) string { return p.Soil })
}

func uniqueField(fn func(Plant) string) []string {
	seen := map[string]struct{}{}
	var result []string
	for _, p := range All {
		v := fn(p)
		if v == "" {
			continue
		}
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			result = append(result, v)
		}
	}
	sort.Strings(result)
	return result
}
