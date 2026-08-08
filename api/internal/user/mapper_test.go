package user

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/mainguyen0112/fleetcontrol/api/gen"
)

func TestCreateParamsFromRequest_DefaultsRoleToViewer(t *testing.T) {
	req := gen.CreateUserRequest{Username: "alice", Password: "secret"}
	_, _, role := CreateParamsFromRequest(req)
	if role != "viewer" {
		t.Errorf("expected default role viewer, got %q", role)
	}
}

func TestCreateParamsFromRequest_PreservesExplicitRole(t *testing.T) {
	admin := gen.CreateUserRequestRole("admin")
	req := gen.CreateUserRequest{Username: "bob", Password: "secret", Role: &admin}
	username, password, role := CreateParamsFromRequest(req)
	if role != "admin" || username != "bob" || password != "secret" {
		t.Errorf("unexpected params: %q %q %q", username, password, role)
	}
}

func TestToResponse_DoesNotLeakPasswordHash(t *testing.T) {
	u := &User{
		ID:           uuid.New(),
		Username:     "carol",
		PasswordHash: "should-never-appear",
		Role:         "viewer",
		CreatedAt:    time.Now(),
	}
	resp := ToResponse(u)
	if resp.Username == nil || *resp.Username != "carol" {
		t.Errorf("username not mapped correctly")
	}
	// gen.User has no PasswordHash field — compile-time guarantee, this test
	// documents the invariant rather than asserting a runtime absence.
}

func TestToResponse_PanicsOnNil(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic on nil User, got none")
		}
	}()
	ToResponse(nil)
}