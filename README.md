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
- For local infrastructure, either:
  - **Docker + Docker Compose** (recommended — runs Postgres and MinIO with one command), or
  - **Native installs** of Postgres and MinIO on your machine.

The app does not depend on Docker — it reads connection settings only from environment variables. Docker is just the easiest way to get Postgres and MinIO running. If you already run them natively, skip Docker and point `.env` at your local instances. See [ADR 0002](docs/adr/0002-infrastructure-backend-agnostic.md).

## Quick start

1. Copy the example environment file and edit it:

   ```bash
   cp .env.example .env
   ```

   The defaults assume the Docker path (Postgres on port 5433, MinIO on
   port 9000). If you run Postgres or MinIO natively, edit the relevant
   values in `.env` to match your setup.

2. Install frontend dependencies:

   ```bash
   make web-install
   ```

3. Start the local infrastructure (Postgres and MinIO):

   - **With Docker (recommended):**

     ```bash
     make infra
     ```

   - **Without Docker:** start your local Postgres and MinIO yourself. The
     `DATABASE_URL` and `STORAGE_*` values in `.env` must point at them.

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

- `make infra` — start Postgres and MinIO in Docker (optional; skip if you run them natively).
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
