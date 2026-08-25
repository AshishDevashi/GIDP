.PHONY: run dev build sqlc-generate tidy up down logs

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

build:
	go build -o ./tmp/main ./cmd/api

sqlc-generate:
	sqlc generate

tidy:
	go mod tidy
