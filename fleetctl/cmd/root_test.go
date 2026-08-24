package cmd

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoginThenHealth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/auth/login":
			fmt.Fprint(w, `{"token":"test-token","expiresIn":3600}`)
		case "/health":
			fmt.Fprint(w, `{"status":"ok","db":"ok"}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	configPath := filepath.Join(t.TempDir(), "config.yaml")
	loginOut := new(bytes.Buffer)
	login := NewRootCommand()
	login.SetOut(loginOut)
	login.SetErr(loginOut)
	login.SetArgs([]string{"--config", configPath, "--server", server.URL, "login", "--username", "admin", "--password", "admin123"})
	if err := login.Execute(); err != nil {
		t.Fatalf("login: %v", err)
	}
	if !strings.Contains(loginOut.String(), "Login successful") {
		t.Fatalf("unexpected login output: %s", loginOut.String())
	}

	healthOut := new(bytes.Buffer)
	health := NewRootCommand()
	health.SetOut(healthOut)
	health.SetErr(healthOut)
	health.SetArgs([]string{"--config", configPath, "health"})
	if err := health.Execute(); err != nil {
		t.Fatalf("health: %v", err)
	}
	if !strings.Contains(healthOut.String(), `"status": "ok"`) {
		t.Fatalf("unexpected health output: %s", healthOut.String())
	}
}

func TestRootExposesPlannedCommands(t *testing.T) {
	root := NewRootCommand()
	for _, name := range []string{"login", "health", "satellite", "user"} {
		if cmd, _, err := root.Find([]string{name}); err != nil || cmd.Name() != name {
			t.Fatalf("missing %s command", name)
		}
	}
}
