# FleetControl CRD Design

> This document distinguishes the current Kubebuilder spike from the accepted reconciliation target. Trust, credential, and liveness ownership are authoritative in [ADR 0001](adr/0001-trust-identity-and-liveness.md).

## 1. API group and version

```text
Group:   fleetcontrol.io
Version: v1alpha1
Kind:    Satellite
Scope:   Namespaced
```

The duplicated-domain API group from the original scaffold was normalized to `fleetcontrol.io` in Phase 5.0. Because the project has not reached `v1`, the schema and reconciliation behavior may still evolve without a formal compatibility guarantee.

## 2. Current Operator spike

The repository currently provides:

- the generated Satellite CRD and RBAC manifests;
- a controller-runtime manager and reconciler;
- a minimal `spec.region` field;
- status fields for phase, ownership, heartbeat, and conditions; and
- `envtest` controller coverage.

The current reconciler only demonstrates the watch/status-update path. It marks a reconciled resource as ready and Operator-managed without contacting the Control Plane. It does not yet:

- call the generated Control Plane client;
- authenticate as an Operator principal;
- create or rotate an Agent credential Secret;
- install or process a finalizer; or
- mirror real heartbeat/liveness state.

Consequently, current CR `Ready` status is prototype behavior and must not be interpreted as proof that an Agent is alive.

## 3. Resource shape

```yaml
apiVersion: fleetcontrol.io/v1alpha1
kind: Satellite
metadata:
  name: hcm-edge
spec:
  region: hcm
status:
  phase: Ready
  managedBy: operator
  lastHeartbeat: "2026-06-21T10:00:00Z"
  conditions:
    - type: Ready
      status: "True"
      reason: HeartbeatObserved
      message: Control Plane reports a recent Agent heartbeat
      observedGeneration: 3
      lastTransitionTime: "2026-06-21T10:00:00Z"
    - type: Synced
      status: "True"
      reason: ControlPlaneInSync
      message: Desired state matches the Control Plane record
      observedGeneration: 3
      lastTransitionTime: "2026-06-21T09:59:00Z"
```

### Spec

| Field | Type | Required | Description |
|---|---|---|---|
| `spec.region` | string | Yes | Logical region/site identifier such as `hcm`, `hn`, or `dn` |

The MVP spec remains intentionally small. Git owns desired spec, and ArgoCD synchronizes it into Kubernetes.

### Status schema

| Field | Type | Intended meaning |
|---|---|---|
| `status.phase` | string | Mirror of the Control Plane runtime phase: `Pending`, `Ready`, or `Unreachable`; `Error` is reserved for future Control Plane runtime failures |
| `status.managedBy` | string | `operator` for a record reconciled from this CR |
| `status.lastHeartbeat` | RFC3339 timestamp | Mirror of the last server-recorded Agent heartbeat |
| `status.conditions` | `[]Condition` | Kubernetes-style reconciliation and observed-runtime conditions |

## 4. Target reconciliation identity

The Operator authenticates with a workload credential and reconciles a CR through an idempotent Control Plane API keyed by `metadata.uid`:

```http
GET /operator/satellites/{sourceUID}
PUT /operator/satellites/{sourceUID}
DELETE /operator/satellites/{sourceUID}
```

The Control Plane stores `source_uid` as immutable identity and retains namespace/name as descriptive metadata. Repeated reconciliation, controller restarts, and resyncs therefore cannot create duplicate database records. Deleting and recreating a CR produces a new Kubernetes UID and intentionally creates a new source identity.

Target behavior:

| Trigger | Operator action |
|---|---|
| CR created or spec changed | Idempotent `PUT`; create/update the Agent Secret as needed; set reconciliation status |
| Periodic requeue | `GET` by source UID; mirror runtime phase and last heartbeat into CR status |
| CR deleted | Idempotent `DELETE`; remove the finalizer only after the Control Plane operation succeeds |
| Control Plane unavailable | Preserve the finalizer, record reconciliation failure, and requeue with backoff |

This replaces the earlier planned POST/PATCH plus read-before-create flow.

## 5. Status ownership

| State | Authoritative owner | Operator responsibility |
|---|---|---|
| Desired region/configuration | Git and CR spec | Reconcile it to the Control Plane |
| `Synced` condition | Operator | Report whether desired state converged |
| Reconciliation failure | Operator | Record the failure and retry |
| Runtime phase | Control Plane | Mirror it; do not derive it independently. `Error` is reserved and is not used for reconciliation failures |
| Last heartbeat timestamp | Control Plane | Mirror it from the API |
| `Ready` condition | Control Plane-derived liveness | Translate the observed API phase into CR condition form |

The Agent reports heartbeat only to the Control Plane. Kubernetes does not observe PostgreSQL changes, so the MVP Operator uses periodic `RequeueAfter` polling to refresh mirrored status.

## 6. Agent credential Secret

For each reconciled Satellite, the Operator owns plaintext credential generation:

1. Generate a cryptographically random token with at least 256 bits of entropy.
2. Store it in a namespaced Kubernetes Secret consumed by the Agent.
3. Register or rotate the token through an authenticated Operator endpoint.
4. Retry safely with the same Secret value after partial failure.

The Control Plane stores only the token digest. Plaintext credentials never appear in the CR spec or status.

## 7. Finalizer target

```text
fleetcontrol.io/finalizer
```

The finalizer will prevent a CR from disappearing before its Control Plane record and credential association have been deleted. Deleting an already-absent Control Plane record returns success, making finalization retry-safe.

The finalizer is planned for Phase 8 and is not present in the current controller.

## 8. Component relationship

```text
Git -> ArgoCD -> Satellite CR
                         ^  |
                  status |  | desired state
                         |  v
                 Fleet Operator <----> Control Plane API ----> PostgreSQL
                                             ^
                                             |
                                      Satellite Agent
                                        (heartbeat)
```

The CRD never communicates directly with an Agent or database. The Operator reconciles desired state and mirrors observed state; the Control Plane owns runtime liveness.
