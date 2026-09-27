package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type fakeStore struct {
	quotes []quote
	err    error
}

func (s *fakeStore) Insert(_ context.Context, q quote) (quote, error) {
	if s.err != nil {
		return quote{}, s.err
	}
	q.ID = int64(len(s.quotes) + 1)
	s.quotes = append(s.quotes, q)
	return q, nil
}

func (s *fakeStore) List(_ context.Context, pickup, dropoff string) ([]quote, error) {
	if s.err != nil {
		return nil, s.err
	}
	result := []quote{}
	for _, q := range s.quotes {
		if q.Pickup == pickup && q.Dropoff == dropoff {
			result = append(result, q)
		}
	}
	return result, nil
}

func TestQuoteAPI(t *testing.T) {
	store := &fakeStore{}
	api := newRouter(store)
	observedAt := time.Now().UTC().Truncate(time.Second)
	input := quote{Pickup: "KL Sentral", Dropoff: "Pavilion", Provider: "Grab", Service: "JustGrab", Price: 18.50, ObservedAt: observedAt}
	body, _ := json.Marshal(input)
	w := httptest.NewRecorder()
	api.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/quotes", bytes.NewReader(body)))
	if w.Code != http.StatusCreated {
		t.Fatalf("POST status = %d, body = %s", w.Code, w.Body.String())
	}
	var created quote
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.ID != 1 || created.Source != "user_submitted" || created.Currency != "MYR" {
		t.Fatalf("unexpected saved quote: %+v", created)
	}

	w = httptest.NewRecorder()
	api.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/quotes?pickup=KL+Sentral&dropoff=Pavilion", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("GET status = %d", w.Code)
	}
	var listed struct {
		Quotes []quote `json:"quotes"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &listed); err != nil {
		t.Fatal(err)
	}
	if len(listed.Quotes) != 1 || listed.Quotes[0].Price != 18.50 {
		t.Fatalf("unexpected list: %+v", listed)
	}
}

func TestInvalidQuotes(t *testing.T) {
	for _, tc := range []struct {
		name  string
		input quote
	}{
		{"missing route", quote{Provider: "Grab", Service: "JustGrab", Price: 10, ObservedAt: time.Now()}},
		{"zero price", quote{Pickup: "A", Dropoff: "B", Provider: "Grab", Service: "JustGrab", ObservedAt: time.Now()}},
		{"negative price", quote{Pickup: "A", Dropoff: "B", Provider: "Grab", Service: "JustGrab", Price: -5, ObservedAt: time.Now()}},
		{"too many decimals", quote{Pickup: "A", Dropoff: "B", Provider: "Grab", Service: "JustGrab", Price: 1.234, ObservedAt: time.Now()}},
		{"old observation", quote{Pickup: "A", Dropoff: "B", Provider: "Grab", Service: "JustGrab", Price: 10, ObservedAt: time.Now().Add(-25 * time.Hour)}},
		{"future observation", quote{Pickup: "A", Dropoff: "B", Provider: "Grab", Service: "JustGrab", Price: 10, ObservedAt: time.Now().Add(time.Hour)}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body, _ := json.Marshal(tc.input)
			w := httptest.NewRecorder()
			newRouter(&fakeStore{}).ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/quotes", bytes.NewReader(body)))
			if w.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400", w.Code)
			}
		})
	}
}

func TestRouteAndStoreErrors(t *testing.T) {
	for _, target := range []string{"/api/quotes", "/api/quotes?pickup=A", "/api/quotes?dropoff=B"} {
		w := httptest.NewRecorder()
		newRouter(&fakeStore{}).ServeHTTP(w, httptest.NewRequest(http.MethodGet, target, nil))
		if w.Code != http.StatusBadRequest {
			t.Errorf("%s: status = %d", target, w.Code)
		}
	}
	w := httptest.NewRecorder()
	newRouter(&fakeStore{err: errors.New("db unavailable")}).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/quotes?pickup=A&dropoff=B", nil))
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("database error status = %d", w.Code)
	}
}

func TestCORS(t *testing.T) {
	t.Setenv("WEB_ORIGIN", "https://example.com")
	w := httptest.NewRecorder()
	withCORS(newRouter(&fakeStore{})).ServeHTTP(w, httptest.NewRequest(http.MethodOptions, "/api/quotes", nil))
	if w.Code != http.StatusNoContent || w.Header().Get("Access-Control-Allow-Origin") != "https://example.com" || w.Header().Get("Access-Control-Allow-Methods") != "GET, POST, OPTIONS" {
		t.Fatalf("unexpected preflight: %d %v", w.Code, w.Header())
	}
}
