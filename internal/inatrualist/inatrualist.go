package inatrualist

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type Plant struct {
	ScientificName   string
	CommonName       string
	Summary          string
	ImageUrl         string
	ImageAttribution string
}

var htmlTagRe = regexp.MustCompile(`<[^>]+>`)

var client = &http.Client{Timeout: 8 * time.Second}

func GetPlantDetails(taxon int) (*Plant, error) {
	url := fmt.Sprintf("https://api.inaturalist.org/v1/taxa/%d", taxon)

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("inat fetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("inat API status %d", resp.StatusCode)
	}

	var payload struct {
		Results []struct {
			Name          string `json:"name"`
			PreferredName string `json:"preferred_common_name"`
			Summary       string `json:"wikipedia_summary"`
			DefaultPhoto  struct {
				MediumURL   string `json:"medium_url"`
				Attribution string `json:"attribution"`
			} `json:"default_photo"`
		} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("inat decode: %w", err)
	}

	if len(payload.Results) == 0 {
		return nil, fmt.Errorf("inat: no results for taxon %d", taxon)
	}

	r := payload.Results[0]
	summary := strings.TrimSpace(htmlTagRe.ReplaceAllString(r.Summary, ""))

	return &Plant{
		ScientificName:   r.Name,
		CommonName:       r.PreferredName,
		Summary:          summary,
		ImageUrl:         r.DefaultPhoto.MediumURL,
		ImageAttribution: r.DefaultPhoto.Attribution,
	}, nil
}
