package supermarket

import (
	"net/url"
	"strconv"
	"strings"
)

// itoa is a tiny wrapper around strconv.Itoa so callers don't import strconv
// just to set a single query param.
func itoa(n int) string { return strconv.Itoa(n) }

// versionPath converts a dotted cookbook version (1.2.3) into the
// underscore form (1_2_3) the Supermarket API uses on its
// /versions/<v> routes. Callers may pass either form; only dotted
// versions get rewritten.
func versionPath(v string) string {
	if strings.ContainsRune(v, '.') {
		return strings.ReplaceAll(v, ".", "_")
	}
	return v
}

// withQuery appends an encoded query string to path if q is non-empty.
func withQuery(path string, q url.Values) string {
	if len(q) == 0 {
		return path
	}
	return path + "?" + q.Encode()
}
