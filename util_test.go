package supermarket

import (
	"net/url"
	"testing"
)

func TestVersionPathConvertsDottedToUnderscore(t *testing.T) {
	cases := map[string]string{
		"1.2.3":     "1_2_3",
		"0.0.1":     "0_0_1",
		"1_2_3":     "1_2_3", // already-underscored stays put
		"latest":    "latest",
		"":          "",
		"1.2.3-pre": "1_2_3-pre",
	}
	for in, want := range cases {
		if got := versionPath(in); got != want {
			t.Errorf("versionPath(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestWithQueryReturnsBarePathWhenEmpty(t *testing.T) {
	if got := withQuery("/api/v1/cookbooks", url.Values{}); got != "/api/v1/cookbooks" {
		t.Errorf("withQuery empty = %q, want bare path", got)
	}
	if got := withQuery("/api/v1/cookbooks", nil); got != "/api/v1/cookbooks" {
		t.Errorf("withQuery nil = %q, want bare path", got)
	}
}

func TestWithQueryAppendsEncodedQuery(t *testing.T) {
	q := url.Values{}
	q.Set("q", "apache")
	q.Set("items", "10")
	got := withQuery("/api/v1/search", q)
	want1 := "/api/v1/search?items=10&q=apache"
	want2 := "/api/v1/search?q=apache&items=10"
	if got != want1 && got != want2 {
		t.Errorf("withQuery = %q, want %q (order may vary)", got, want1)
	}
}
