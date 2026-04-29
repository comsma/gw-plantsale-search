-- +goose Up
ALTER TABLE plants ADD COLUMN bloom_sort INTEGER;

UPDATE plants SET bloom_sort =
    CASE SPLIT_PART(COALESCE(bloom, ''), '-', 1)
        WHEN 'Jan' THEN 1
        WHEN 'Feb' THEN 2
        WHEN 'Mar' THEN 3
        WHEN 'Apr' THEN 4
        WHEN 'May' THEN 5
        WHEN 'Jun' THEN 6
        WHEN 'Jul' THEN 7
        WHEN 'Aug' THEN 8
        WHEN 'Sep' THEN 9
        WHEN 'Oct' THEN 10
        WHEN 'Nov' THEN 11
        WHEN 'Dec' THEN 12
        ELSE NULL
    END;

DROP VIEW plant_search_view;

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
    p.bloom_sort,
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

ALTER TABLE plants DROP COLUMN bloom_sort;

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
