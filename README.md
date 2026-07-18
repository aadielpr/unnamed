# EventLens

A web app for collecting photos from event guests into one shared gallery.

## Stack

- **Backend:** Go with the [Echo](https://echo.labstack.com/) web framework
- **Frontend:** [SolidJS](https://www.solidjs.com/) SPA built with [Vite](https://vitejs.dev/)
- **Database:** Postgres
- **Object storage:** S3-compatible (MinIO for local dev, Cloudflare R2 for production)
- **Migrations:** [goose](https://github.com/pressly/goose)
- **Package manager:** [pnpm](https://pnpm.io/)

## Prerequisites

- Go 1.26+
- Node.js 22+ (with [pnpm](https://pnpm.io/installation))
- Docker + Docker Compose (for the local database and object storage)
- A running Postgres server for local development (started via `make infra`)

## Quick start

1. Copy the example environment file and edit it:

   ```bash
   cp .env.example .env
   ```

2. Install frontend dependencies:

   ```bash
   make web-install
   ```

3. Start the local infrastructure (Postgres and MinIO):

   ```bash
   make infra
   ```

4. Apply database migrations:

   ```bash
   make migrate
   ```

5. Run the API and the frontend dev server:

   ```bash
   make dev
   ```

   The API uses `air` for live reload. The frontend runs the Vite dev server.

6. Check health:

   ```bash
   curl http://localhost:8080/api/health
   ```

   Expected response:

   ```json
   {"status":"ok","db":"reachable"}
   ```

## Available commands

- `make infra` — start Postgres and MinIO in Docker.
- `make run` — run the Go API with `air` live reload.
- `make web-dev` — run the Vite dev server.
- `make dev` — run both the API and the frontend dev server together.
- `make web-install` — install frontend dependencies.
- `make web-build` — build the SPA for production.
- `make migrate` — apply database migrations.
- `make migrate-down` — roll back the last migration.
- `make test` — run all Go tests.

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
├── docker-compose.yml  # Local infrastructure (Postgres + MinIO)
├── Dockerfile          # Production container build
├── Makefile            # Common commands
└── .air.toml           # Air live-reload configuration
```
