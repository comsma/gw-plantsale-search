-- +goose Up
CREATE TABLE plants (
    id                   TEXT           NOT NULL PRIMARY KEY,
    common               TEXT           NOT NULL,
    scientific           TEXT,
    inatrualist_taxon_id TEXT,
    section              TEXT,
    color                TEXT,
    bloom                TEXT,
    height               TEXT,
    height_sort          TEXT,
    sun                  TEXT,
    water                TEXT,
    price                NUMERIC(10,2)  NOT NULL DEFAULT 0,
    available            BOOLEAN        NOT NULL DEFAULT true,
    fts_common tsvector generated always as (to_tsvector('simple', common)) stored,
    fts_scientific tsvector generated always as (to_tsvector('simple', scientific)) stored
);

CREATE INDEX idx_plants_fts_common ON plants USING gin (fts_common);
CREATE INDEX idx_plants_fts_scientific ON plants USING gin (fts_scientific);

-- +goose Down
DROP INDEX idx_plants_fts_common;
DROP INDEX idx_plants_fts_scientific;
DROP TABLE plants;
