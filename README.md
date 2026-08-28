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

## Next milestones

1. Replace typed place names with map autocomplete and coordinates.
2. Store timestamped, user-contributed or authorized fare observations in PostgreSQL.
3. Add provider integrations only where permission or an official API is available.
4. Add freshness, toll, promotion, and availability indicators.
