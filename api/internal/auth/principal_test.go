package auth

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestNewHumanPrincipal(t *testing.T) {
	tests := []struct {
		name string
		role HumanRole
	}{
		{name: "admin", role: RoleAdmin},
		{name: "viewer", role: RoleViewer},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			principal, err := NewHumanPrincipal("user-123", tt.role)
			if err != nil {
				t.Fatalf("NewHumanPrincipal returned an error: %v", err)
			}
			if principal.Kind() != ActorHuman || principal.Subject() != "user-123" {
				t.Fatalf("unexpected human principal: %+v", principal)
			}

			role, ok := principal.HumanRole()
			if !ok || role != tt.role {
				t.Fatalf("expected role %q, got %q (present=%v)", tt.role, role, ok)
			}
			if err := principal.Validate(); err != nil {
				t.Fatalf("constructed human principal is invalid: %v", err)
			}
			if satelliteID, ok := principal.SatelliteID(); ok || satelliteID != uuid.Nil {
				t.Fatalf("human principal must not have a satellite binding: %s", satelliteID)
			}
		})
	}
}

func TestNewOperatorPrincipal(t *testing.T) {
	principal, err := NewOperatorPrincipal("fleet-operator")
	if err != nil {
		t.Fatalf("NewOperatorPrincipal returned an error: %v", err)
	}
	if principal.Kind() != ActorOperator || principal.Subject() != "fleet-operator" {
		t.Fatalf("unexpected operator principal: %+v", principal)
	}
	if role, ok := principal.HumanRole(); ok || role != "" {
		t.Fatalf("operator principal must not have a human role: %q", role)
	}
	if err := principal.Validate(); err != nil {
		t.Fatalf("constructed operator principal is invalid: %v", err)
	}
	if satelliteID, ok := principal.SatelliteID(); ok || satelliteID != uuid.Nil {
		t.Fatalf("operator principal must not have a satellite binding: %s", satelliteID)
	}
}

func TestNewAgentPrincipal(t *testing.T) {
	wantSatelliteID := uuid.New()
	principal, err := NewAgentPrincipal("agent-credential-123", wantSatelliteID)
	if err != nil {
		t.Fatalf("NewAgentPrincipal returned an error: %v", err)
	}
	if principal.Kind() != ActorAgent || principal.Subject() != "agent-credential-123" {
		t.Fatalf("unexpected agent principal: %+v", principal)
	}
	if role, ok := principal.HumanRole(); ok || role != "" {
		t.Fatalf("agent principal must not have a human role: %q", role)
	}
	if err := principal.Validate(); err != nil {
		t.Fatalf("constructed agent principal is invalid: %v", err)
	}
	if satelliteID, ok := principal.SatelliteID(); !ok || satelliteID != wantSatelliteID {
		t.Fatalf("expected satellite binding %s, got %s (present=%v)", wantSatelliteID, satelliteID, ok)
	}
}

func TestPrincipalConstructorsRejectInvalidInput(t *testing.T) {
	tests := []struct {
		name      string
		construct func() (Principal, error)
	}{
		{
			name: "empty human subject",
			construct: func() (Principal, error) {
				return NewHumanPrincipal("", RoleAdmin)
			},
		},
		{
			name: "blank human subject",
			construct: func() (Principal, error) {
				return NewHumanPrincipal("  ", RoleAdmin)
			},
		},
		{
			name: "missing human role",
			construct: func() (Principal, error) {
				return NewHumanPrincipal("user-123", "")
			},
		},
		{
			name: "unknown human role",
			construct: func() (Principal, error) {
				return NewHumanPrincipal("user-123", HumanRole("superadmin"))
			},
		},
		{
			name: "empty operator subject",
			construct: func() (Principal, error) {
				return NewOperatorPrincipal("")
			},
		},
		{
			name: "blank operator subject",
			construct: func() (Principal, error) {
				return NewOperatorPrincipal("\t")
			},
		},
		{
			name: "empty agent subject",
			construct: func() (Principal, error) {
				return NewAgentPrincipal("", uuid.New())
			},
		},
		{
			name: "blank agent subject",
			construct: func() (Principal, error) {
				return NewAgentPrincipal("\t", uuid.New())
			},
		},
		{
			name: "nil agent satellite",
			construct: func() (Principal, error) {
				return NewAgentPrincipal("agent-credential-123", uuid.Nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			principal, err := tt.construct()
			if !errors.Is(err, ErrInvalidPrincipal) {
				t.Fatalf("expected ErrInvalidPrincipal, got %v", err)
			}
			if principal != (Principal{}) {
				t.Fatalf("invalid constructor must return a zero Principal: %+v", principal)
			}
		})
	}
}

func TestPrincipalValidateRejectsMixedOrUnknownIdentity(t *testing.T) {
	satelliteID := uuid.New()

	tests := []struct {
		name      string
		principal Principal
	}{
		{name: "zero value", principal: Principal{}},
		{name: "unknown actor", principal: Principal{kind: ActorKind("service"), subject: "service-1"}},
		{name: "human without role", principal: Principal{kind: ActorHuman, subject: "user-1"}},
		{name: "human with satellite", principal: Principal{kind: ActorHuman, subject: "user-1", role: RoleAdmin, satelliteID: satelliteID}},
		{name: "operator with role", principal: Principal{kind: ActorOperator, subject: "operator-1", role: RoleAdmin}},
		{name: "operator with satellite", principal: Principal{kind: ActorOperator, subject: "operator-1", satelliteID: satelliteID}},
		{name: "agent with role", principal: Principal{kind: ActorAgent, subject: "agent-1", role: RoleViewer, satelliteID: satelliteID}},
		{name: "agent without satellite", principal: Principal{kind: ActorAgent, subject: "agent-1"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.principal.Validate(); !errors.Is(err, ErrInvalidPrincipal) {
				t.Fatalf("expected ErrInvalidPrincipal, got %v", err)
			}
		})
	}
}
