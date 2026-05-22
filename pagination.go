package supermarket

import "net/url"

// Page is the envelope every Supermarket list/search response uses.
// Items carries the actual result rows; Start and Total describe where
// this slice sits inside the full result set.
type Page[T any] struct {
	Start int `json:"start"`
	Total int `json:"total"`
	Items []T `json:"items"`
}

// HasMore reports whether the server has more results past this page.
func (p Page[T]) HasMore() bool {
	return p.Start+len(p.Items) < p.Total
}

// NextStart returns the Start offset a caller should pass to fetch the
// page that immediately follows p. Returns 0 when HasMore is false.
func (p Page[T]) NextStart() int {
	if !p.HasMore() {
		return 0
	}
	return p.Start + len(p.Items)
}

// applyPageQuery sets start/items query params when non-zero.
func applyPageQuery(q url.Values, start, items int) {
	if start > 0 {
		q.Set("start", itoa(start))
	}
	if items > 0 {
		q.Set("items", itoa(items))
	}
}
