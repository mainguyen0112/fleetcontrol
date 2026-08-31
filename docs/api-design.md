# FleetControl API Design

> `api/openapi/openapi.yaml` is the authoritative product HTTP contract. This document explains the current runtime and the accepted security direction. `/openapi.yaml` and `/docs` are explicit router-level documentation delivery routes rather than generated product operations. Trust and identity rules are authoritative in [ADR 0001](adr/0001-trust-identity-and-liveness.md).

## 1. Current runtime contract

The Phase 3 server uses the generated `ServerInterface`, and production and integration tests share the same adapter and router.

### Public routes

- `POST /auth/login`
- `GET /health`
- `GET /version`

The router also exposes these unauthenticated documentation routes outside the generated `ServerInterface`:

- `GET /openapi.yaml`
- `GET /docs`

### Human-JWT routes

| Routes | Current enforcement |
|---|---|
| `GET /satellites`, `GET /satellites/{id}` | Any valid human JWT |
| `POST /satellites`, `PATCH /satellites/{id}`, `DELETE /satellites/{id}` | Any valid human JWT; actor-specific authorization is not implemented yet |
| `POST /satellites/{id}/heartbeat` | Any valid human JWT; per-Satellite Agent authentication is not implemented yet |
| `GET /users`, `POST /users`, `DELETE /users/{id}` | Human JWT with `role=admin` |

There is currently no refresh route, Satellite filtering contract, dedicated status route, authenticated Operator creation route, or background transition to `Unreachable`.

### Known trust-boundary gaps

- `X-FleetControl-Operator: true` currently bypasses the Operator-owned update/delete guard. It is caller-controlled and will be removed.
- Every valid human JWT can currently mutate Satellites and submit any Satellite's heartbeat.
- JWT validation does not yet enforce the complete issuer, audience, subject, algorithm, and actor-kind contract.
- Request decoding is not yet backed by OpenAPI runtime validation.
- Some middleware and handlers do not yet return the documented JSON media type and status mapping consistently.

These gaps are explicit Phase 5 work, not supported integration behavior.

## 2. Accepted target boundary

Authentication converts an actor-specific credential into a typed Principal. Authorization then applies this baseline matrix:

| Principal | API capability |
|---|---|
| Human + Viewer | Read Satellite state |
| Human + Admin | Viewer capabilities, user administration, and manual Satellite lifecycle |
| Operator | Idempotent Operator-owned Satellite reconciliation and Agent credential registration/rotation |
| Agent | Heartbeat for the credential's bound Satellite only |

OpenAPI records whether a route requires credentials. The service remains authoritative for principal-kind, human-role, and resource-ownership checks that OpenAPI cannot express.

## 3. Satellite ownership

The REST representation uses the field name `managed_by`; the Kubernetes CR status uses `managedBy`.

`managed_by` is response/resource provenance and is assigned by the server:

- authorized human creation produces `manual`;
- authenticated Operator reconciliation produces `operator`; and
- no request body or ordinary header can select or overwrite it.

Manual routes reject mutation of Operator-owned resources even for a human administrator.

The future Operator contract is keyed by immutable Kubernetes source UID:

```http
GET /operator/satellites/{sourceUID}
PUT /operator/satellites/{sourceUID}
DELETE /operator/satellites/{sourceUID}
```

- `GET` returns the current record and API-owned liveness state for CR status mirroring, or `404` if it has not been materialized.
- `PUT` returns `201` when created and `200` when reconciled or updated.
- Repeated `PUT` requests converge on the same record.
- `DELETE` returns `204` when deleted or already absent.

These routes are Phase 6 work and are not in the current OpenAPI document.

## 4. Heartbeat and liveness

The target heartbeat request is an authenticated assertion of presence, not a client-selected status update:

- the Agent credential is bound to one Satellite;
- the path Satellite ID must match that binding;
- the Control Plane writes `last_seen_at` using server time;
- new records begin `Pending`, and the Control Plane derives `Ready` and `Unreachable` from heartbeat state;
- `Error` is reserved for a future Control Plane-detected runtime failure and is not an Operator reconciliation status; and
- the Operator later mirrors that API-owned state into CR status.

The MVP heartbeat loop and liveness timeout are delivered in Phases 6 and 7. There is no implemented 90-second background worker today.

## 5. Error and validation contract

All API errors converge on this JSON envelope:

```json
{
  "error": {
    "code": "SATELLITE_MANAGED_BY_OPERATOR",
    "message": "This satellite is managed by GitOps and cannot be edited manually."
  }
}
```

Phase 5 will make status mapping consistent:

| Status | Meaning |
|---|---|
| `400` | Malformed or contract-invalid input |
| `401` | Missing, invalid, or expired credential |
| `403` | Authenticated principal lacks the required capability or ownership permission |
| `404` | Requested resource does not exist |
| `409` | Uniqueness or state conflict |
| `500` | Unexpected internal error; storage details are not exposed |

Runtime request validation, typed domain errors, and generated-client integration tests must verify these responses before Phase 5 is complete.
