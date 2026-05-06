.PHONY: run test build docker-up docker-down

run:
	go run ./cmd/api

test:
	go test ./...

build:
	go build -o bin/saas-api ./cmd/api

docker-up:
	docker compose up --build

docker-down:
	docker compose down

