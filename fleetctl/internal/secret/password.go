package secret

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"

	"golang.org/x/term"
)

type fileDescriptor interface {
	Fd() uintptr
}

// PasswordReader reads secrets from a terminal without echoing them, or from
// standard input when an automation-friendly flow is explicitly requested.
type PasswordReader struct {
	isTerminal   func(int) bool
	readTerminal func(int) ([]byte, error)
}

func NewPasswordReader() PasswordReader {
	return PasswordReader{
		isTerminal:   term.IsTerminal,
		readTerminal: term.ReadPassword,
	}
}

func (r PasswordReader) Read(in io.Reader, out io.Writer, fromStdin bool) (string, error) {
	if fromStdin {
		password, err := bufio.NewReader(in).ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return "", fmt.Errorf("read password from stdin: %w", err)
		}
		return validate(strings.TrimRight(password, "\r\n"))
	}

	input, ok := in.(fileDescriptor)
	if !ok || r.isTerminal == nil || !r.isTerminal(int(input.Fd())) {
		return "", fmt.Errorf("password input is not a terminal; use --password-stdin")
	}
	if r.readTerminal == nil {
		return "", fmt.Errorf("terminal password reader is unavailable")
	}

	if _, err := fmt.Fprint(out, "Password: "); err != nil {
		return "", fmt.Errorf("write password prompt: %w", err)
	}
	password, err := r.readTerminal(int(input.Fd()))
	if _, writeErr := fmt.Fprintln(out); writeErr != nil && err == nil {
		return "", fmt.Errorf("finish password prompt: %w", writeErr)
	}
	if err != nil {
		return "", fmt.Errorf("read password: %w", err)
	}
	return validate(string(password))
}

func validate(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("password must not be empty")
	}
	return password, nil
}
