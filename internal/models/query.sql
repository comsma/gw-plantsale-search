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
    plant_search_view.id,
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
    available,
    image_url,
    EXISTS(SELECT 1 FROM favorites_list f WHERE f.plant_id = plant_search_view.id AND f.user_id = sqlc.arg(user_id)) AS is_favorited

FROM plant_search_view
WHERE available = true
  AND (
    sqlc.arg(query)::text = ''
    OR search_vector @@ websearch_to_tsquery('english', sqlc.arg(query)::text)
  )
  AND (sqlc.arg(section)::text = '' OR section = sqlc.arg(section)::text)
  AND (sqlc.arg(color)::text   = '' OR color   = sqlc.arg(color)::text)
  AND (sqlc.arg(sun)::text     = '' OR sun     = sqlc.arg(sun)::text)
  AND (sqlc.arg(water)::text   = '' OR water   = sqlc.arg(water)::text)
ORDER BY
    CASE WHEN sqlc.arg(query)::text != '' THEN ts_rank(search_vector, websearch_to_tsquery('english', sqlc.arg(query)::text)) END DESC NULLS LAST,
    CASE WHEN sqlc.arg(sort_column)::text = 'height' THEN height_sort END ASC NULLS LAST,
    CASE WHEN sqlc.arg(sort_column)::text = 'price'  THEN price END ASC NULLS LAST,
    common ASC;

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
    i.attribution,
    EXISTS(SELECT 1 FROM favorites_list f WHERE f.plant_id = p.id AND f.user_id = sqlc.arg(user_id)) AS is_favorited

FROM plants p
LEFT JOIN inatrualist i ON i.plant_id = p.id
LEFT JOIN favorites_list f ON f.plant_id = p.id
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
) ON CONFLICT (plant_id) DO UPDATE SET
    summary      = EXCLUDED.summary,
    image_url    = EXCLUDED.image_url,
    attribution  = EXCLUDED.attribution,
    last_updated = NOW();

-- name: CreateFavoritePlant :exec
INSERT INTO favorites_list (
    plant_id,
    user_id
) values (
          sqlc.arg(plant_id),
          sqlc.arg(user_id)
         );

-- name: DeleteFavoritePlant :exec
DELETE FROM favorites_list
    where
        plant_id = sqlc.arg(plant_id) AND
        user_id = sqlc.arg(user_id);

-- name: GetFavoritePlants :many
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
FROM favorites_list f
INNER JOIN plants p ON p.id = f.plant_id
LEFT JOIN inatrualist i ON i.plant_id = p.id
WHERE f.user_id = sqlc.arg(user_id)
ORDER BY p.common ASC;

-- name: GetFavoriteCounts :many
SELECT
    p.id,
    p.common,
    COUNT(f.id) AS favorite_count
FROM plants p
INNER JOIN favorites_list f ON f.plant_id = p.id
GROUP BY p.id, p.common
ORDER BY favorite_count DESC, p.common ASC;

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
