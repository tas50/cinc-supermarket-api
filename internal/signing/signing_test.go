package signing

import "testing"

func TestContentHash(t *testing.T) {
	// base64(SHA1("")) — the empty-body content hash mixlib emits.
	got := ContentHash([]byte(""))
	want := "2jmj7l5rSw0yVb/vlWAYkK/YBwk="
	if got != want {
		t.Fatalf("ContentHash empty = %q, want %q", got, want)
	}
}

func TestCanonicalPath(t *testing.T) {
	cases := map[string]string{
		"/api/v1/cookbooks":   "/api/v1/cookbooks",
		"//api/v1//cookbooks": "/api/v1/cookbooks",
		"/api/v1/cookbooks/":  "/api/v1/cookbooks",
		"/":                   "/",
		"":                    "/",
	}
	for in, want := range cases {
		if got := CanonicalPath(in); got != want {
			t.Errorf("CanonicalPath(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestCanonicalRequest pins the exact v1.0/1.1 five-line block. The expected
// values (hashed path, content hash, hashed user id) come from the
// mixlib-authentication reference generator (see testdata/gen_vectors.rb).
func TestCanonicalRequest(t *testing.T) {
	r := Request{
		Method:    "GET",
		Path:      "/api/v1/cookbooks",
		Body:      nil,
		UserID:    "tester",
		Timestamp: "2024-01-01T00:00:00Z",
	}
	got := CanonicalRequest(r)
	want := "Method:GET\n" +
		"Hashed Path:lup0rpFkH/dwYCC2lwsseCTAje0=\n" +
		"X-Ops-Content-Hash:2jmj7l5rSw0yVb/vlWAYkK/YBwk=\n" +
		"X-Ops-Timestamp:2024-01-01T00:00:00Z\n" +
		"X-Ops-UserId:q02NKl9IChNwZ9oXEAJxzRdmB6E="
	if got != want {
		t.Fatalf("CanonicalRequest =\n%q\nwant\n%q", got, want)
	}
}

// TestCanonicalRequestHashesUserID guards the v1.1 quirk: the canonical
// request hashes the user id, it does not embed it raw (that's v1.0/1.3).
func TestCanonicalRequestHashesUserID(t *testing.T) {
	r := Request{Method: "GET", Path: "/x", UserID: "tester", Timestamp: "t"}
	canon := CanonicalRequest(r)
	if contains(canon, "X-Ops-UserId:tester") {
		t.Errorf("canonical request embeds raw user id; v1.1 must hash it:\n%s", canon)
	}
	if !contains(canon, "X-Ops-UserId:"+hashString([]byte("tester"))) {
		t.Errorf("canonical request missing hashed user id:\n%s", canon)
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
