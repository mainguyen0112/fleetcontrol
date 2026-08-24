package satellite

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/mainguyen0112/fleetcontrol/fleetctl/internal/runtime"
)

func TestDeleteRejectsOperatorManagedSatellite(t *testing.T) {
	deleteCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			deleteCalled = true
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"id":"`+uuid.Nil.String()+`","name":"edge","region":"r1","status":"Ready","managed_by":"operator","created_at":"2026-08-24T00:00:00Z","last_seen_at":null}`)
	}))
	defer server.Close()

	opts := &runtime.Options{ConfigPath: filepath.Join(t.TempDir(), "config.yaml")}
	if err := opts.Save(runtime.Config{Server: server.URL, Token: "token"}); err != nil {
		t.Fatal(err)
	}
	cmd := NewCommand(opts)
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetArgs([]string{"delete", uuid.Nil.String()})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "managed by the Operator") {
		t.Fatalf("unexpected error: %v", err)
	}
	if deleteCalled {
		t.Fatal("DELETE request must not be sent for Operator-managed satellite")
	}
}
