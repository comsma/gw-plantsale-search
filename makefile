# Include variables from the .envrc file
include .envrc

# ==================================================================================== #
# HELPERS
# ==================================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

.PHONY: serve/dev
## serve/dev: run the web server locally
serve/dev:
	npm run css:watch & \
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
.PHONY: db/migrations/new
db/migrations/new:
	DATABASE_URL="$(DATABASE_URL)" go run ./cmd/plantsale migrate create --name=$(name)


## sqlc/generate: regenerate models from queries
.PHONY: sqlc/generate
sqlc/generate:
	sqlc generate

## ingest/plants: ingest plants from JSON and fetch iNaturalist data
.PHONY: ingest/plants
ingest/plants:
	DATABASE_URL="$(DATABASE_URL)" go run ./cmd/plantsale ingest plants --plant-list=plant_sale_list.json

## ingest/refresh-inat: re-fetch iNaturalist data for all plants in the database
.PHONY: ingest/refresh-inat
ingest/refresh-inat:
	DATABASE_URL="$(DATABASE_URL)" go run ./cmd/plantsale ingest refresh-inat
