package supermarket

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// TestTransportSetsCommonHeaders covers the headers every request gets
// regardless of resource: User-Agent (always), Accept (always), and
// Content-Type only when there's a body to send.
func TestTransportSetsCommonHeaders(t *testing.T) {
	var (
		gotUA          string
		gotAccept      string
		gotContentType string
	)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/cookbooks", func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		gotAccept = r.Header.Get("Accept")
		gotContentType = r.Header.Get("Content-Type")
		_, _ = io.WriteString(w, `{"total":0,"start":0,"items":[]}`)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	c := newTestClient(t, srv, false)

	if _, _, err := c.Cookbooks.List(context.Background(), ListOptions{}); err != nil {
		t.Fatalf("List: %v", err)
	}
	if gotUA != "cinc-supermarket-go" {
		t.Errorf("User-Agent = %q, want default cinc-supermarket-go", gotUA)
	}
	if gotAccept != "application/json" {
		t.Errorf("Accept = %q, want application/json", gotAccept)
	}
	if gotContentType != "" {
		t.Errorf("Content-Type on a GET (no body) = %q, want empty", gotContentType)
	}
}

// TestTransportRetriesGETOn5xx confirms idempotent GETs are retried up
// to MaxRetries times against a transient 500.
func TestTransportRetriesGETOn5xx(t *testing.T) {
	var hits atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/cookbooks/apache2", func(w http.ResponseWriter, _ *http.Request) {
		if hits.Add(1) <= 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, _ = io.WriteString(w, `{"name":"apache2"}`)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	c := newTestClient(t, srv, false)

	if _, _, err := c.Cookbooks.Get(context.Background(), "apache2"); err != nil {
		t.Fatalf("Get after retry: %v", err)
	}
	if got := hits.Load(); got != 3 {
		t.Errorf("server hits = %d, want 3 (2 transient + 1 success)", got)
	}
}

// TestTransportDoesNotRetryNonGET makes sure POSTs and DELETEs never
// silently re-execute on 5xx — replays of writes are dangerous.
func TestTransportDoesNotRetryNonGET(t *testing.T) {
	var hits atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/cookbooks/apache2", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %q", r.Method)
		}
		hits.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	c := newTestClient(t, srv, true)

	if _, _, err := c.Cookbooks.Delete(context.Background(), "apache2"); err == nil {
		t.Error("expected an error from a 500-returning DELETE")
	}
	if got := hits.Load(); got != 1 {
		t.Errorf("DELETE hits = %d, want 1 (no retries on writes)", got)
	}
}

// TestTransportRespectsMaxRetriesZero turns retries off entirely.
func TestTransportRespectsMaxRetriesZero(t *testing.T) {
	var hits atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/cookbooks", func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusBadGateway)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	c, err := NewClient(Config{BaseURL: srv.URL}, WithMaxRetries(0))
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := c.Cookbooks.List(context.Background(), ListOptions{}); err == nil {
		t.Error("expected an error from 502-returning GET with retries disabled")
	}
	if got := hits.Load(); got != 1 {
		t.Errorf("hits = %d, want 1 with MaxRetries=0", got)
	}
}

// TestTransportDoesNotRetryOnContextCancel — a context cancel is
// terminal; retrying would burn cycles after the caller has already
// given up.
func TestTransportDoesNotRetryOnContextCancel(t *testing.T) {
	var hits atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/cookbooks", func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		// Wait long enough for the caller's context to fire.
		select {
		case <-r.Context().Done():
		case <-time.After(2 * time.Second):
		}
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	c := newTestClient(t, srv, false)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	if _, _, err := c.Cookbooks.List(ctx, ListOptions{}); err == nil {
		t.Error("expected a context-deadline error")
	}
	if got := hits.Load(); got != 1 {
		t.Errorf("hits = %d, want 1 (no retry on context cancel)", got)
	}
}

// TestTransportPreservesQueryStringOnTheWire — the v1.3 canonical
// request strips the query string before signing (signing.CanonicalPath
// is exercised by the signing package), but the request that hits the
// server still carries the query. This guards against accidentally
// signing/sending the unstripped path.
func TestTransportPreservesQueryStringOnTheWire(t *testing.T) {
	var (
		gotPath  string
		gotQuery string
	)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/cookbooks", func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		_, _ = io.WriteString(w, `{"total":0,"start":0,"items":[]}`)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	c := newTestClient(t, srv, false)

	if _, _, err := c.Cookbooks.List(context.Background(), ListOptions{Start: 5, Items: 2}); err != nil {
		t.Fatalf("List: %v", err)
	}
	if gotPath != "/api/v1/cookbooks" {
		t.Errorf("server saw path = %q, want /api/v1/cookbooks", gotPath)
	}
	if !strings.Contains(gotQuery, "start=5") || !strings.Contains(gotQuery, "items=2") {
		t.Errorf("server saw query = %q, want start=5 and items=2", gotQuery)
	}
}

// TestStreamReturnsErrorWithoutBody — non-2xx on the stream path closes
// the body and returns the error envelope, not an open ReadCloser.
func TestStreamReturnsErrorWithoutBody(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/universe", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = io.WriteString(w, `{"error_messages":["maintenance"]}`)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	c := newTestClient(t, srv, false)

	body, _, err := c.Universe.GetStream(context.Background())
	if err == nil {
		t.Fatal("expected an error from a 503 on the stream path")
	}
	if body != nil {
		t.Error("non-nil body returned alongside an error")
	}
}
