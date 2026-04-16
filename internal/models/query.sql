-- name: GetAllPlants :many
SELECT * FROM plants ORDER BY common;

-- name: GetAllPlantsWithInatrualist :many
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
    i.summary,
    i.image_url,
    i.attribution
FROM plants p
LEFT JOIN inatrualist i ON i.plant_id = p.id
ORDER BY p.common;

-- name: SearchPlants :many
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
    i.image_url
FROM plants p
LEFT JOIN inatrualist i ON i.plant_id = p.id
WHERE p.available = true
  AND (
    sqlc.arg(query) = ''
    OR MATCH(p.common, p.scientific) AGAINST (sqlc.arg(query) IN NATURAL LANGUAGE MODE)
    OR MATCH(i.summary) AGAINST (sqlc.arg(query) IN NATURAL LANGUAGE MODE)
  )
  AND (sqlc.arg(section) = '' OR p.section = sqlc.arg(section))
  AND (sqlc.arg(color)   = '' OR p.color   = sqlc.arg(color))
  AND (sqlc.arg(sun)     = '' OR p.sun     = sqlc.arg(sun))
  AND (sqlc.arg(water)   = '' OR p.water   = sqlc.arg(water))
ORDER BY
    CASE sqlc.arg(sort_column)
        WHEN 'common' THEN p.common
        WHEN 'height' THEN p.height_sort
        WHEN 'price' THEN p.price
        ELSE p.common
    END ASC,
    CASE WHEN sqlc.arg(query) = '' THEN 0 ELSE (
        (MATCH(p.common, p.scientific) AGAINST (sqlc.arg(query) IN NATURAL LANGUAGE MODE))+
        (MATCH(i.summary) AGAINST (sqlc.arg(query) IN NATURAL LANGUAGE MODE))
    ) END ASC,
    p.common ASC;

-- name: GetPlantWithInatrualist :one
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
    i.summary,
    i.image_url,
    i.attribution
FROM plants p
LEFT JOIN inatrualist i ON i.plant_id = p.id
WHERE p.id = sqlc.arg(id)
LIMIT 1;

-- name: CreatePlant :exec
INSERT INTO plants (
    id,
    common,
    scientific,
    inatrualist_taxon_id,
    section,
    color,
    bloom,
    height,
    height_sort,
    sun,
    water,
    price,
    available
) VALUES (
    sqlc.arg(id),
    sqlc.arg(common),
    sqlc.arg(scientific),
    sqlc.arg(inatrualist_taxon_id),
    sqlc.arg(section),
    sqlc.arg(color),
    sqlc.arg(bloom),
    sqlc.arg(height),
    sqlc.arg(height_sort),
    sqlc.arg(sun),
    sqlc.arg(water),
    sqlc.arg(price),
    sqlc.arg(available)
);

-- name: UpsertInatrualistData :exec
INSERT INTO inatrualist (
    plant_id,
    summary,
    image_url,
    attribution,
    last_updated
) VALUES (
    sqlc.arg(plant_id),
    sqlc.arg(summary),
    sqlc.arg(image_url),
    sqlc.arg(attribution),
    NOW()
) ON DUPLICATE KEY UPDATE
    summary      = VALUES(summary),
    image_url    = VALUES(image_url),
    attribution  = VALUES(attribution),
    last_updated = NOW();

-- name: GetDistinctSections :many
SELECT DISTINCT section FROM plants
WHERE section IS NOT NULL AND section != '' AND available = true
ORDER BY section;

-- name: GetDistinctSuns :many
SELECT DISTINCT sun FROM plants
WHERE sun IS NOT NULL AND sun != '' AND available = true
ORDER BY sun;

-- name: GetDistinctWaters :many
SELECT DISTINCT water FROM plants
WHERE water IS NOT NULL AND water != '' AND available = true
ORDER BY water;
