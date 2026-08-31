package auth

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// ActorKind identifies the type of authenticated caller.
type ActorKind string

const (
	ActorHuman    ActorKind = "human"
	ActorOperator ActorKind = "operator"
	ActorAgent    ActorKind = "agent"
)

// HumanRole identifies a human caller's authorization role.
// Machine principals never have a HumanRole.
type HumanRole string

const (
	RoleAdmin  HumanRole = "admin"
	RoleViewer HumanRole = "viewer"
)

// ErrInvalidPrincipal is returned when a principal violates an identity
// invariant defined by ADR 0001.
var ErrInvalidPrincipal = errors.New("invalid principal")

// Principal is the server-created identity produced after authentication.
// Its fields are private so callers must use an actor-specific constructor.
type Principal struct {
	kind        ActorKind
	subject     string
	role        HumanRole
	satelliteID uuid.UUID
}

// NewHumanPrincipal constructs a human principal with an admin or viewer role.
func NewHumanPrincipal(subject string, role HumanRole) (Principal, error) {
	return newPrincipal(Principal{
		kind:    ActorHuman,
		subject: subject,
		role:    role,
	})
}

// NewOperatorPrincipal constructs a workload principal for the Fleet Operator.
func NewOperatorPrincipal(subject string) (Principal, error) {
	return newPrincipal(Principal{
		kind:    ActorOperator,
		subject: subject,
	})
}

// NewAgentPrincipal constructs an Agent principal bound to one Satellite.
func NewAgentPrincipal(subject string, satelliteID uuid.UUID) (Principal, error) {
	return newPrincipal(Principal{
		kind:        ActorAgent,
		subject:     subject,
		satelliteID: satelliteID,
	})
}

func newPrincipal(principal Principal) (Principal, error) {
	if err := principal.Validate(); err != nil {
		return Principal{}, err
	}
	return principal, nil
}

// Kind returns the authenticated caller's actor kind.
func (p Principal) Kind() ActorKind {
	return p.kind
}

// Subject returns the stable identifier asserted by the verified credential.
func (p Principal) Subject() string {
	return p.subject
}

// HumanRole returns the role and true for a human principal.
// Machine principals return the empty role and false.
func (p Principal) HumanRole() (HumanRole, bool) {
	if p.kind != ActorHuman {
		return "", false
	}
	return p.role, true
}

// SatelliteID returns the Satellite binding and true for an Agent
// principal. Other principal kinds return uuid.Nil and false.
func (p Principal) SatelliteID() (uuid.UUID, bool) {
	if p.kind != ActorAgent || p.satelliteID == uuid.Nil {
		return uuid.Nil, false
	}
	return p.satelliteID, true
}

// Validate checks every cross-field Principal invariant from ADR 0001.
func (p Principal) Validate() error {
	if strings.TrimSpace(p.subject) == "" {
		return invalidPrincipal("subject is required")
	}

	switch p.kind {
	case ActorHuman:
		if p.role != RoleAdmin && p.role != RoleViewer {
			return invalidPrincipal("human role must be admin or viewer")
		}
		if p.satelliteID != uuid.Nil {
			return invalidPrincipal("human principal cannot bind a satellite")
		}
	case ActorOperator:
		if p.role != "" {
			return invalidPrincipal("operator principal cannot have a human role")
		}
		if p.satelliteID != uuid.Nil {
			return invalidPrincipal("operator principal cannot bind a satellite")
		}
	case ActorAgent:
		if p.role != "" {
			return invalidPrincipal("agent principal cannot have a human role")
		}
		if p.satelliteID == uuid.Nil {
			return invalidPrincipal("agent principal requires a satellite binding")
		}
	default:
		return invalidPrincipal("actor kind is not recognized")
	}

	return nil
}

func invalidPrincipal(reason string) error {
	return fmt.Errorf("%w: %s", ErrInvalidPrincipal, reason)
}
