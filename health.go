package supermarket

import (
	"context"
	"net/http"
)

// Health is the small status object returned by /health (and the
// /status alias).
type Health struct {
	Status string `json:"status"`
}

// HealthService accesses the health / metrics endpoints.
type HealthService struct{ client *Client }

// Status hits /api/v1/health and returns the parsed status object.
func (s *HealthService) Status(ctx context.Context) (*Health, *Response, error) {
	h, resp, err := doJSON[Health](ctx, s.client, request{
		method: http.MethodGet,
		path:   "/api/v1/health",
	})
	if err != nil {
		return nil, resp, err
	}
	return &h, resp, nil
}

// Metrics hits /api/v1/metrics and returns the raw decoded JSON. The
// shape is server-version-specific so callers receive map[string]any.
func (s *HealthService) Metrics(ctx context.Context) (map[string]any, *Response, error) {
	return doJSON[map[string]any](ctx, s.client, request{
		method: http.MethodGet,
		path:   "/api/v1/metrics",
	})
}
