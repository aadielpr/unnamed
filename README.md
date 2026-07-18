# EventLens

A web app for collecting photos from event guests into one shared gallery.

## Stack

- **Backend:** Go with the [Echo](https://echo.labstack.com/) web framework
- **Frontend:** [SolidJS](https://www.solidjs.com/) SPA built with [Vite](https://vitejs.dev/)
- **Database:** Postgres
- **Object storage:** S3-compatible (MinIO for local dev, Cloudflare R2 for production)
- **Migrations:** [goose](https://github.com/pressly/goose)

## Prerequisites

- Go 1.26+
- Node.js 22+
- Docker + Docker Compose (for the local dev stack)
- Postgres and MinIO, or Docker Compose to run them

## Quick start

1. Copy the example environment file and edit it:

   ```bash
   cp .env.example .env
   ```

2. Start the full local stack:

   ```bash
   make dev
   ```

   This builds the SPA and Go server, runs Postgres and MinIO, applies migrations, and starts the API on [http://localhost:8080](http://localhost:8080).

3. Check health:

   ```bash
   curl http://localhost:8080/api/health
   ```

   Expected response:

   ```json
   {"status":"ok","db":"reachable"}
   ```

## Development without Docker

If you already have Postgres and MinIO running locally:

1. Install frontend dependencies:

   ```bash
   make web-install
   ```

2. Build the SPA:

   ```bash
   make web-build
   ```

3. Apply database migrations:

   ```bash
   make migrate
   ```

4. Run the server:

   ```bash
   make run
   ```

## Migrations

Migrations live in `migrations/` and use `goose`.

- Apply migrations: `make migrate`
- Roll back the last migration: `make migrate-down`
- Check status: `go run github.com/pressly/goose/v3/cmd/goose@latest -dir migrations postgres "$DATABASE_URL" status`

## Tests

Run the full test suite:

```bash
make test
```

Tests run against a real Postgres database and a real S3-compatible store (MinIO by default). Set `TEST_DATABASE_URL`, `TEST_STORAGE_ENDPOINT`, and related variables to point to your test infrastructure.

## Project structure

```
.
├── cmd/server/         # Go server entry point
├── internal/           # Go application code
│   ├── config/         # Environment configuration
│   ├── db/             # Postgres connection
│   ├── handlers/       # HTTP handlers
│   ├── server/         # Echo server setup
│   └── storage/        # S3-compatible storage interface
├── migrations/         # Database migrations
├── web/                # SolidJS SPA
│   ├── src/            # Source files
│   └── dist/           # Build output (served by Go)
├── docker-compose.yml  # Local dev stack
├── Dockerfile          # Production container build
└── Makefile            # Common commands
```
