.PHONY: dev infra run web-dev web-install web-build test migrate migrate-down

# .env is the single source of truth for connection settings; it overrides
# nothing if absent. Both Docker and native-local setups edit .env, not Make.
-include .env

GOOSE := go run github.com/pressly/goose/v3/cmd/goose@latest

# Start local infrastructure (Postgres and MinIO) in Docker.
infra:
	docker compose up -d --build

# Run the Go API server with live reload.
run:
	air

# Start the Vite dev server for the frontend.
web-dev:
	cd web && pnpm run dev

# Run both the API and the frontend dev server at the same time.
dev:
	cd web && pnpm exec concurrently --kill-others --names "api,web" "make -C .. run" "pnpm run dev"

# Install frontend dependencies.
web-install:
	cd web && pnpm install

# Build the SPA for production.
web-build:
	cd web && pnpm run build

# Run all Go tests.
test:
	go test ./...

# Apply database migrations.
migrate:
	$(GOOSE) -dir migrations postgres "$(DATABASE_URL)" up

# Roll back the last database migration.
migrate-down:
	$(GOOSE) -dir migrations postgres "$(DATABASE_URL)" down
