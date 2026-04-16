package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/pressly/goose/v3"
	"github.com/urfave/cli/v2"

	"github.com/comsma/gw-plantsale-search/internal/inatrualist"
	_ "github.com/comsma/gw-plantsale-search/internal/migrations"
	"github.com/comsma/gw-plantsale-search/internal/models"
	"github.com/comsma/gw-plantsale-search/internal/plants"
	"github.com/comsma/gw-plantsale-search/internal/server"
)

func main() {
	app := &cli.App{
		Name:  "plantsale",
		Usage: "plant sale search application",
		Commands: []*cli.Command{
			{
				Name:   "serve",
				Usage:  "start the web server",
				Action: runServe,
			},
			{
				Name:  "migrate",
				Usage: "database migration tools",
				Subcommands: []*cli.Command{
					{Name: "up", Usage: "apply all pending migrations", Action: runMigrateUp},
					{Name: "down", Usage: "roll back the last migration", Action: runMigrateDown},
					{Name: "status", Usage: "print migration status", Action: runMigrateStatus},
				},
			},
			{
				Name:  "ingest",
				Usage: "data ingestion tools",
				Subcommands: []*cli.Command{
					{Name: "plants", Usage: "ingest plants from JSON and fetch iNaturalist data", Action: ingestPlants, Flags: []cli.Flag{
						&cli.StringFlag{
							Name:    "plant-list",
							Aliases: []string{"pl"},
							Value:   "plants.json",
							Usage:   "Location of plant list",
						},
					}},
					{Name: "refresh-inat", Usage: "re-fetch iNaturalist data for all plants in the database", Action: refreshInat},
				},
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func openDB() (*sql.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL not set")
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}
	return db, nil
}

// ── serve ────────────────────────────────────────────────────────────────────

func runServe(_ *cli.Context) error {
	db, err := openDB()
	if err != nil {
		return err
	}
	defer db.Close()
	return server.Start(db)
}

// ── migrate ──────────────────────────────────────────────────────────────────

func runMigrateUp(_ *cli.Context) error {
	db, err := openDB()
	if err != nil {
		return err
	}
	defer db.Close()
	goose.SetDialect("mysql")
	return goose.Up(db, ".")
}

func runMigrateDown(_ *cli.Context) error {
	db, err := openDB()
	if err != nil {
		return err
	}
	defer db.Close()
	goose.SetDialect("mysql")
	return goose.Down(db, ".")
}

func runMigrateStatus(_ *cli.Context) error {
	db, err := openDB()
	if err != nil {
		return err
	}
	defer db.Close()
	goose.SetDialect("mysql")
	return goose.Status(db, ".")
}

// ── ingest ───────────────────────────────────────────────────────────────────

func ingestPlants(cliCtx *cli.Context) error {

	db, err := openDB()
	if err != nil {
		return err
	}
	defer db.Close()

	q := models.New(db)
	ctx := context.Background()
	var created, skipped, inatOk, inatFail int

	plantList, err := plants.LoadPlants(cliCtx.String("plant-list"))
	if err != nil {
		return fmt.Errorf("failed to load plants: %w", err)
	}

	for _, plant := range plantList {
		id := strconv.Itoa(plant.Taxon)

		err := q.CreatePlant(ctx, models.CreatePlantParams{
			ID:                 id,
			Common:             plant.Common,
			Scientific:         sql.NullString{String: plant.Scientific, Valid: plant.Scientific != ""},
			InatrualistTaxonID: sql.NullString{String: id, Valid: true},
			Section:            sql.NullString{String: plant.Section, Valid: plant.Section != ""},
			Color:              sql.NullString{String: plant.Color, Valid: plant.Color != ""},
			Bloom:              sql.NullString{String: plant.Bloom, Valid: plant.Bloom != ""},
			Height:             sql.NullString{String: plant.Height, Valid: plant.Height != ""},
			HeightSort:         sql.NullString{String: fmt.Sprintf("%g", plant.HeightSort), Valid: true},
			Sun:                sql.NullString{String: plant.Sun, Valid: plant.Sun != ""},
			Water:              sql.NullString{String: plant.Soil, Valid: plant.Soil != ""},
			Price:              parsePrice(plant.Price),
			Available:          true,
		})
		if err != nil {
			if isDuplicateKey(err) {
				skipped++
				continue
			}
			return fmt.Errorf("create plant %q: %w", plant.Common, err)
		}
		created++

		if ok := fetchAndUpsertInat(ctx, q, id, plant.Taxon, plant.Common); ok {
			inatOk++
		} else {
			inatFail++
		}
	}

	fmt.Printf("done: %d created, %d skipped, %d inat ok, %d inat failed\n", created, skipped, inatOk, inatFail)
	return nil
}

func refreshInat(_ *cli.Context) error {
	db, err := openDB()
	if err != nil {
		return err
	}
	defer db.Close()

	q := models.New(db)
	ctx := context.Background()

	allPlants, err := q.GetAllPlants(ctx)
	if err != nil {
		return fmt.Errorf("get plants: %w", err)
	}

	var ok, failed int
	for _, p := range allPlants {
		if !p.InatrualistTaxonID.Valid || p.InatrualistTaxonID.String == "" {
			log.Printf("skipping %q: no taxon ID", p.Common)
			continue
		}
		taxon, err := strconv.Atoi(p.InatrualistTaxonID.String)
		if err != nil {
			log.Printf("skipping %q: invalid taxon ID %q", p.Common, p.InatrualistTaxonID.String)
			continue
		}
		if fetchAndUpsertInat(ctx, q, p.ID, taxon, p.Common) {
			ok++
		} else {
			failed++
		}
	}

	fmt.Printf("done: %d inat ok, %d failed\n", ok, failed)
	return nil
}

func fetchAndUpsertInat(ctx context.Context, q *models.Queries, id string, taxon int, common string) bool {
	details, err := inatrualist.GetPlantDetails(taxon)
	if err != nil {
		log.Printf("inat failed for %q (taxon %d): %v", common, taxon, err)
		return false
	}
	if err := q.UpsertInatrualistData(ctx, models.UpsertInatrualistDataParams{
		PlantID:     id,
		Summary:     details.Summary,
		ImageUrl:    details.ImageUrl,
		Attribution: details.ImageAttribution,
	}); err != nil {
		log.Printf("upsert inat failed for %q: %v", common, err)
		return false
	}
	return true
}

func parsePrice(s string) string {
	s = strings.TrimPrefix(s, "$")
	if i := strings.Index(s, "/"); i >= 0 {
		s = s[:i]
	}
	return strings.TrimSpace(s)
}

func isDuplicateKey(err error) bool {
	return strings.Contains(err.Error(), "Duplicate entry") || strings.Contains(err.Error(), "1062")
}
