-- +goose Up
CREATE TABLE plants (
    id          VARCHAR(20)    NOT NULL PRIMARY KEY,
    common VARCHAR(255) NOT NULL,
    scientific VARCHAR(255),
    inatrualist_taxon_id VARCHAR(255),
    section VARCHAR(255),
    color VARCHAR(255),
    bloom VARCHAR(255),
    height VARCHAR(255),
    height_sort VARCHAR(255),
    sun VARCHAR(255),
    water VARCHAR(255),
    price DECIMAL(10,2) NOT NULL DEFAULT 0,
    available BOOLEAN NOT NULL DEFAULT true,
    FULLTEXT KEY ft_plants_search (common, scientific)
);

-- +goose Down
DROP TABLE plants;
