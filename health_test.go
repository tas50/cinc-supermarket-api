package supermarket

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthStatusReturnsStatus(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":"ok"}`)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	c := newTestClient(t, srv, false)

	h, _, err := c.Health.Status(context.Background())
	if err != nil {
		t.Fatalf("Health.Status: %v", err)
	}
	if h.Status != "ok" {
		t.Errorf("status = %q", h.Status)
	}
}

func TestHealthStatusIgnoresUnknownFields(t *testing.T) {
	// Forward-compat: new fields on the server response shouldn't
	// break decoding into our Health struct.
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/health", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"status":"ok","new_field":"future","version":42}`)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	c := newTestClient(t, srv, false)

	h, _, err := c.Health.Status(context.Background())
	if err != nil {
		t.Fatalf("Health.Status: %v", err)
	}
	if h.Status != "ok" {
		t.Errorf("status = %q", h.Status)
	}
}

func TestHealthMetricsReturnsRawJSON(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/metrics", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"cookbook_count":42,"users":{"total":7}}`)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	c := newTestClient(t, srv, false)

	got, _, err := c.Health.Metrics(context.Background())
	if err != nil {
		t.Fatalf("Health.Metrics: %v", err)
	}
	if got["cookbook_count"].(float64) != 42 {
		t.Errorf("metrics = %+v", got)
	}
}
