package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type quote struct {
	ID         int64     `json:"id"`
	Provider   string    `json:"provider"`
	Service    string    `json:"service"`
	Price      float64   `json:"price"`
	Currency   string    `json:"currency"`
	Pickup     string    `json:"pickup"`
	Dropoff    string    `json:"dropoff"`
	ObservedAt time.Time `json:"observedAt"`
	Source     string    `json:"source"`
}

type quoteStore interface {
	Insert(context.Context, quote) (quote, error)
	List(context.Context, string, string) ([]quote, error)
}

type postgresStore struct{ pool *pgxpool.Pool }

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, envOr("DATABASE_URL", "postgres://farelah:farelah@localhost:5432/farelah"))
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("database unavailable: %v", err)
	}
	if err := migrate(ctx, pool); err != nil {
		log.Fatalf("database migration failed: %v", err)
	}

	port := envOr("PORT", "8080")
	server := &http.Server{Addr: ":" + port, Handler: withCORS(newRouter(postgresStore{pool})), ReadHeaderTimeout: 5 * time.Second}
	log.Printf("API listening on http://localhost:%s", port)
	log.Fatal(server.ListenAndServe())
}

func newRouter(store quoteStore) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /api/quotes", func(w http.ResponseWriter, r *http.Request) {
		pickup := strings.TrimSpace(r.URL.Query().Get("pickup"))
		dropoff := strings.TrimSpace(r.URL.Query().Get("dropoff"))
		if pickup == "" || dropoff == "" {
			writeError(w, http.StatusBadRequest, "pickup and dropoff are required")
			return
		}
		quotes, err := store.List(r.Context(), pickup, dropoff)
		if err != nil {
			log.Printf("list quotes: %v", err)
			writeError(w, http.StatusInternalServerError, "could not load quotes")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"pickup": pickup, "dropoff": dropoff, "quotes": quotes})
	})
	mux.HandleFunc("POST /api/quotes", func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 4096)
		var input quote
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&input); err != nil {
			writeError(w, http.StatusBadRequest, "invalid quote JSON")
			return
		}
		var extra any
		if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
			writeError(w, http.StatusBadRequest, "expected one JSON object")
			return
		}
		input.Pickup = strings.TrimSpace(input.Pickup)
		input.Dropoff = strings.TrimSpace(input.Dropoff)
		input.Provider = strings.TrimSpace(input.Provider)
		input.Service = strings.TrimSpace(input.Service)
		if err := validateQuote(input, time.Now()); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		input.Currency = "MYR"
		input.Source = "user_submitted"
		created, err := store.Insert(r.Context(), input)
		if err != nil {
			log.Printf("insert quote: %v", err)
			writeError(w, http.StatusInternalServerError, "could not save quote")
			return
		}
		writeJSON(w, http.StatusCreated, created)
	})
	return mux
}

func validateQuote(q quote, now time.Time) error {
	if q.Pickup == "" || q.Dropoff == "" || q.Provider == "" || q.Service == "" {
		return errors.New("pickup, dropoff, provider, and service are required")
	}
	for _, value := range []string{q.Pickup, q.Dropoff, q.Provider, q.Service} {
		if len(value) > 120 {
			return errors.New("text fields must be 120 characters or fewer")
		}
	}
	if math.IsNaN(q.Price) || math.IsInf(q.Price, 0) || q.Price <= 0 || q.Price > 10000 || math.Abs(math.Round(q.Price*100)-q.Price*100) > 0.000001 {
		return errors.New("price must be between RM0.01 and RM10000 with at most two decimal places")
	}
	if q.ObservedAt.IsZero() || q.ObservedAt.After(now.Add(5*time.Minute)) || q.ObservedAt.Before(now.Add(-24*time.Hour)) {
		return errors.New("observedAt must be within the last 24 hours")
	}
	return nil
}

func migrate(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS fare_observations (
		id BIGSERIAL PRIMARY KEY,
		pickup TEXT NOT NULL,
		dropoff TEXT NOT NULL,
		provider TEXT NOT NULL,
		service TEXT NOT NULL,
		price_cents INTEGER NOT NULL CHECK (price_cents > 0),
		currency TEXT NOT NULL DEFAULT 'MYR',
		observed_at TIMESTAMPTZ NOT NULL,
		source TEXT NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS fare_route_time_idx ON fare_observations (lower(pickup), lower(dropoff), observed_at DESC);`)
	return err
}

func (s postgresStore) Insert(ctx context.Context, q quote) (quote, error) {
	err := s.pool.QueryRow(ctx, `INSERT INTO fare_observations
		(pickup, dropoff, provider, service, price_cents, currency, observed_at, source)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`,
		q.Pickup, q.Dropoff, q.Provider, q.Service, int(math.Round(q.Price*100)), q.Currency, q.ObservedAt, q.Source).Scan(&q.ID)
	return q, err
}

func (s postgresStore) List(ctx context.Context, pickup, dropoff string) ([]quote, error) {
	rows, err := s.pool.Query(ctx, `SELECT DISTINCT ON (lower(provider), lower(service))
		id, provider, service, price_cents, currency, pickup, dropoff, observed_at, source
		FROM fare_observations
		WHERE lower(pickup) = lower($1) AND lower(dropoff) = lower($2)
		AND observed_at >= NOW() - INTERVAL '24 hours'
		ORDER BY lower(provider), lower(service), observed_at DESC, id DESC`, pickup, dropoff)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	quotes := []quote{}
	for rows.Next() {
		var q quote
		var cents int
		if err := rows.Scan(&q.ID, &q.Provider, &q.Service, &cents, &q.Currency, &q.Pickup, &q.Dropoff, &q.ObservedAt, &q.Source); err != nil {
			return nil, err
		}
		q.Price = float64(cents) / 100
		quotes = append(quotes, q)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	sort.Slice(quotes, func(i, j int) bool { return quotes[i].Price < quotes[j].Price })
	return quotes, nil
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("encode response: %v", err)
	}
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", envOr("WEB_ORIGIN", "http://localhost:5173"))
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
