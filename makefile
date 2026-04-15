## serve/dev: run the web server locally
serve/dev:
	DATABASE_URL="$(DATABASE_URL)" go run ./cmd/plantsale serve

## db/up: apply all pending migrations
.PHONY: db/up
db/up:
	DATABASE_URL="$(DATABASE_URL)" go run ./cmd/plantsale migrate up

## db/down: roll back the last migration
.PHONY: db/down
db/down:
	DATABASE_URL="$(DATABASE_URL)" go run ./cmd/plantsale migrate down

## db/status: print migration status
.PHONY: db/status
db/status:
	DATABASE_URL="$(DATABASE_URL)" go run ./cmd/plantsale migrate status

## db/migrations/new name=<name>: create a new migration file
db/migrations/new:
	DATABASE_URL="$(DATABASE_URL)" go run ./cmd/plantsale migrate create $(name)

## sqlc/generate: regenerate models from queries
sqlc/generate:
	sqlc generate

## ingest/plants: ingest plants from JSON and fetch iNaturalist data
ingest/plants:
	DATABASE_URL="$(DATABASE_URL)" go run ./cmd/plantsale ingest plants

## ingest/refresh-inat: re-fetch iNaturalist data for all plants in the database
ingest/refresh-inat:
	DATABASE_URL="$(DATABASE_URL)" go run ./cmd/plantsale ingest refresh-inat
