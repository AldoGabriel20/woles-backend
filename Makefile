APP_NAME   := woles-backend
BIN_DIR    := bin
MAIN       := ./cmd/main.go
BINARY     := $(BIN_DIR)/$(APP_NAME)
MIGRATION_DIR := internal/migration

.PHONY: run build test migrate-up migrate-down generate-mock lint

## run: start the API server locally (requires .env)
run:
	go run $(MAIN)

## build: compile the binary into bin/
build:
	mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BINARY) $(MAIN)

## test: run all unit and integration tests
test:
	go test ./... -v -race -timeout 120s

## migrate-up: apply all pending database migrations
migrate-up:
	go run ./cmd/migrate/main.go up

## migrate-down: roll back the last database migration
migrate-down:
	go run ./cmd/migrate/main.go down

## generate-mock: generate mocks for all port interfaces
generate-mock:
	go generate ./internal/port/...

## lint: run golangci-lint
lint:
	golangci-lint run ./...
