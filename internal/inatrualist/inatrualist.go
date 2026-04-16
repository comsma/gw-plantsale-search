package inatrualist

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/http2"
	"golang.org/x/time/rate"
)

type Plant struct {
	ScientificName   string
	CommonName       string
	Summary          string
	ImageUrl         string
	ImageAttribution string
}

var htmlTagRe = regexp.MustCompile(`<[^>]+>`)

// limiter enforces the iNaturalist API limit of 60 requests/minute.
var limiter = rate.NewLimiter(rate.Every(2*time.Second), 1)

var client = func() *http.Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12},
		DialContext: (&net.Dialer{
			Timeout:   15 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
	}
	err := http2.ConfigureTransport(transport)
	if err != nil {
		log.Printf("http2.ConfigureTransport: %v", err)
		return nil
	}
	return &http.Client{Timeout: 15 * time.Second, Transport: transport}
}()

func GetPlantDetails(taxon int) (*Plant, error) {
	if err := limiter.Wait(context.Background()); err != nil {
		return nil, fmt.Errorf("rate limiter: %w", err)
	}

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
