package signing

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"strconv"
	"testing"
)

// referenceVector is one entry of testdata/vectors.json, produced by the
// real mixlib-authentication gem (testdata/gen_vectors.rb). It is the gold
// standard: if the Go signer's output diverges from these bytes, the public
// Chef Supermarket would reject the request with a 401.
type referenceVector struct {
	Name        string            `json:"name"`
	Method      string            `json:"method"`
	Path        string            `json:"path"`
	Body        string            `json:"body"` // base64
	UserID      string            `json:"user_id"`
	Timestamp   string            `json:"timestamp"`
	ContentHash string            `json:"content_hash"`
	Canonical   string            `json:"canonical"`
	Headers     map[string]string `json:"headers"`
}

func loadVectors(t *testing.T) []referenceVector {
	t.Helper()
	data, err := os.ReadFile("testdata/vectors.json")
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Vectors []referenceVector `json:"vectors"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	if len(doc.Vectors) == 0 {
		t.Fatal("no reference vectors loaded")
	}
	return doc.Vectors
}

// TestMatchesMixlibReference proves byte-for-byte agreement with the real
// mixlib-authentication v1.1 (SHA-1) implementation: the canonical request,
// content hash, X-Ops-Sign value, and — because RSA PKCS#1 v1.5 signing is
// deterministic — every X-Ops-Authorization-N chunk.
func TestMatchesMixlibReference(t *testing.T) {
	key := testKey(t)
	for _, v := range loadVectors(t) {
		t.Run(v.Name, func(t *testing.T) {
			body, err := base64.StdEncoding.DecodeString(v.Body)
			if err != nil {
				t.Fatal(err)
			}
			r := Request{
				Method:    v.Method,
				Path:      v.Path,
				Body:      body,
				UserID:    v.UserID,
				Timestamp: v.Timestamp,
			}

			if got := CanonicalRequest(r); got != v.Canonical {
				t.Errorf("canonical request mismatch:\ngot:  %q\nwant: %q", got, v.Canonical)
			}
			if got := ContentHash(body); got != v.ContentHash {
				t.Errorf("content hash = %q, want %q", got, v.ContentHash)
			}

			h, err := SignHeaders(r, key)
			if err != nil {
				t.Fatal(err)
			}
			if got, want := h.Get("X-Ops-Sign"), v.Headers["X-Ops-Sign"]; got != want {
				t.Errorf("X-Ops-Sign = %q, want %q", got, want)
			}
			if got, want := h.Get("X-Ops-Content-Hash"), v.Headers["X-Ops-Content-Hash"]; got != want {
				t.Errorf("X-Ops-Content-Hash = %q, want %q", got, want)
			}
			if got, want := h.Get("X-Ops-Userid"), v.Headers["X-Ops-Userid"]; got != want {
				t.Errorf("X-Ops-Userid = %q, want %q", got, want)
			}
			if got, want := h.Get("X-Ops-Timestamp"), v.Headers["X-Ops-Timestamp"]; got != want {
				t.Errorf("X-Ops-Timestamp = %q, want %q", got, want)
			}
			// Deterministic signature: every authorization chunk must match.
			for i := 1; ; i++ {
				key := "X-Ops-Authorization-" + strconv.Itoa(i)
				want, ok := v.Headers[key]
				if !ok {
					if got := h.Get(key); got != "" {
						t.Errorf("extra %s = %q (mixlib has none)", key, got)
					}
					break
				}
				if got := h.Get(key); got != want {
					t.Errorf("%s = %q, want %q", key, got, want)
				}
			}
		})
	}
}
