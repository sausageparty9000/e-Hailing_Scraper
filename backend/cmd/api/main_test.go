package main

import (
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestQuotesRequireRoute(t *testing.T) {
	for _, query := range []string{"", "pickup=KL", "dropoff=Pavilion"} {
		t.Run(query, func(t *testing.T) {
			w := httptest.NewRecorder()
			quotesHandler(w, httptest.NewRequest(http.MethodGet, "/api/quotes?"+query, nil))
			if w.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400", w.Code)
			}
			var body map[string]string
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil || body["error"] == "" {
				t.Fatalf("expected JSON error, got %s", w.Body.String())
			}
		})
	}
}

func TestQuotes(t *testing.T) {
	for _, tc := range []struct {
		name, distance string
		prices         []float64
	}{
		{"default distance", "", []float64{20.23, 17.83, 16.53}},
		{"explicit distance", "10", []float64{22.40, 20, 18.70}},
		{"zero defaults", "0", []float64{20.23, 17.83, 16.53}},
		{"negative defaults", "-1", []float64{20.23, 17.83, 16.53}},
		{"invalid defaults", "abc", []float64{20.23, 17.83, 16.53}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			before := time.Now()
			w := httptest.NewRecorder()
			quotesHandler(w, httptest.NewRequest(http.MethodGet, "/api/quotes?pickup=KL+Sentral&dropoff=Pavilion&distanceKm="+tc.distance, nil))
			if w.Code != http.StatusOK || w.Header().Get("Content-Type") != "application/json" {
				t.Fatalf("unexpected response: status %d, headers %v", w.Code, w.Header())
			}
			var body quoteResponse
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if body.Pickup != "KL Sentral" || body.Dropoff != "Pavilion" {
				t.Fatalf("route not preserved: %+v", body)
			}
			if body.ObservedAt.Before(before) || body.ObservedAt.After(time.Now()) {
				t.Fatalf("invalid observation time: %v", body.ObservedAt)
			}
			if len(body.Quotes) != 3 {
				t.Fatalf("got %d quotes, want 3", len(body.Quotes))
			}
			providers := []string{"Grab", "AirAsia Ride", "Maxim"}
			for i, q := range body.Quotes {
				if q.Provider != providers[i] || math.Abs(q.Price-tc.prices[i]) > 0.000001 {
					t.Errorf("unexpected quote %d: %+v", i, q)
				}
				if !q.IsSample || q.Currency != "MYR" {
					t.Errorf("quote must identify sample data and currency: %+v", q)
				}
			}
		})
	}
}

func TestCORS(t *testing.T) {
	for _, origin := range []string{"", "https://example.com"} {
		t.Run(origin, func(t *testing.T) {
			t.Setenv("WEB_ORIGIN", origin)
			wantOrigin := origin
			if wantOrigin == "" {
				wantOrigin = "http://localhost:5173"
			}
			for _, method := range []string{http.MethodGet, http.MethodOptions} {
				called := false
				handler := withCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					called = true
					w.WriteHeader(http.StatusAccepted)
				}))
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, httptest.NewRequest(method, "/api/quotes", nil))
				if w.Header().Get("Access-Control-Allow-Origin") != wantOrigin {
					t.Errorf("unexpected allowed origin: %s", w.Header().Get("Access-Control-Allow-Origin"))
				}
				if method == http.MethodOptions {
					if called || w.Code != http.StatusNoContent || w.Body.Len() != 0 {
						t.Error("preflight should return empty 204 without calling downstream handler")
					}
					if w.Header().Get("Access-Control-Allow-Methods") != "GET, OPTIONS" || w.Header().Get("Access-Control-Allow-Headers") != "Content-Type" {
						t.Error("missing preflight permissions")
					}
				} else if !called || w.Code != http.StatusAccepted {
					t.Error("GET should reach downstream handler")
				}
			}
		})
	}
}
