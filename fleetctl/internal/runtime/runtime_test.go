package runtime

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	goruntime "runtime"
	"testing"

	"github.com/mainguyen0112/fleetcontrol/api/gen"
)

func TestSaveLoadAndAuthenticatedClient(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "config.yaml")
	opts := &Options{ConfigPath: path}
	want := Config{Server: "http://example.test", Token: "secret-token"}
	if err := opts.Save(want); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := opts.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
	if goruntime.GOOS != "windows" {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat config: %v", err)
		}
		if got := info.Mode().Perm(); got != 0o600 {
			t.Fatalf("config permissions = %o, want 600", got)
		}
	}

	_, editors, err := opts.Client(true)
	if err != nil {
		t.Fatalf("Client: %v", err)
	}
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, want.Server, nil)
	if err := editors[0](context.Background(), req); err != nil {
		t.Fatalf("editor: %v", err)
	}
	if got := req.Header.Get("Authorization"); got != "Bearer secret-token" {
		t.Fatalf("Authorization = %q", got)
	}
}

func TestClientUsesDefaultTimeout(t *testing.T) {
	opts := &Options{ConfigPath: filepath.Join(t.TempDir(), "missing.yaml")}
	client, _, err := opts.Client(false)
	if err != nil {
		t.Fatalf("Client: %v", err)
	}
	generated, ok := client.ClientInterface.(*gen.Client)
	if !ok {
		t.Fatalf("generated client = %T", client.ClientInterface)
	}
	httpClient, ok := generated.Client.(*http.Client)
	if !ok {
		t.Fatalf("HTTP client = %T", generated.Client)
	}
	if httpClient.Timeout != DefaultHTTPTimeout {
		t.Fatalf("timeout = %s, want %s", httpClient.Timeout, DefaultHTTPTimeout)
	}
}

func TestClientUsesInjectedHTTPClient(t *testing.T) {
	want := &http.Client{}
	opts := &Options{ConfigPath: filepath.Join(t.TempDir(), "missing.yaml"), HTTPClient: want}
	client, _, err := opts.Client(false)
	if err != nil {
		t.Fatalf("Client: %v", err)
	}
	generated := client.ClientInterface.(*gen.Client)
	if generated.Client != want {
		t.Fatalf("HTTP client = %p, want %p", generated.Client, want)
	}
}

func TestClientRequiresLogin(t *testing.T) {
	opts := &Options{ConfigPath: filepath.Join(t.TempDir(), "missing.yaml")}
	if _, _, err := opts.Client(true); err == nil {
		t.Fatal("expected missing-token error")
	}
}
