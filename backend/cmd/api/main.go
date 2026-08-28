package main

import (
	"encoding/json"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"
)

type quote struct {
	Provider       string  `json:"provider"`
	Service        string  `json:"service"`
	Price          float64 `json:"price"`
	Currency       string  `json:"currency"`
	PickupMinutes  int     `json:"pickupMinutes"`
	JourneyMinutes int     `json:"journeyMinutes"`
	IsSample       bool    `json:"isSample"`
}

type quoteResponse struct {
	Pickup     string    `json:"pickup"`
	Dropoff    string    `json:"dropoff"`
	ObservedAt time.Time `json:"observedAt"`
	Quotes     []quote   `json:"quotes"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /api/quotes", quotesHandler)

	port := envOr("PORT", "8080")
	server := &http.Server{
		Addr:              ":" + port,
		Handler:           withCORS(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("API listening on http://localhost:%s", port)
	log.Fatal(server.ListenAndServe())
}

func quotesHandler(w http.ResponseWriter, r *http.Request) {
	pickup := r.URL.Query().Get("pickup")
	dropoff := r.URL.Query().Get("dropoff")
	if pickup == "" || dropoff == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "pickup and dropoff are required"})
		return
	}

	distance, _ := strconv.ParseFloat(r.URL.Query().Get("distanceKm"), 64)
	if distance <= 0 {
		distance = 8.5
	}
	base := math.Round((5.5+distance*1.45)*100) / 100

	writeJSON(w, http.StatusOK, quoteResponse{
		Pickup: pickup, Dropoff: dropoff, ObservedAt: time.Now(),
		Quotes: []quote{
			{Provider: "Grab", Service: "JustGrab", Price: base + 2.40, Currency: "MYR", PickupMinutes: 4, JourneyMinutes: 18, IsSample: true},
			{Provider: "AirAsia Ride", Service: "Compact", Price: base, Currency: "MYR", PickupMinutes: 7, JourneyMinutes: 19, IsSample: true},
			{Provider: "Maxim", Service: "Economy", Price: base - 1.30, Currency: "MYR", PickupMinutes: 9, JourneyMinutes: 20, IsSample: true},
		},
	})
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
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
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
