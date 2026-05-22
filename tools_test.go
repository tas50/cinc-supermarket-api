package supermarket

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestToolsListReturnsPaginatedResults(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/tools", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"total":1,"start":0,"items":[{"tool_name":"berks","tool_type":"library","tool_owner":"chef","tool":"u","tool_description":"d","tool_source_url":"s"}]}`)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	c := newTestClient(t, srv, false)

	page, _, err := c.Tools.List(context.Background(), ToolsListOptions{Order: "recently_added"})
	if err != nil {
		t.Fatalf("Tools.List: %v", err)
	}
	if len(page.Items) != 1 || page.Items[0].Name != "berks" {
		t.Errorf("page = %+v", page)
	}
}

func TestToolsGetReturnsTool(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/tools/berks", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"name":"Berkshelf","slug":"berks","type":"library","instructions":"gem install berkshelf","owner":"chef"}`)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	c := newTestClient(t, srv, false)

	tool, _, err := c.Tools.Get(context.Background(), "berks")
	if err != nil {
		t.Fatalf("Tools.Get: %v", err)
	}
	if tool.Slug != "berks" || tool.Owner != "chef" {
		t.Errorf("tool = %+v", tool)
	}
}

func TestToolsSearchUsesToolsSearchPath(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/tools-search", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"total":0,"start":0,"items":[]}`)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	c := newTestClient(t, srv, false)

	if _, _, err := c.Tools.Search(context.Background(), SearchOptions{Q: "test"}); err != nil {
		t.Fatalf("Tools.Search: %v", err)
	}
}
