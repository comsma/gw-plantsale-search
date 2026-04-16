package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upInatrualist, downInatrualist)
}

func upInatrualist(tx *sql.Tx) error {
	_, err := tx.Exec(`CREATE TABLE inatrualist (
    plant_id     VARCHAR(20)   NOT NULL PRIMARY KEY,
    summary      VARCHAR(1000) NOT NULL,
    image_url    VARCHAR(255)  NOT NULL,
    attribution  VARCHAR(255)  NOT NULL,
    last_updated DATETIME      NOT NULL,
    FOREIGN KEY (plant_id) REFERENCES plants(id),
    FULLTEXT KEY ft_inatrualist_summary (summary)
)`)
	return err
}

func downInatrualist(tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE inatrualist`)
	return err
}
