.PHONY: build run test lint

build:
	go build ./cmd/server

run:
	go run ./cmd/server

test:
	go test ./...

lint:
	golangci-lint run ./...
