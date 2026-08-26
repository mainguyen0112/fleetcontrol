package user

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mainguyen0112/fleetcontrol/fleetctl/internal/runtime"
	"github.com/mainguyen0112/fleetcontrol/fleetctl/internal/secret"
)

func TestCreateUser(t *testing.T) {
	var request struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/users" {
			t.Errorf("request = %s %s", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer admin-token" {
			t.Errorf("Authorization = %q", got)
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Errorf("decode request: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{"id":"00000000-0000-0000-0000-000000000001","username":"alice","role":"viewer","created_at":"2026-08-27T00:00:00Z"}`)
	}))
	defer server.Close()

	opts := authenticatedOptions(t, server.URL)
	cmd := NewCommand(opts, secret.NewPasswordReader())
	out := new(bytes.Buffer)
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetIn(strings.NewReader("new-user-secret\n"))
	cmd.SetArgs([]string{"create", "--username", "alice", "--role", "viewer", "--password-stdin"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("create user: %v", err)
	}

	if request.Username != "alice" || request.Password != "new-user-secret" || request.Role != "viewer" {
		t.Fatalf("request = %+v", request)
	}
	if !strings.Contains(out.String(), `"username": "alice"`) {
		t.Fatalf("output = %s", out.String())
	}
}

func TestListUsers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/users" {
			t.Errorf("request = %s %s", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer admin-token" {
			t.Errorf("Authorization = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `[{"id":"00000000-0000-0000-0000-000000000001","username":"alice","role":"admin","created_at":"2026-08-27T00:00:00Z"}]`)
	}))
	defer server.Close()

	cmd := NewCommand(authenticatedOptions(t, server.URL), secret.NewPasswordReader())
	out := new(bytes.Buffer)
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"list"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("list users: %v", err)
	}
	if !strings.Contains(out.String(), `"role": "admin"`) {
		t.Fatalf("output = %s", out.String())
	}
}

func TestCreateUserRejectsInvalidRoleAndPasswordFlag(t *testing.T) {
	opts := &runtime.Options{ConfigPath: filepath.Join(t.TempDir(), "missing.yaml")}
	cmd := NewCommand(opts, secret.NewPasswordReader())
	create, _, err := cmd.Find([]string{"create"})
	if err != nil {
		t.Fatalf("find create: %v", err)
	}
	if create.Flags().Lookup("password") != nil {
		t.Fatal("user create must not accept passwords through a command-line flag")
	}
	if create.Flags().Lookup("password-stdin") == nil {
		t.Fatal("user create must expose --password-stdin for automation")
	}

	cmd.SetArgs([]string{"create", "--username", "alice", "--role", "owner", "--password-stdin"})
	cmd.SetIn(strings.NewReader("secret\n"))
	if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), "role must be admin or viewer") {
		t.Fatalf("expected role validation error, got %v", err)
	}
}

func authenticatedOptions(t *testing.T, serverURL string) *runtime.Options {
	t.Helper()
	opts := &runtime.Options{ConfigPath: filepath.Join(t.TempDir(), "config.yaml")}
	if err := opts.Save(runtime.Config{Server: serverURL, Token: "admin-token"}); err != nil {
		t.Fatalf("save config: %v", err)
	}
	return opts
}
