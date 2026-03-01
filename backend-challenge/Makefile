.PHONY: build run test lint migrate seed docker-build docker-up docker-down clean

build:
	go build -o bin/server ./cmd/server/

run:
	go run ./cmd/server/

test:
	go test -v -race -coverprofile=coverage.out ./...

lint:
	golangci-lint run ./...

migrate:
	go run ./scripts/migrate.go

seed:
	go run ./scripts/seed.go

docker-build:
	docker build -t order-food-api .

docker-up:
	docker compose up -d

docker-down:
	docker compose down -v

clean:
	rm -rf bin/ coverage.out
