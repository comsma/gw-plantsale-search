package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"strconv"
	"strings"

	"github.com/comsma/gw-plantsale-search/internal/config"
	"github.com/comsma/gw-plantsale-search/internal/indexer"
	"github.com/comsma/gw-plantsale-search/internal/migrations"
	"github.com/comsma/gw-plantsale-search/internal/models"
	"github.com/comsma/gw-plantsale-search/internal/plants"
	"github.com/comsma/gw-plantsale-search/internal/server"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/urfave/cli/v2"
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
					{Name: "create", Usage: "create new migration file", Action: runMigrateCreate, Flags: []cli.Flag{
						&cli.StringFlag{
							Name:    "name",
							Aliases: []string{"n"},
							Usage:   "name of the migration",
						},
					}},
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
				},
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func openDB() (*pgxpool.Pool, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL not set")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, err
	}
	return pool, nil
}

func openSQLDB() (*sql.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL not set")
	}
	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("migrate: parse config: %w", err)
	}

	return stdlib.OpenDB(*config), nil
}

// ── serve ────────────────────────────────────────────────────────────────────

func runServe(_ *cli.Context) error {
	cfg := config.Load()

	db, err := openDB()
	if err != nil {
		return err
	}
	defer db.Close()

	syncer := indexer.New(models.New(db))
	return server.Start(db, syncer, cfg)
}

// ── migrate ──────────────────────────────────────────────────────────────────

func runMigrateUp(_ *cli.Context) error {
	db, err := openSQLDB()
	if err != nil {
		return err
	}
	defer db.Close()
	goose.SetBaseFS(migrations.FS)
	goose.SetDialect("postgres")
	return goose.Up(db, ".")
}

func runMigrateDown(_ *cli.Context) error {
	db, err := openSQLDB()
	if err != nil {
		return err
	}
	defer db.Close()
	goose.SetBaseFS(migrations.FS)
	goose.SetDialect("postgres")
	return goose.Down(db, ".")
}

func runMigrateStatus(_ *cli.Context) error {
	db, err := openSQLDB()
	if err != nil {
		return err
	}
	defer db.Close()
	goose.SetBaseFS(migrations.FS)
	goose.SetDialect("postgres")
	return goose.Status(db, ".")
}

func runMigrateCreate(cliCtx *cli.Context) error {
	db, err := openSQLDB()
	if err != nil {
		return err
	}
	defer db.Close()
	goose.SetBaseFS(migrations.FS)
	goose.SetDialect("postgres")
	return goose.Create(db, "./internal/migrations", cliCtx.String("name"), "sql")
}

// ── ingest ───────────────────────────────────────────────────────────────────

func ingestPlants(cliCtx *cli.Context) error {
	db, err := openDB()
	if err != nil {
		return err
	}
	defer db.Close()
	ctx := context.Background()

	q := models.New(db)
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
			Scientific:         pgtype.Text{String: plant.Scientific, Valid: plant.Scientific != ""},
			InatrualistTaxonID: pgtype.Text{String: id, Valid: true},
			Section:            pgtype.Text{String: plant.Section, Valid: plant.Section != ""},
			Color:              pgtype.Text{String: plant.Color, Valid: plant.Color != ""},
			Bloom:              pgtype.Text{String: plant.Bloom, Valid: plant.Bloom != ""},
			Height:             pgtype.Text{String: plant.Height, Valid: plant.Height != ""},
			HeightSort:         pgtype.Text{String: fmt.Sprintf("%g", plant.HeightSort), Valid: true},
			Sun:                pgtype.Text{String: plant.Sun, Valid: plant.Sun != ""},
			Water:              pgtype.Text{String: plant.Soil, Valid: plant.Soil != ""},
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
	}

	fmt.Printf("done: %d created, %d skipped, %d inat ok, %d inat failed\n", created, skipped, inatOk, inatFail)
	return nil
}

func parsePrice(s string) pgtype.Numeric {
	s = strings.TrimPrefix(s, "$")
	if i := strings.Index(s, "/"); i >= 0 {
		s = s[:i]
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return pgtype.Numeric{Valid: false}
	}

	var intPart, fracPart string
	var exp int32
	if dot := strings.Index(s, "."); dot >= 0 {
		intPart = s[:dot]
		fracPart = s[dot+1:]
		exp = -int32(len(fracPart))
	} else {
		intPart = s
	}

	n := new(big.Int)
	if _, ok := n.SetString(intPart+fracPart, 10); !ok {
		return pgtype.Numeric{Valid: false}
	}
	return pgtype.Numeric{Int: n, Exp: exp, Valid: true}
}

func isDuplicateKey(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
