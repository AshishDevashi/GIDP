.PHONY: run dev build sqlc-generate tidy up down logs migrate-up migrate-down

DB_URL ?= postgres://postgres:postgres@localhost:5432/wolf_platform?sslmode=disable

run:
	go run ./cmd/api

dev:
	air

up:
	docker compose up -d --build

down:
	docker compose down

logs:
	docker compose logs -f

migrate-up:
	GOOSE_DRIVER=postgres GOOSE_DBSTRING="$(DB_URL)" goose -dir db/migrations up

migrate-down:
	GOOSE_DRIVER=postgres GOOSE_DBSTRING="$(DB_URL)" goose -dir db/migrations down

build:
	go build -o ./tmp/main ./cmd/api

sqlc-generate:
	sqlc generate

tidy:
	go mod tidy
