-- +goose Up
CREATE VIEW plant_search_view AS
SELECT
    p.id,
    p.common,
    p.scientific,
    p.inatrualist_taxon_id,
    p.section,
    p.color,
    p.bloom,
    p.height,
    p.height_sort,
    p.sun,
    p.water,
    p.price,
    p.available,
    i.image_url,
    setweight(p.fts_common, 'A') ||
    setweight(p.fts_scientific, 'B') ||
    setweight(i.fts_summary, 'C') AS search_vector
FROM plants p
LEFT JOIN inatrualist i ON i.plant_id = p.id;

-- +goose Down
DROP VIEW plant_search_view;
