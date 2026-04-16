package plants

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
)

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

// LoadPlants loads the plants from the given file path.
func LoadPlants(filePath string) ([]Plant, error) {
	var plants []Plant
	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(file, &plants); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Plant List: %w", err)
	}
	return plants, nil
}
