# ADR 0001: Trust, Identity, and Liveness Ownership

- **Status:** Accepted
- **Date:** 2026-08-27
- **Scope:** Control Plane API, `fleetctl`, Fleet Operator, and Satellite Agent

> Acceptance records the target architecture. It does not mean every part of this ADR is implemented yet. The delivery sequence defines which later phases deliver each invariant.

## Context

FleetControl has three kinds of caller with different trust levels:

- people using `fleetctl` or the REST API;
- the Fleet Operator reconciling Kubernetes `Satellite` resources; and
- a Satellite Agent proving that one edge site is alive.

The current implementation has only one credential model: a human JWT containing `user_id` and `role`. Any valid JWT can reach every Satellite mutation and heartbeat route, including a token for a `viewer`. User routes alone require the `admin` role.

Operator authority is currently inferred from the client-controlled header `X-FleetControl-Operator: true` on update and delete requests. Heartbeats also use the human JWT middleware and are not bound to a specific Satellite. These are transitional implementation gaps, not accepted security boundaries.

The persisted `managed_by` field is useful resource provenance, but it cannot prove who sent a request. Treating `admin`, `viewer`, `operator`, and `agent` as values in one role list would also mix human authorization with workload identity.

FleetControl therefore needs one explicit trust model before adding Operator-to-API reconciliation or a real Agent.

## Decision

### 1. Principal kinds and human roles are separate concepts

Every successfully authenticated request will carry a server-created typed principal equivalent to:

```go
type ActorKind string

const (
    ActorHuman    ActorKind = "human"
    ActorOperator ActorKind = "operator"
    ActorAgent    ActorKind = "agent"
)

type HumanRole string

const (
    RoleAdmin  HumanRole = "admin"
    RoleViewer HumanRole = "viewer"
)

type Principal struct {
    Kind        ActorKind
    Subject     string
    Role        HumanRole
    SatelliteID *uuid.UUID
}
```

The invariants are:

- `Subject` is required and identifies the authenticated principal.
- When `Kind == human`, `Role` must be exactly `admin` or `viewer` and `SatelliteID` must be nil.
- When `Kind == operator`, `Role` must be empty and `SatelliteID` must be nil.
- When `Kind == agent`, `Role` must be empty and `SatelliteID` must be non-nil, binding that credential to one Satellite.
- Operator and Agent identities are not special human roles.
- Request bodies, query parameters, and ordinary HTTP headers never create or upgrade a principal.

### 2. Authentication, authorization, and ownership checks are separate layers

The request path will be:

```text
credential
    -> actor-specific authenticator
    -> typed Principal in request context
    -> endpoint capability check
    -> domain ownership invariant
    -> handler/service operation
```

Authentication proves the caller's identity. Authorization decides whether that principal kind and role may invoke a capability. Domain logic then enforces resource-specific invariants such as preventing manual mutation of an Operator-owned Satellite.

Handlers and services must not accept trust shortcuts such as `fromOperator bool`, and `X-FleetControl-Operator` will be removed. Only a verified Operator credential may produce `ActorOperator`.

Authorization is deny-by-default. Adding a route requires an explicit decision about its permitted principal kinds and roles.

### 3. Each actor kind has an appropriate credential

#### Human

Humans continue to authenticate with username and password. Passwords remain bcrypt-hashed in PostgreSQL. Successful login returns a short-lived JWT.

Human JWT validation will require:

- an explicit algorithm allowlist;
- a strong configured signing key, with no production fallback secret;
- `iss`, `aud`, `sub`, `iat`, and `exp` claims;
- `actor_kind=human`; and
- a valid human role (`admin` or `viewer`).

The Control Plane validates all required claims before constructing a principal. Tokens issued before this claim contract is introduced may be deliberately invalidated; users can log in again.

#### Operator

The Operator uses a workload credential supplied through a Kubernetes Secret. The credential must authenticate as exactly one `ActorOperator`; neither a header nor `managed_by` may establish Operator identity. Bootstrap, storage, and rotation details will be implemented with the Operator identity work in Phase 6.

#### Agent

Each Satellite Agent receives a unique, revocable credential bound to exactly one Satellite:

- the Operator generates the plaintext token using a cryptographically secure random source with at least 256 bits of entropy;
- the Operator stores that plaintext in the Satellite's Kubernetes Secret;
- the Operator registers or rotates the credential with the Control Plane over TLS;
- the Control Plane stores only a SHA-256 digest with a unique index, never the plaintext token;
- random machine credentials are not bcrypt-hashed because deterministic lookup is required and their security comes from generated entropy rather than human memorability;
- any application-level digest comparison uses constant-time comparison;
- rotation is idempotent and invalidates the previous credential; and
- the Agent reloads its token file when sending heartbeat requests so Secret rotation does not require a restart.

Credentials and their plaintext values must never appear in logs, API responses after registration, CR specs, or CR status.

### 4. Permission matrix

| Principal | Allowed capabilities |
|---|---|
| Unauthenticated | Login, health, version, OpenAPI document, and Swagger UI |
| Human + Viewer | Read Satellite state |
| Human + Admin | Viewer capabilities, user administration, and manual Satellite lifecycle |
| Operator | Read and reconcile Operator-owned Satellites and register or rotate their Agent credentials |
| Agent | Submit heartbeat for its bound Satellite only |

Additional invariants:

- A human administrator cannot mutate an Operator-owned Satellite through manual routes.
- An Operator cannot administer human users or submit Agent heartbeats.
- An Agent cannot read fleet state or select another Satellite by changing a path parameter.
- A path containing a Satellite ID must match the authenticated Agent's `SatelliteID` before a heartbeat is accepted.

OpenAPI declares whether credentials are required. The server remains authoritative for the finer principal and resource checks in this matrix.

### 5. Resource ownership is server-derived

`managed_by` describes how a Satellite record is owned; it is not a caller identity or permission.

- Manual Satellite routes always derive `managed_by=manual` from an authorized human workflow.
- Dedicated Operator routes always derive `managed_by=operator` from an authenticated Operator principal.
- Clients cannot select or overwrite `managed_by` in request bodies or headers.
- Operator-owned records use the Kubernetes object's immutable UID as `source_uid`, with namespace and name retained as descriptive metadata.
- FleetControl will not add a speculative `source_type` field until a second real declarative source exists.

The planned reconciliation contract is idempotent:

```http
GET /operator/satellites/{sourceUID}
PUT /operator/satellites/{sourceUID}
DELETE /operator/satellites/{sourceUID}
```

- `GET` returns the current record and API-owned liveness state for status mirroring, or `404 Not Found` when no record has been materialized.
- `PUT` returns `201 Created` for the first materialization and `200 OK` for reconciliation of an existing record.
- Repeating the same `PUT` converges on the same record.
- `DELETE` returns `204 No Content` whether the record was deleted now or was already absent.

This contract prevents the Operator from implementing a separate read-before-create workflow and makes retries safe after partial network failures.

### 6. The Control Plane owns runtime liveness

Liveness has one authoritative writer:

- An authenticated Agent asserts only that its bound Satellite is alive; it does not choose `Ready`, `Unreachable`, timestamps, or another Satellite ID.
- The Control Plane records server time for the heartbeat and owns runtime phase transitions: new records begin `Pending`, a valid heartbeat produces `Ready`, and expiry of the heartbeat timeout produces `Unreachable`.
- `Error` remains reserved for a future Control Plane-detected runtime failure and is not emitted by the MVP. Operator reconciliation failures never write the runtime phase.
- The Operator owns reconciliation state such as the `Synced` condition and reconciliation errors.
- The Operator periodically fetches Control Plane state and mirrors the API-owned phase and last heartbeat into CR `Ready` and `lastHeartbeat` status fields.
- The Operator does not independently declare a Satellite ready merely because reconciliation succeeded.

For the MVP, periodic `RequeueAfter` polling is sufficient for status mirroring. Event streaming or watch delivery is outside the MVP.

### 7. OpenAPI is the HTTP contract

`api/openapi/openapi.yaml` remains the single source of truth for product API paths, request and response schemas, and authentication requirements. `/openapi.yaml` and `/docs` are explicit router-level documentation delivery routes rather than generated product operations.

Phase 5 will align runtime behavior with that contract by:

- declaring security schemes and per-operation requirements without trying to encode the entire permission matrix in OpenAPI;
- installing request validation at the HTTP boundary;
- returning a consistent JSON error envelope and media type;
- mapping typed domain errors to stable status codes;
- removing hidden privilege headers and undocumented behavior; and
- testing the same production router and generated client used at runtime.

The public routes are limited to login, health, version, the OpenAPI document, and Swagger UI. Every other route must declare and enforce an actor-specific authentication policy.

## Delivery sequence

- **Phase 5:** typed Principal, authentication layering, JWT hardening, authorization, typed errors, OpenAPI alignment, runtime validation, and security-focused integration tests.
- **Phase 6:** Operator source identity, idempotent Operator endpoints, Agent credential registration and rotation, and Control Plane liveness transitions.
- **Phase 7:** real Agent, token-file reload, and a sequential heartbeat/backoff loop with at most one request in flight.
- **Phase 8:** Operator reconciliation, Secret provisioning, finalizers, and periodic CR status mirroring.

Agent, Operator integration, and GitOps implementation must not start before the Phase 5 trust boundary is enforced and verified.

## Consequences

### Positive

- Identity cannot be escalated with a caller-controlled header or resource field.
- Human roles remain understandable and do not accumulate machine identities.
- Agent credentials can be revoked and rotated independently per Satellite.
- Reconciliation retries are idempotent.
- Runtime liveness has one authority, preventing conflicting `Ready` writers.
- Authorization tests can be expressed as a finite principal/capability matrix.

### Costs and migration impact

- Authentication middleware and tests must be reorganized around typed principals.
- JWT claim changes require existing users to log in again.
- Phase 6 requires database migrations for source identity and Agent credential digests.
- The Operator needs Secret lifecycle and credential-rotation logic.
- CR readiness can lag the Control Plane by one polling interval in the MVP.

## Rejected alternatives

- **Treat `admin`, `viewer`, `operator`, and `agent` as peer roles:** rejected because it conflates human authorization with workload identity.
- **Trust `X-FleetControl-Operator` or `managed_by`:** rejected because both are caller-controlled or resource data, not proof of identity.
- **Use one fleet-wide Agent credential:** rejected because compromise would allow one Agent to impersonate every Satellite.
- **Use bcrypt for random Agent tokens:** rejected because salted password hashes do not support indexed deterministic lookup.
- **Let the API generate the plaintext Agent token:** rejected because a lost success response makes retries ambiguous; Operator-generated input makes registration idempotent.
- **Let the Operator decide runtime readiness:** rejected because heartbeat data lives in the Control Plane and Kubernetes does not observe those changes directly.
- **Build a second declarative Apply Engine:** rejected for the MVP because Git, ArgoCD, CRDs, and the Operator are the production reconciliation path.

## References

- [RFC 7519: JSON Web Token](https://www.rfc-editor.org/rfc/rfc7519)
- [RFC 8725: JSON Web Token Best Current Practices](https://www.rfc-editor.org/rfc/rfc8725)
- [OWASP JSON Web Token Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/JSON_Web_Token_Cheat_Sheet.html)
- [Go `crypto/subtle` package](https://pkg.go.dev/crypto/subtle)
