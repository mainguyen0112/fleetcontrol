package secret

import (
	"bytes"
	"strings"
	"testing"
)

type descriptorReader struct {
	*strings.Reader
}

func (descriptorReader) Fd() uintptr { return 42 }

func TestPasswordReaderFromStdin(t *testing.T) {
	reader := NewPasswordReader()
	password, err := reader.Read(strings.NewReader("secret value\r\nignored"), new(bytes.Buffer), true)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if password != "secret value" {
		t.Fatalf("password = %q", password)
	}
}

func TestPasswordReaderPromptsOnTerminal(t *testing.T) {
	reader := PasswordReader{
		isTerminal: func(fd int) bool { return fd == 42 },
		readTerminal: func(fd int) ([]byte, error) {
			if fd != 42 {
				t.Fatalf("fd = %d", fd)
			}
			return []byte("hidden-secret"), nil
		},
	}
	out := new(bytes.Buffer)
	password, err := reader.Read(descriptorReader{strings.NewReader("")}, out, false)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if password != "hidden-secret" {
		t.Fatalf("password = %q", password)
	}
	if out.String() != "Password: \n" {
		t.Fatalf("prompt = %q", out.String())
	}
}

func TestPasswordReaderRejectsNonTerminalAndEmptyInput(t *testing.T) {
	reader := NewPasswordReader()
	if _, err := reader.Read(strings.NewReader("secret\n"), new(bytes.Buffer), false); err == nil || !strings.Contains(err.Error(), "--password-stdin") {
		t.Fatalf("expected non-terminal guidance, got %v", err)
	}
	if _, err := reader.Read(strings.NewReader("\n"), new(bytes.Buffer), true); err == nil || !strings.Contains(err.Error(), "must not be empty") {
		t.Fatalf("expected empty-password error, got %v", err)
	}
}
