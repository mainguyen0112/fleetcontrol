package user

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

// fakeRepo is a minimal in-memory stand-in for Repository, used to test
// Service business logic (role validation, defaulting) without a real DB.
type fakeRepo struct {
	created *User
}

func (f *fakeRepo) Create(ctx context.Context, u *User) (*User, error) {
	f.created = u
	return u, nil
}

func (f *fakeRepo) GetByUsername(ctx context.Context, username string) (*User, error) {
	return nil, nil
}

func (f *fakeRepo) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	return nil, nil
}

func (f *fakeRepo) List(ctx context.Context) ([]*User, error) {
	return nil, nil
}

func (f *fakeRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func TestService_Create_RejectsInvalidRole(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo)

	_, err := svc.Create(context.Background(), "eve", "pw", "superadmin")

	if !errors.Is(err, ErrInvalidRole) {
		t.Errorf("expected ErrInvalidRole, got %v", err)
	}
	if repo.created != nil {
		t.Errorf("expected repo.Create not to be called on invalid role, but it was")
	}
}

func TestService_Create_AcceptsAdminRole(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo)

	created, err := svc.Create(context.Background(), "alice", "pw", "admin")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.Role != "admin" {
		t.Errorf("expected role admin, got %q", created.Role)
	}
	if created.Username != "alice" {
		t.Errorf("expected username alice, got %q", created.Username)
	}
}

func TestService_Create_AcceptsViewerRole(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo)

	created, err := svc.Create(context.Background(), "bob", "pw", "viewer")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.Role != "viewer" {
		t.Errorf("expected role viewer, got %q", created.Role)
	}
}

func TestService_Create_HashesPassword(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo)

	created, err := svc.Create(context.Background(), "carol", "plaintext-pw", "viewer")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.PasswordHash == "" {
		t.Errorf("expected PasswordHash to be set")
	}
	if created.PasswordHash == "plaintext-pw" {
		t.Errorf("PasswordHash must not equal the plaintext password")
	}
}