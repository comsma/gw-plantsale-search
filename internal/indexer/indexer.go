package indexer

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"sync/atomic"

	"github.com/comsma/gw-plantsale-search/internal/inatrualist"
	"github.com/comsma/gw-plantsale-search/internal/models"
	"github.com/comsma/gw-plantsale-search/internal/search"
)

// Syncer fetches iNaturalist data for all plants and keeps the search index
// up to date. Trigger is safe to call concurrently; duplicate calls while a
// sync is already running are silently dropped.
type Syncer struct {
	queries *models.Queries
	search  *search.Index
	running atomic.Bool
}

func New(queries *models.Queries, idx *search.Index) *Syncer {
	return &Syncer{queries: queries, search: idx}
}

// Build performs an initial synchronous population of the search index from
// the database. Call this once at startup before serving requests.
func (s *Syncer) Build(ctx context.Context) error {
	rows, err := s.queries.GetAllPlantsWithInatrualist(ctx)
	if err != nil {
		return fmt.Errorf("get plants: %w", err)
	}

	docs := make([]search.PlantDoc, 0, len(rows))
	for _, r := range rows {
		if !r.Available {
			continue
		}
		docs = append(docs, search.PlantDoc{
			ID:         r.ID,
			Common:     r.Common,
			Scientific: r.Scientific.String,
			Section:    r.Section.String,
			Color:      r.Color.String,
			Bloom:      r.Bloom.String,
			Height:     r.Height.String,
			HeightSort: r.HeightSort.String,
			Sun:        r.Sun.String,
			Water:      r.Water.String,
			Price:      r.Price,
			Summary:    r.Summary.String,
			ImageURL:   r.ImageUrl.String,
		})
	}

	if err := s.search.IndexBatch(docs); err != nil {
		return fmt.Errorf("index batch: %w", err)
	}
	log.Printf("indexer: initial index built — %d plants", len(docs))
	return nil
}

// Trigger starts a background iNaturalist sync. If one is already running it
// returns immediately without starting a second one.
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
	log.Print("indexer: iNaturalist sync started")

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

		// Re-index the plant with updated iNaturalist data.
		row, err := s.queries.GetPlantWithInatrualist(ctx, p.ID)
		if err != nil {
			log.Printf("indexer: re-fetch failed for %q: %v", p.Common, err)
		} else if row.Available {
			if err := s.search.IndexPlant(rowToDoc(row)); err != nil {
				log.Printf("indexer: re-index failed for %q: %v", p.Common, err)
			}
		}

		ok++
	}

	log.Printf("indexer: sync complete — %d ok, %d failed", ok, failed)
	return nil
}

func rowToDoc(r models.GetPlantWithInatrualistRow) search.PlantDoc {
	return search.PlantDoc{
		ID:         r.ID,
		Common:     r.Common,
		Scientific: r.Scientific.String,
		Section:    r.Section.String,
		Color:      r.Color.String,
		Bloom:      r.Bloom.String,
		Height:     r.Height.String,
		HeightSort: r.HeightSort.String,
		Sun:        r.Sun.String,
		Water:      r.Water.String,
		Price:      r.Price,
		Summary:    r.Summary.String,
		ImageURL:   r.ImageUrl.String,
	}
}
