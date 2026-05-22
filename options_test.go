package supermarket

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestWithHTTPClientUsesGivenClient(t *testing.T) {
	custom := &http.Client{Timeout: 7 * time.Second}
	c, err := NewClient(Config{BaseURL: "https://x.test"}, WithHTTPClient(custom))
	if err != nil {
		t.Fatal(err)
	}
	if c.httpClient != custom {
		t.Error("WithHTTPClient did not install the provided *http.Client")
	}
}

func TestWithHTTPClientIgnoresNil(t *testing.T) {
	c, err := NewClient(Config{BaseURL: "https://x.test"}, WithHTTPClient(nil))
	if err != nil {
		t.Fatal(err)
	}
	if c.httpClient == nil {
		t.Error("nil *http.Client should be silently ignored, not installed")
	}
}

func TestWithMaxRetriesIgnoresNegative(t *testing.T) {
	c, err := NewClient(Config{BaseURL: "https://x.test"}, WithMaxRetries(-3))
	if err != nil {
		t.Fatal(err)
	}
	if c.opts.maxRetries == -3 {
		t.Errorf("negative MaxRetries should be ignored, got %d", c.opts.maxRetries)
	}
}

func TestWithUserAgentOverridesDefault(t *testing.T) {
	var got string
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("User-Agent")
		_, _ = io.WriteString(w, `{"status":"ok"}`)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	c, err := NewClient(Config{BaseURL: srv.URL}, WithUserAgent("my-tool/1.0"))
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := c.Health.Status(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got != "my-tool/1.0" {
		t.Errorf("User-Agent = %q, want my-tool/1.0", got)
	}
}

func TestWithSkipTLSVerifyEnablesInsecureTransport(t *testing.T) {
	c, err := NewClient(Config{BaseURL: "https://x.test"}, WithSkipTLSVerify(true))
	if err != nil {
		t.Fatal(err)
	}
	tr, ok := c.httpClient.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("transport = %T, want *http.Transport", c.httpClient.Transport)
	}
	if tr.TLSClientConfig == nil || !tr.TLSClientConfig.InsecureSkipVerify {
		t.Errorf("InsecureSkipVerify = false, want true")
	}
}
