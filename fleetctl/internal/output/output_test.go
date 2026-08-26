package output

import (
	"errors"
	"testing"
)

func TestStatusErrorParsesStructuredAPIError(t *testing.T) {
	tests := []struct {
		status int
		code   string
	}{
		{status: 401, code: "UNAUTHORIZED"},
		{status: 403, code: "FORBIDDEN"},
		{status: 404, code: "NOT_FOUND"},
		{status: 409, code: "CONFLICT"},
	}
	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			body := []byte(`{"error":{"code":"` + tt.code + `","message":"friendly message"}}`)
			err := StatusError(tt.status, body, 200)
			var apiErr *APIError
			if !errors.As(err, &apiErr) {
				t.Fatalf("error = %T %v", err, err)
			}
			if apiErr.StatusCode != tt.status || apiErr.Code != tt.code || apiErr.Message != "friendly message" {
				t.Fatalf("APIError = %+v", apiErr)
			}
		})
	}
}

func TestStatusErrorDoesNotExposeUnstructuredResponseBody(t *testing.T) {
	err := StatusError(502, []byte("proxy trace containing internal details"), 200)
	want := "API request failed (HTTP 502): Bad Gateway"
	if err == nil || err.Error() != want {
		t.Fatalf("error = %v, want %q", err, want)
	}
}

func TestStatusErrorAcceptsExpectedStatus(t *testing.T) {
	if err := StatusError(204, nil, 200, 204); err != nil {
		t.Fatalf("StatusError: %v", err)
	}
}
