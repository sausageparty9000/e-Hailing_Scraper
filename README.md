# farelah

A starter e-hailing fare comparison application for Malaysia. The interface currently displays explicitly labelled sample data; it does not scrape provider applications or claim to show live fares.

## Project layout

```text
frontend/  React, TypeScript and Vite interface
backend/   Go HTTP API
docker/    Docker Compose configuration and launcher
```

## Run everything with Docker

From Git Bash, WSL, Linux, or macOS:

```sh
sh ./docker/run.sh
```

Open `http://localhost:3000`. To use a different port:

```sh
APP_PORT=4000 sh ./docker/run.sh
```

Stop all services and remove their containers and network:

```sh
sh ./docker/stop.sh
```

## Run the web interface locally

```powershell
cd frontend
npm install
npm run dev
```

Open `http://localhost:5173`.

## Run the Go API locally

Install Go 1.24 or newer, then run:

```powershell
cd backend
go run ./cmd/api
```

The API listens on `http://localhost:8080`. Vite forwards `/api` requests to it during development.

## Unit tests

Run the Go API tests with Docker Desktop running (no local Go installation needed):

```sh
sh ./docker/test.sh
```

From PowerShell, you can run the same tests directly:

```powershell
docker compose -f docker/compose.yml run --build --rm --no-deps backend-test
```

The isolated test container reports test results and coverage, then is removed. It does not start the application services. A failed test returns a nonzero exit code. The test service is excluded from normal startup.

With Go installed locally:

```sh
cd backend
go test ./... -v -cover -count=1
```

Tests live alongside the API in `backend/cmd/api/main_test.go`. They cover required route inputs, sample fares for default and explicit distances, distance fallbacks, response metadata, and CORS preflight behavior. These tests use in-memory HTTP requests and do not contact ride providers. Frontend tests and database integration tests are not included yet.

## Next milestones

1. Replace typed place names with map autocomplete and coordinates.
2. Store timestamped, user-contributed or authorized fare observations in PostgreSQL.
3. Add provider integrations only where permission or an official API is available.
4. Add freshness, toll, promotion, and availability indicators.
