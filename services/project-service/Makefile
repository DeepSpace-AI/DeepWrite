APP_NAME=project-service

.PHONY: build run test lint

build:
	go build -o bin/$(APP_NAME) ./cmd/server

run:
	go run ./cmd/server

test:
	go test ./...

lint:
	golangci-lint run ./...
