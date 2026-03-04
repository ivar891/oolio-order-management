.PHONY: build run test lint docker-build docker-up docker-down clean

build:
	go build -o bin/server ./cmd/server/

run:
	go run ./cmd/server/

test:
	go test -v -race -coverprofile=coverage.out ./cmd/... ./internal/...

lint:
	golangci-lint run ./...

docker-build:
	docker build -t order-food-api .

docker-up:
	docker compose up -d

docker-down:
	docker compose down

clean:
	rm -rf bin/ coverage.out
