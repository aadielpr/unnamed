.PHONY: dev test web-install web-build migrate migrate-down

DATABASE_URL ?= postgres://eventlens:eventlens@localhost:5432/eventlens?sslmode=disable
GOOSE := go run github.com/pressly/goose/v3/cmd/goose@latest

# Start the full local development stack with Docker Compose.
dev:
	docker compose up --build

# Install frontend dependencies.
web-install:
	cd web && npm install

# Build the SPA for production.
web-build:
	cd web && npm run build

# Run all Go tests.
test:
	go test ./...

# Apply database migrations.
migrate:
	$(GOOSE) -dir migrations postgres "$(DATABASE_URL)" up

# Roll back the last database migration.
migrate-down:
	$(GOOSE) -dir migrations postgres "$(DATABASE_URL)" down

# Run the Go server locally (requires DATABASE_URL and storage env vars).
run:
	go run ./cmd/server
