package signing

import "testing"

func TestContentHash(t *testing.T) {
	got := ContentHash([]byte(""))
	want := "47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU="
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

func TestCanonicalRequest(t *testing.T) {
	r := Request{
		Method:    "GET",
		Path:      "/api/v1/cookbooks/apache2",
		Body:      nil,
		UserID:    "me",
		Timestamp: "2024-01-01T00:00:00Z",
	}
	got := CanonicalRequest(r)
	want := "Method:GET\n" +
		"Path:/api/v1/cookbooks/apache2\n" +
		"X-Ops-Content-Hash:47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=\n" +
		"X-Ops-Sign:version=1.3\n" +
		"X-Ops-Timestamp:2024-01-01T00:00:00Z\n" +
		"X-Ops-UserId:me\n" +
		"X-Ops-Server-API-Version:1"
	if got != want {
		t.Fatalf("CanonicalRequest =\n%q\nwant\n%q", got, want)
	}
}
