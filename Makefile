.PHONY: dev build test test-integration lint migrate sqlc proto swagger clean

# Go hot-reload with air
dev:
	go tool air

# Build binary
build:
	go build -o ./tmp/server ./cmd/server

# Run all Go tests
test:
	go test ./... -v -race

# Run integration tests (requires docker compose postgres/redis)
test-integration:
	go test -tags integration ./... -v -race

# Lint with golangci-lint
lint:
	go tool golangci-lint run ./...

# Run database migrations
migrate:
	go tool goose -dir migrations postgres "$(DATABASE_URL)" up

# Generate sqlc code
sqlc:
	go tool sqlc generate

# Generate protobuf/gRPC code
proto:
	go tool buf generate

# Generate OpenAPI/Swagger docs
swagger:
	go tool swag init -g cmd/server/main.go -o docs/swagger

# Remove build artifacts
clean:
	rm -rf tmp/ dist/ gen/
