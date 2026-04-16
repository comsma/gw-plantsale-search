package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upPlants, downPlants)
}

func upPlants(tx *sql.Tx) error {
	_, err := tx.Exec(`CREATE TABLE plants (
    id                   VARCHAR(20)    NOT NULL PRIMARY KEY,
    common               VARCHAR(255)   NOT NULL,
    scientific           VARCHAR(255),
    inatrualist_taxon_id VARCHAR(255),
    section              VARCHAR(255),
    color                VARCHAR(255),
    bloom                VARCHAR(255),
    height               VARCHAR(255),
    height_sort          VARCHAR(255),
    sun                  VARCHAR(255),
    water                VARCHAR(255),
    price                DECIMAL(10,2)  NOT NULL DEFAULT 0,
    available            BOOLEAN        NOT NULL DEFAULT true,
    FULLTEXT KEY ft_plants_search (common, scientific)
)`)
	return err
}

func downPlants(tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE plants`)
	return err
}
