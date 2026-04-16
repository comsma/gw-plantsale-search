package indexer

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"sync/atomic"

	"github.com/comsma/gw-plantsale-search/internal/inatrualist"
	"github.com/comsma/gw-plantsale-search/internal/models"
)

// Syncer fetches iNaturalist data for all plants in the background.
// Trigger is safe to call concurrently; duplicate calls while a sync is
// already running are silently dropped.
type Syncer struct {
	queries *models.Queries
	running atomic.Bool
}

func New(queries *models.Queries) *Syncer {
	return &Syncer{queries: queries}
}

// Trigger starts a background sync. If a sync is already running it returns
// immediately without starting a second one.
func (s *Syncer) Trigger() {
	if !s.running.CompareAndSwap(false, true) {
		log.Println("indexer: sync already in progress, skipping")
		return
	}
	go func() {
		defer s.running.Store(false)
		if err := s.sync(); err != nil {
			log.Printf("indexer: sync error: %v", err)
		}
	}()
}

func (s *Syncer) sync() error {
	ctx := context.Background()
	log.Print("Inatrualist Syncer started")

	plants, err := s.queries.GetAllPlants(ctx)
	if err != nil {
		return fmt.Errorf("get plants: %w", err)
	}

	var ok, failed int
	for _, p := range plants {
		if !p.InatrualistTaxonID.Valid || p.InatrualistTaxonID.String == "" {
			continue
		}
		taxon, err := strconv.Atoi(p.InatrualistTaxonID.String)
		if err != nil {
			log.Printf("indexer: skipping %q: invalid taxon ID %q", p.Common, p.InatrualistTaxonID.String)
			continue
		}

		details, err := inatrualist.GetPlantDetails(taxon)
		if err != nil {
			log.Printf("indexer: inat failed for %q (taxon %d): %v", p.Common, taxon, err)
			failed++
			continue
		}

		log.Printf("indexer: syncing %q (taxon %d)", p.Common, taxon)

		if err := s.queries.UpsertInatrualistData(ctx, models.UpsertInatrualistDataParams{
			PlantID:     p.ID,
			Summary:     details.Summary,
			ImageUrl:    details.ImageUrl,
			Attribution: details.ImageAttribution,
		}); err != nil {
			log.Printf("indexer: upsert failed for %q: %v", p.Common, err)
			failed++
			continue
		}
		ok++
	}

	log.Printf("indexer: sync complete — %d ok, %d failed", ok, failed)
	return nil
}
