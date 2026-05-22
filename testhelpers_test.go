package supermarket

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http/httptest"
	"testing"
)

// testRSAKey generates a small RSA key suitable for tests that need
// signed requests. 2048-bit so signatures are realistic.
func testRSAKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	return key
}

// newTestClient builds a Client pointed at srv. When signed is true the
// client carries credentials and can hit write endpoints.
func newTestClient(t *testing.T, srv *httptest.Server, signed bool) *Client {
	t.Helper()
	cfg := Config{BaseURL: srv.URL}
	if signed {
		cfg.Username = "tester"
		cfg.Key = testRSAKey(t)
	}
	c, err := NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	return c
}
