package signing

import (
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"os"
	"strconv"
	"strings"
	"testing"
)

func testKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	data, err := os.ReadFile("../../testdata/test_key.pem")
	if err != nil {
		t.Fatal(err)
	}
	block, _ := pem.Decode(data)
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	return key
}

// joinAuth reassembles the X-Ops-Authorization-N chunks back into the
// base64 signature.
func joinAuth(h interface{ Get(string) string }) string {
	var sig strings.Builder
	for i := 1; ; i++ {
		chunk := h.Get("X-Ops-Authorization-" + strconv.Itoa(i))
		if chunk == "" {
			break
		}
		sig.WriteString(chunk)
	}
	return sig.String()
}

func TestSignHeaders_ContainsExpectedKeys(t *testing.T) {
	key := testKey(t)
	r := Request{Method: "GET", Path: "/api/v1/cookbooks", UserID: "u",
		Timestamp: "2024-01-01T00:00:00Z"}
	h, err := SignHeaders(r, key)
	if err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"X-Ops-Sign", "X-Ops-Userid", "X-Ops-Timestamp",
		"X-Ops-Content-Hash", "X-Ops-Server-Api-Version", "X-Ops-Authorization-1"} {
		if h.Get(k) == "" {
			t.Errorf("missing header %s", k)
		}
	}
}

// TestSignHeaders_SignHeaderValue pins the X-Ops-Sign header to the exact
// mixlib v1.1 string, including the trailing semicolon.
func TestSignHeaders_SignHeaderValue(t *testing.T) {
	key := testKey(t)
	h, err := SignHeaders(Request{Method: "GET", Path: "/x", UserID: "u", Timestamp: "t"}, key)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := h.Get("X-Ops-Sign"), "algorithm=sha1;version=1.1;"; got != want {
		t.Errorf("X-Ops-Sign = %q, want %q", got, want)
	}
}

// TestSignHeaders_UserIDHeaderIsRaw checks the v1.1 split-brain: the header
// carries the raw user id even though the canonical request hashes it.
func TestSignHeaders_UserIDHeaderIsRaw(t *testing.T) {
	key := testKey(t)
	h, err := SignHeaders(Request{Method: "GET", Path: "/x", UserID: "tester", Timestamp: "t"}, key)
	if err != nil {
		t.Fatal(err)
	}
	if got := h.Get("X-Ops-UserId"); got != "tester" {
		t.Errorf("X-Ops-UserId header = %q, want raw %q", got, "tester")
	}
}

// TestSignHeaders_ContentHashIsSHA1 asserts X-Ops-Content-Hash is
// base64(SHA1(body)), the v1.0/1.1 content hash.
func TestSignHeaders_ContentHashIsSHA1(t *testing.T) {
	key := testKey(t)
	body := []byte(`{"a":1}`)
	h, err := SignHeaders(Request{Method: "POST", Path: "/x", Body: body, UserID: "u", Timestamp: "t"}, key)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := h.Get("X-Ops-Content-Hash"), ContentHash(body); got != want {
		t.Errorf("X-Ops-Content-Hash = %q, want %q", got, want)
	}
	// SHA1("") would be 2jmj...; make sure we didn't hash an empty body.
	if h.Get("X-Ops-Content-Hash") == ContentHash(nil) {
		t.Errorf("content hash matches empty body; body was not hashed")
	}
}

// TestSignHeaders_SignatureVerifies is the always-on round-trip check: a
// v1.0/1.1 signature is a raw-message PKCS#1 v1.5 RSA signature over the
// canonical-request STRING, so it must verify with crypto.Hash(0).
func TestSignHeaders_SignatureVerifies(t *testing.T) {
	key := testKey(t)
	r := Request{Method: "POST", Path: "/api/v1/cookbooks", Body: []byte(`{"a":1}`),
		UserID: "u", Timestamp: "2024-01-01T00:00:00Z"}
	h, err := SignHeaders(r, key)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := base64.StdEncoding.DecodeString(joinAuth(h))
	if err != nil {
		t.Fatal(err)
	}
	if err := rsa.VerifyPKCS1v15(&key.PublicKey, crypto.Hash(0), []byte(CanonicalRequest(r)), raw); err != nil {
		t.Fatalf("signature does not verify: %v", err)
	}
}

func TestChunk60(t *testing.T) {
	got := chunk60(strings.Repeat("x", 130))
	if len(got) != 3 || len(got[0]) != 60 || len(got[2]) != 10 {
		t.Fatalf("chunk60 produced %d chunks: lens %d/%d/%d",
			len(got), len(got[0]), len(got[1]), len(got[2]))
	}
}
