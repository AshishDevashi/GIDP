.PHONY: run dev build sqlc-generate tidy

run:
	go run ./cmd/api

dev:
	air

build:
	go build -o ./tmp/main ./cmd/api

sqlc-generate:
	sqlc generate

tidy:
	go mod tidy
