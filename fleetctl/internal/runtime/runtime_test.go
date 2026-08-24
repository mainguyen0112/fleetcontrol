package runtime

import (
	"context"
	"net/http"
	"path/filepath"
	"testing"
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

func TestClientRequiresLogin(t *testing.T) {
	opts := &Options{ConfigPath: filepath.Join(t.TempDir(), "missing.yaml")}
	if _, _, err := opts.Client(true); err == nil {
		t.Fatal("expected missing-token error")
	}
}
