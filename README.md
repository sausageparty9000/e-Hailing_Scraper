# farelah

A starter e-hailing fare comparison application for Malaysia. People can submit fares they observe in provider apps, and the site compares the most recent observation for each provider and service on a route. It does not scrape provider applications or claim to show live fares.

## Project layout

```text
frontend/  React, TypeScript and Vite interface
backend/   Go HTTP API and PostgreSQL storage
docker/    Docker Compose configuration and launchers
```

## Run everything with Docker

From Git Bash, WSL, Linux, or macOS:

```sh
sh ./docker/run.sh
```

Open `http://localhost:3000`. Docker also starts PostgreSQL with a persistent volume. To use a different port:

```sh
APP_PORT=4000 sh ./docker/run.sh
```

Stop all services and remove their containers and network:

```sh
sh ./docker/stop.sh
```

The stop script keeps the database volume. Running `run.sh` again restores previously submitted fares. The database is available only within the Docker network; its starter credentials are in `docker/compose.yml`, so change them before deploying outside a local development machine.

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

The API listens on `http://localhost:8080`. Vite forwards `/api` requests to it during development. For local Go development, start PostgreSQL separately and set `DATABASE_URL` to its connection string.

## Fare observations

Choose a route, select **Compare rides**, then use **Add a fare** to enter a price just seen in a provider's official app. Results are user submitted observations, never live quotes. The API accepts observations from the past 24 hours and displays the latest one for each provider and service on the selected route, sorted by price. Quotes older than 30 minutes are labelled **Older quote**. Route names are matched without regard to letter case, but spelling must otherwise match.

The API offers `GET /api/quotes?pickup=...&dropoff=...` and `POST /api/quotes` with JSON fields `pickup`, `dropoff`, `provider`, `service`, `price`, and `observedAt` (ISO 8601 timestamp). PostgreSQL stores prices as integer cents and retains them when containers restart.

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

Tests live alongside the API in `backend/cmd/api/main_test.go`. They cover quote submission, retrieval, validation, database error responses, and CORS preflight behavior. These tests use in-memory HTTP requests and do not contact ride providers. Frontend tests and database integration tests are not included yet.

## Next milestones

1. Replace typed place names with map autocomplete and coordinates.
2. Add route coordinates so differently spelled place names can still match.
3. Add provider integrations where permission or an official API is available.
4. Add toll, promotion, and availability indicators.
