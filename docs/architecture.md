# FleetControl Architecture

> This document separates the code that exists today from the accepted target architecture. Trust and identity decisions are authoritative in [ADR 0001](adr/0001-trust-identity-and-liveness.md).

## 1. Current state

The repository currently contains:

- a Control Plane REST API backed by PostgreSQL;
- a Phase 4 `fleetctl` MVP that calls the API through the generated OpenAPI client;
- a Kubebuilder Operator foundation with the `fleetcontrol.io/v1alpha1` Satellite CRD;
- an OpenAPI-generated server/client boundary with unit and PostgreSQL integration tests; and
- CI for generated-code drift, API tests, CLI tests, and Operator `envtest` tests.

The Operator does not yet call the Control Plane, the Agent is not implemented, and ArgoCD is not wired into an end-to-end flow. The current API authenticates human JWTs, but actor-specific authorization and workload credentials are Phase 5 and Phase 6 work.

```text
fleetctl ──generated client──> Control Plane API ──> PostgreSQL

Fleet Operator prototype ──> Satellite CR status only

Satellite Agent and end-to-end GitOps flow: planned
```

## 2. Target architecture

```text
 Developer / Platform Engineer
             │
             ▼
      Git Repository ──> ArgoCD ──> Satellite CR
                                         │
                                         ▼
                               Fleet Operator
                               (generated client)
                                         │
                                         ▼
fleetctl (dev/debug) ─────────> Control Plane API <──────── Satellite Agent
 (generated client)                 │                         (heartbeat)
                                    ▼
                               PostgreSQL
```

Only the Control Plane API accesses PostgreSQL. `fleetctl`, the Operator, and the Agent are API clients with distinct authenticated principals and permissions.

## 3. Source of truth and ownership

| Concern | Authoritative owner |
|---|---|
| Production desired Satellite state | Git, synchronized to Kubernetes by ArgoCD |
| Declarative reconciliation | Fleet Operator |
| Manual development/debug lifecycle | Human administrator through API or `fleetctl` |
| Resource provenance (`managed_by`) | Control Plane, derived from the authenticated route/principal |
| Runtime heartbeat timestamp and liveness phase | Control Plane |
| CR `Synced` and reconciliation-error state | Fleet Operator |
| CR `Ready` and `lastHeartbeat` | Fleet Operator mirror of Control Plane state |

`managed_by` has two values:

- `operator` for a record materialized through an authenticated Operator workflow; and
- `manual` for a record created through an authorized human workflow.

It is resource provenance, not caller identity. Clients cannot select it, and manual routes reject mutation of Operator-owned records.

## 4. Trust boundary

FleetControl distinguishes three actor kinds:

- `human`, with a separate `admin` or `viewer` role;
- `operator`, representing the reconciliation workload; and
- `agent`, bound to one Satellite.

Authentication converts a verified actor-specific credential into a typed Principal. Authorization checks that principal against an explicit capability, and the domain layer applies ownership rules. Headers and request bodies do not establish workload identity.

See [ADR 0001](adr/0001-trust-identity-and-liveness.md) for the permission matrix, credential lifecycle, idempotent Operator API, and liveness ownership decision.

## 5. Core components

| Component | Responsibility | Status |
|---|---|---|
| Control Plane API | HTTP contract, authorization, fleet metadata, and runtime liveness | Implemented foundation; Phase 5 hardening in progress |
| PostgreSQL | Durable users, Satellite state, and future credential digests | Implemented foundation |
| `fleetctl` | Manual development and debugging workflows | Phase 4 MVP complete |
| Fleet Operator | Reconcile Satellite CR desired state with the Control Plane | Foundation only; integration planned for Phase 8 |
| Satellite Agent | Authenticate as one Satellite and report heartbeat | Planned for Phase 7 |
| ArgoCD | Synchronize Git desired state into Kubernetes | Planned for Phase 9 |

FleetControl deliberately has no second production Apply Engine. Git, ArgoCD, CRDs, and the Operator form the declarative reconciliation path.
