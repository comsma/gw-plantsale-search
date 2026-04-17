-- +goose Up
CREATE TABLE inatrualist (
    plant_id     TEXT NOT NULL PRIMARY KEY,
    summary      TEXT NOT NULL,
    image_url    TEXT NOT NULL,
    attribution  TEXT NOT NULL,
    last_updated DATE NOT NULL,
    fts_summary tsvector generated always as (to_tsvector('english', summary)) stored,
    FOREIGN KEY (plant_id) REFERENCES plants(id)
);

CREATE INDEX idx_inatrualist_fts_summary ON inatrualist USING gin (fts_summary);

-- +goose Down
DROP INDEX idx_inatrualist_fts_summary;
DROP TABLE inatrualist;
