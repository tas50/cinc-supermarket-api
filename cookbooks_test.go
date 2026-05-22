package supermarket

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCookbooksListReturnsPaginatedResults(t *testing.T) {
	var gotQuery string
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/cookbooks", func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"total":42,"start":10,"items":[
			{"cookbook_name":"apache2","cookbook_description":"Web","cookbook":"u1","cookbook_maintainer":"alice"},
			{"cookbook_name":"nginx","cookbook_description":"Web","cookbook":"u2","cookbook_maintainer":"bob"}
		]}`)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	c := newTestClient(t, srv, false)

	page, _, err := c.Cookbooks.List(context.Background(), ListOptions{
		Start: 10, Items: 2, Order: "most_downloaded", User: "alice",
	})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if page.Total != 42 || page.Start != 10 || len(page.Items) != 2 {
		t.Errorf("page = %+v", page)
	}
	if page.Items[0].Name != "apache2" {
		t.Errorf("first item Name = %q", page.Items[0].Name)
	}
	if !strings.Contains(gotQuery, "start=10") ||
		!strings.Contains(gotQuery, "items=2") ||
		!strings.Contains(gotQuery, "order=most_downloaded") ||
		!strings.Contains(gotQuery, "user=alice") {
		t.Errorf("query = %q", gotQuery)
	}
	if !page.HasMore() {
		t.Error("HasMore should be true with start=10 + 2 items < total=42")
	}
}

func TestCookbooksGetReturnsFullRecord(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/cookbooks/apache2", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"name":"apache2","maintainer":"alice","latest_version":"1.2.0","metrics":{"downloads":{"total":99,"versions":{"1.2.0":99}},"followers":3}}`)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	c := newTestClient(t, srv, false)

	cb, _, err := c.Cookbooks.Get(context.Background(), "apache2")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if cb.Name != "apache2" || cb.LatestVersion != "1.2.0" || cb.Metrics.Downloads.Total != 99 {
		t.Errorf("cb = %+v", cb)
	}
}

func TestCookbooksGetMapsNotFoundFromErrorCode(t *testing.T) {
	// Supermarket returns 400 with error_code=NOT_FOUND for missing
	// resources; the library should normalize that to ErrNotFound so
	// callers can use errors.Is without thinking about HTTP codes.
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/cookbooks/missing", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, `{"error_messages":["No such cookbook"],"error_code":"NOT_FOUND"}`)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	c := newTestClient(t, srv, false)

	_, _, err := c.Cookbooks.Get(context.Background(), "missing")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("err = %v, want errors.Is(err, ErrNotFound)", err)
	}
}

func TestCookbooksGetVersionConvertsDottedToUnderscore(t *testing.T) {
	var gotPath string
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/cookbooks/apache2/versions/1_2_3", func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"version":"1.2.3","license":"Apache-2.0","cookbook":"u","file":"f"}`)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	c := newTestClient(t, srv, false)

	v, _, err := c.Cookbooks.GetVersion(context.Background(), "apache2", "1.2.3")
	if err != nil {
		t.Fatalf("GetVersion: %v", err)
	}
	if gotPath != "/api/v1/cookbooks/apache2/versions/1_2_3" {
		t.Errorf("path = %q, want underscore form", gotPath)
	}
	if v.Version != "1.2.3" {
		t.Errorf("Version = %q", v.Version)
	}
}

func TestCookbooksGetVersionAcceptsLatest(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/cookbooks/apache2/versions/latest", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"version":"2.0.0","cookbook":"u","file":"f"}`)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	c := newTestClient(t, srv, false)

	v, _, err := c.Cookbooks.GetVersion(context.Background(), "apache2", "latest")
	if err != nil {
		t.Fatalf("GetVersion latest: %v", err)
	}
	if v.Version != "2.0.0" {
		t.Errorf("Version = %q, want 2.0.0", v.Version)
	}
}

func TestCookbooksDownloadStreamsBody(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/cookbooks/apache2/versions/1_0_0/download", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/x-gzip")
		_, _ = w.Write([]byte("tarball-bytes"))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	c := newTestClient(t, srv, false)

	body, _, err := c.Cookbooks.Download(context.Background(), "apache2", "1.0.0")
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	defer body.Close()
	got, err := io.ReadAll(body)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "tarball-bytes" {
		t.Errorf("body = %q", got)
	}
}

func TestCookbooksShareRequiresCredentials(t *testing.T) {
	srv := httptest.NewServer(http.NewServeMux())
	t.Cleanup(srv.Close)
	c := newTestClient(t, srv, false)

	_, _, err := c.Cookbooks.Share(context.Background(), "apache2", "Web Servers", strings.NewReader("tar"))
	if !errors.Is(err, ErrUnauthenticatedWrite) {
		t.Errorf("err = %v, want ErrUnauthenticatedWrite", err)
	}
}

func TestCookbooksShareSendsMultipartAndIsSigned(t *testing.T) {
	var (
		gotMethod      string
		gotContentType string
		gotMetaJSON    string
		gotTarball     []byte
		gotSign        string
	)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/cookbooks", func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotContentType = r.Header.Get("Content-Type")
		gotSign = r.Header.Get("X-Ops-Sign")
		mediaType, params, err := mime.ParseMediaType(gotContentType)
		if err != nil || mediaType != "multipart/form-data" {
			t.Fatalf("Content-Type = %q (%v)", gotContentType, err)
		}
		mr := multipart.NewReader(r.Body, params["boundary"])
		for {
			part, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatal(err)
			}
			body, err := io.ReadAll(part)
			if err != nil {
				t.Fatal(err)
			}
			switch part.FormName() {
			case "cookbook":
				gotMetaJSON = string(body)
			case "tarball":
				gotTarball = body
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"name":"apache2","category":"Web Servers"}`)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	c := newTestClient(t, srv, true)

	cb, _, err := c.Cookbooks.Share(context.Background(), "apache2", "Web Servers", bytes.NewReader([]byte("FAKE-TAR")))
	if err != nil {
		t.Fatalf("Share: %v", err)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	if gotSign != "version=1.3" {
		t.Errorf("X-Ops-Sign = %q, want signed request", gotSign)
	}
	var meta map[string]string
	if err := json.Unmarshal([]byte(gotMetaJSON), &meta); err != nil {
		t.Fatalf("cookbook part not JSON: %v\nbody: %s", err, gotMetaJSON)
	}
	if meta["category"] != "Web Servers" {
		t.Errorf("cookbook part = %v", meta)
	}
	if string(gotTarball) != "FAKE-TAR" {
		t.Errorf("tarball = %q", gotTarball)
	}
	if cb.Name != "apache2" {
		t.Errorf("response cookbook = %+v", cb)
	}
}

func TestCookbooksDeleteSignsRequest(t *testing.T) {
	var sign string
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/cookbooks/apache2", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %q, want DELETE", r.Method)
		}
		sign = r.Header.Get("X-Ops-Sign")
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"name":"apache2"}`)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	c := newTestClient(t, srv, true)

	if _, _, err := c.Cookbooks.Delete(context.Background(), "apache2"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if sign != "version=1.3" {
		t.Errorf("X-Ops-Sign = %q, want signed request", sign)
	}
}

func TestCookbooksDeleteVersionRefusesWithoutCredentials(t *testing.T) {
	srv := httptest.NewServer(http.NewServeMux())
	t.Cleanup(srv.Close)
	c := newTestClient(t, srv, false)

	_, _, err := c.Cookbooks.DeleteVersion(context.Background(), "apache2", "1.0.0")
	if !errors.Is(err, ErrUnauthenticatedWrite) {
		t.Errorf("err = %v, want ErrUnauthenticatedWrite", err)
	}
}

func TestCookbooksContingentReturnsList(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/cookbooks/apache2/contingent", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `[
			{"cookbook_name":"my-stack","cookbook":"u","cookbook_description":"d","cookbook_maintainer":"alice"}
		]`)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	c := newTestClient(t, srv, false)

	got, _, err := c.Cookbooks.Contingent(context.Background(), "apache2")
	if err != nil {
		t.Fatalf("Contingent: %v", err)
	}
	if len(got) != 1 || got[0].Name != "my-stack" {
		t.Errorf("got %+v", got)
	}
}
