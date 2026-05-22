package supermarket

import "net/http"

// Response wraps the raw HTTP response returned by an API call. The
// body has already been consumed and closed by the time the caller
// sees the Response.
type Response struct {
	HTTPResponse *http.Response
	StatusCode   int
}
