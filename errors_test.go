package supermarket

import (
	"errors"
	"strings"
	"testing"
)

func TestErrorResponseUnwrapsByErrorCode(t *testing.T) {
	body := []byte(`{"error_messages":["that's missing"],"error_code":"NOT_FOUND"}`)
	er := newErrorResponse("GET", "/api/v1/cookbooks/missing", 400, body)
	if !errors.Is(er, ErrNotFound) {
		t.Errorf("error_code NOT_FOUND should map to ErrNotFound; got %v", er)
	}
	if er.Code != "NOT_FOUND" {
		t.Errorf("Code = %q, want NOT_FOUND", er.Code)
	}
	if len(er.Messages) != 1 || er.Messages[0] != "that's missing" {
		t.Errorf("Messages = %v", er.Messages)
	}
}

func TestErrorResponseUnwrapsByStatusWhenNoCode(t *testing.T) {
	er := newErrorResponse("DELETE", "/api/v1/cookbooks/foo", 403, []byte(`{}`))
	if !errors.Is(er, ErrForbidden) {
		t.Errorf("403 should map to ErrForbidden; got %v", er)
	}
}

func TestErrorResponseFallsBackToRawBody(t *testing.T) {
	er := newErrorResponse("GET", "/x", 500, []byte("plain text"))
	if len(er.Messages) != 1 || er.Messages[0] != "plain text" {
		t.Errorf("Messages = %v, want [plain text]", er.Messages)
	}
}

func TestErrorResponseUnwrapsAllSupermarketCodes(t *testing.T) {
	cases := []struct {
		code string
		want error
	}{
		{"NOT_FOUND", ErrNotFound},
		{"UNAUTHORIZED", ErrUnauthorized},
		{"FORBIDDEN", ErrForbidden},
		{"INVALID_DATA", ErrInvalidData},
	}
	for _, tc := range cases {
		body := []byte(`{"error_messages":["x"],"error_code":"` + tc.code + `"}`)
		er := newErrorResponse("GET", "/x", 400, body)
		if !errors.Is(er, tc.want) {
			t.Errorf("error_code %q -> %v, want %v", tc.code, er, tc.want)
		}
	}
}

func TestErrorResponseUnwrapStatusCodeFallbacks(t *testing.T) {
	cases := []struct {
		status int
		want   error
	}{
		{401, ErrUnauthorized},
		{403, ErrForbidden},
		{404, ErrNotFound},
		{500, nil},
	}
	for _, tc := range cases {
		er := newErrorResponse("GET", "/x", tc.status, []byte(`{}`))
		if got := er.Unwrap(); got != tc.want {
			t.Errorf("status %d -> %v, want %v", tc.status, got, tc.want)
		}
	}
}

func TestErrorResponseErrorStringFormat(t *testing.T) {
	withCode := newErrorResponse("GET", "/x", 400, []byte(`{"error_messages":["nope"],"error_code":"NOT_FOUND"}`))
	if got := withCode.Error(); !strings.Contains(got, "NOT_FOUND") || !strings.Contains(got, "nope") {
		t.Errorf("Error() = %q, want code + message in output", got)
	}
	withoutCode := newErrorResponse("GET", "/x", 500, []byte("boom"))
	if got := withoutCode.Error(); strings.Contains(got, "NOT_FOUND") || !strings.Contains(got, "boom") {
		t.Errorf("Error() without code = %q", got)
	}
	noMessages := newErrorResponse("GET", "/x", 500, nil)
	if got := noMessages.Error(); !strings.Contains(got, "(no message)") {
		t.Errorf("Error() with empty body = %q, want sentinel '(no message)'", got)
	}
}

func TestErrorResponseHandlesMalformedJSON(t *testing.T) {
	er := newErrorResponse("GET", "/x", 400, []byte("{this is not json"))
	if len(er.Messages) != 1 {
		t.Errorf("Messages = %v, want a single fallback message", er.Messages)
	}
	if er.Code != "" {
		t.Errorf("Code = %q, want empty when body is malformed", er.Code)
	}
}
