package supermarket

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSearchCookbooksAlwaysSendsQ(t *testing.T) {
	// The q parameter is required by the API. The library must send
	// it even when empty — otherwise the request is meaningless.
	var got string
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/search", func(w http.ResponseWriter, r *http.Request) {
		got = r.URL.RawQuery
		_, _ = io.WriteString(w, `{"total":0,"start":0,"items":[]}`)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	c := newTestClient(t, srv, false)

	if _, _, err := c.Search.Cookbooks(context.Background(), SearchOptions{}); err != nil {
		t.Fatalf("Search.Cookbooks: %v", err)
	}
	if !strings.Contains(got, "q=") {
		t.Errorf("query = %q, want q= to be present even when empty", got)
	}
}

func TestSearchCookbooksURLEncodesSpecialChars(t *testing.T) {
	var got string
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/search", func(w http.ResponseWriter, r *http.Request) {
		got = r.URL.RawQuery
		_, _ = io.WriteString(w, `{"total":0,"start":0,"items":[]}`)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	c := newTestClient(t, srv, false)

	if _, _, err := c.Search.Cookbooks(context.Background(), SearchOptions{Q: "a b&c=d"}); err != nil {
		t.Fatalf("Search.Cookbooks: %v", err)
	}
	if !strings.Contains(got, "q=a+b%26c%3Dd") {
		t.Errorf("query = %q, want characters to be URL-encoded", got)
	}
}

func TestSearchCookbooksSendsQueryAndPaging(t *testing.T) {
	var got string
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/search", func(w http.ResponseWriter, r *http.Request) {
		got = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"total":1,"start":0,"items":[{"cookbook_name":"apache2","cookbook":"u","cookbook_description":"","cookbook_maintainer":""}]}`)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	c := newTestClient(t, srv, false)

	page, _, err := c.Search.Cookbooks(context.Background(), SearchOptions{Q: "apache", Start: 0, Items: 5})
	if err != nil {
		t.Fatalf("Search.Cookbooks: %v", err)
	}
	if page.Total != 1 || page.Items[0].Name != "apache2" {
		t.Errorf("page = %+v", page)
	}
	if got != "items=5&q=apache" && got != "q=apache&items=5" {
		t.Errorf("query = %q, want q=apache and items=5 in some order", got)
	}
}
