package supermarket

import (
	"context"
	"net/http"
	"net/url"
)

// SearchOptions configures a search query. Q is required; Start and
// Items follow the same convention as ListOptions.
type SearchOptions struct {
	Q     string
	Start int
	Items int
}

// SearchService accesses the fuzzy-search endpoints.
type SearchService struct{ client *Client }

// Cookbooks runs a fuzzy search over cookbook name, description, and
// maintainer.
func (s *SearchService) Cookbooks(ctx context.Context, opts SearchOptions) (Page[CookbookSummary], *Response, error) {
	q := url.Values{}
	q.Set("q", opts.Q)
	applyPageQuery(q, opts.Start, opts.Items)
	return doJSON[Page[CookbookSummary]](ctx, s.client, request{
		method: http.MethodGet,
		path:   withQuery("/api/v1/search", q),
	})
}
