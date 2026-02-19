.PHONY: dev build test lint migrate sqlc proto swagger clean

# Go hot-reload with air
dev:
	air

# Build binary
build:
	go build -o ./tmp/server ./cmd/server

# Run all Go tests
test:
	go test ./... -v -race

# Lint with golangci-lint
lint:
	golangci-lint run ./...

# Run database migrations
migrate:
	@echo "TODO: implement with golang-migrate"

# Generate sqlc code
sqlc:
	sqlc generate

# Generate protobuf/gRPC code
proto:
	buf generate

# Generate OpenAPI/Swagger docs
swagger:
	swag init -g cmd/server/main.go -o docs/swagger

# Remove build artifacts
clean:
	rm -rf tmp/ dist/ gen/
