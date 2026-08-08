package user

import (
	"github.com/mainguyen0112/fleetcontrol/api/gen"
)

// defaultRole is used when the client does not specify a role on creation.
const defaultRole = "viewer"

// CreateParamsFromRequest translates a generated CreateUserRequest into the
// primitive parameters Service.Create expects. This is not a pure 1:1 mapper —
// it also applies the API-level default role ("viewer") when the client omits it.
func CreateParamsFromRequest(req gen.CreateUserRequest) (username, password, role string) {
	role = defaultRole
	if req.Role != nil {
		role = string(*req.Role)
	}
	return req.Username, req.Password, role
}

// ToResponse converts a domain User into the generated API model.
// Callers must guarantee u != nil — a nil User means "not found", which
// callers are expected to translate into 404 before calling ToResponse.
// This function panics on nil rather than silently returning an empty User,
// so a caller bug surfaces immediately instead of leaking as `{}` in a response.
//
// PasswordHash is intentionally never mapped — gen.User has no such field.
func ToResponse(u *User) gen.User {
	if u == nil {
		panic("user: ToResponse called with nil User — caller must check for not-found before mapping")
	}

	id := u.ID
	username := u.Username
	role := gen.UserRole(u.Role)
	createdAt := u.CreatedAt

	return gen.User{
		Id:        &id,
		Username:  &username,
		Role:      &role,
		CreatedAt: &createdAt,
	}
}

// ToResponseList converts a slice of domain Users into generated API models.
func ToResponseList(users []*User) []gen.User {
	resp := make([]gen.User, 0, len(users))
	for _, u := range users {
		resp = append(resp, ToResponse(u))
	}
	return resp
}