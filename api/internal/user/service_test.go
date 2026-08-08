package user

import (
	"context"
	"errors"
	"testing"
)

type fakeRepo struct {
	created *User
}

func (f *fakeRepo) Create(ctx context.Context, u *User) (*User, error) {
	f.created = u
	return u, nil
}
func (f *fakeRepo) GetByUsername(ctx context.Context, username string) (*User, error) { return nil, nil }
func (f *fakeRepo) GetByID(ctx context.Context, id interface{ String() string }) (*User, error) {
	return nil, nil
}
func (f *fakeRepo) List(ctx context.Context) ([]*User, error) { return nil, nil }
func (f *fakeRepo) Delete(ctx context.Context, id interface{ String() string }) error { return nil }

func TestService_Create_RejectsInvalidRole(t *testing.T) {
	svc := &Service{repo: nil} // repo won't be reached — validation short-circuits first
	_, err := svc.Create(context.Background(), "eve", "pw", "superadmin")
	if !errors.Is(err, ErrInvalidRole) {
		t.Errorf("expected ErrInvalidRole, got %v", err)
	}
}