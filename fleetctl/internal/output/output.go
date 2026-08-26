package output

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type APIError struct {
	StatusCode int
	Code       string
	Message    string
}

func (e *APIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("API error %s (HTTP %d): %s", e.Code, e.StatusCode, e.Message)
	}
	return fmt.Sprintf("API request failed (HTTP %d): %s", e.StatusCode, e.Message)
}

func JSON(w io.Writer, value any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(value)
}

func StatusError(status int, body []byte, expected ...int) error {
	for _, want := range expected {
		if status == want {
			return nil
		}
	}
	var envelope struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &envelope); err == nil && envelope.Error.Message != "" {
		return &APIError{StatusCode: status, Code: envelope.Error.Code, Message: envelope.Error.Message}
	}
	message := http.StatusText(status)
	if message == "" {
		message = "unexpected response"
	}
	return &APIError{StatusCode: status, Message: message}
}
