-- +goose Up
CREATE TABLE inatrualist (
    plant_id VARCHAR(20) NOT NULL PRIMARY KEY,
    summary VARCHAR(1000) NOT NULL,
    image_url VARCHAR(255) NOT NULL,
    attribution VARCHAR(255) NOT NULL,
    last_updated DATETIME NOT NULL,
    FOREIGN KEY (plant_id) REFERENCES plants(id),
    FULLTEXT KEY ft_inatrualist_summary (summary)
);

-- +goose Down
DROP TABLE inatrualist;
