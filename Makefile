.PHONY: all build run test test-race test-cover tidy lint clean migrate-up migrate-down seed

APP_NAME = geotrack-api
BUILD_DIR = bin

all: build

build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME) cmd/api/main.go
	@echo "Build successful: $(BUILD_DIR)/$(APP_NAME)"

run:
	@echo "Running GeoTrack API..."
	go run cmd/api/main.go

migrate-up:
	@echo "Running database migrations up..."
	go run cmd/migrate/main.go up

migrate-down:
	@echo "Rolling back database migrations..."
	go run cmd/migrate/main.go down

seed:
	@echo "Seeding initial entity data..."
	go run cmd/migrate/main.go seed

test:
	@echo "Running unit tests..."
	go test -v ./...

test-race:
	@echo "Running tests with data race detector..."
	go test -race -v ./...

test-cover:
	@echo "Generating coverage report..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage HTML generated at coverage.html"

tidy:
	@echo "Tidying go module dependencies..."
	go mod tidy

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR) coverage.out coverage.html

