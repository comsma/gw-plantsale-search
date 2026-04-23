-- +goose Up
CREATE TABLE favorites_list (
    id                     serial         primary key,
    plant_id               TEXT           NOT NULL,
    user_id                TEXT           NOT NULL,
    FOREIGN KEY (plant_id) REFERENCES plants(id),
    UNIQUE (plant_id, user_id)

);
-- +goose Down
DROP TABLE favorites_list;
