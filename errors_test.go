package supermarket

import (
	"errors"
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
