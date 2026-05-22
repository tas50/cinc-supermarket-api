package supermarket

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUniverseGetDecodesNestedMap(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/universe", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"apache2": {
				"1.0.0": {"location_path":"u1","location_type":"supermarket","dependencies":{"openssl":">= 0.0.0"}},
				"2.0.0": {"location_path":"u2","location_type":"supermarket","dependencies":{}}
			}
		}`)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	c := newTestClient(t, srv, false)

	uni, _, err := c.Universe.Get(context.Background())
	if err != nil {
		t.Fatalf("Universe.Get: %v", err)
	}
	apache := uni["apache2"]
	if apache["1.0.0"].Dependencies["openssl"] != ">= 0.0.0" {
		t.Errorf("uni = %+v", uni)
	}
	if apache["2.0.0"].LocationPath != "u2" {
		t.Errorf("uni[apache2][2.0.0] = %+v", apache["2.0.0"])
	}
}

func TestUniverseGetHandlesEmptyGraph(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/universe", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{}`)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	c := newTestClient(t, srv, false)

	uni, _, err := c.Universe.Get(context.Background())
	if err != nil {
		t.Fatalf("Universe.Get: %v", err)
	}
	if len(uni) != 0 {
		t.Errorf("expected empty universe, got %v", uni)
	}
}

func TestUniverseGetStreamLetsCallerDecodeIncrementally(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/universe", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"apache2":{"1.0.0":{"location_path":"u","location_type":"supermarket","dependencies":{}}}}`)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	c := newTestClient(t, srv, false)

	body, _, err := c.Universe.GetStream(context.Background())
	if err != nil {
		t.Fatalf("Universe.GetStream: %v", err)
	}
	defer body.Close()
	var uni Universe
	if err := json.NewDecoder(body).Decode(&uni); err != nil {
		t.Fatal(err)
	}
	if _, ok := uni["apache2"]["1.0.0"]; !ok {
		t.Errorf("decoded universe = %+v", uni)
	}
}
