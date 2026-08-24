package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

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
	message := strings.TrimSpace(string(body))
	if message == "" {
		message = "empty response"
	}
	return fmt.Errorf("API returned HTTP %d: %s", status, message)
}
