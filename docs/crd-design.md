# FleetControl - CRD Design

## 1. Overview

The `Satellite` Custom Resource Definition (CRD) represents an edge site managed by FleetControl. It is the **declarative entry point** for the GitOps workflow: when a `Satellite` resource is created, updated, or deleted in the cluster, the Fleet Operator reconciles it against the Control Plane API (see `architecture.md` for the Source of Truth decision).

## 2. API Group & Version

```text
Group:   fleetcontrol.io
Version: v1alpha1
Kind:    Satellite
```

`v1alpha1` signals the API is still evolving — fields may change without a formal deprecation cycle until it reaches `v1`.

## 3. Full Resource Example

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
      reason: HeartbeatReceived
      lastTransitionTime: "2026-06-21T10:00:00Z"
    - type: Synced
      status: "True"
      reason: ControlPlaneInSync
      lastTransitionTime: "2026-06-21T09:59:00Z"
```

## 4. Spec Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `spec.region` | string | Yes | Logical region/site identifier (e.g. `hcm`, `hn`, `dn`) |

Intentionally minimal for the MVP — additional fields (e.g. `spec.tags`, `spec.config`) can be added in later phases without breaking existing manifests, since `v1alpha1` allows additive changes.

## 5. Status Fields (managed by the Operator — never set manually)

| Field | Type | Description |
|---|---|---|
| `status.phase` | string | `Pending` \| `Ready` \| `Error` \| `Unreachable` — mirrors the Satellite's status in the Control Plane API |
| `status.managedBy` | string | Always `operator` for resources created via this CRD |
| `status.lastHeartbeat` | string (RFC3339) | Last heartbeat timestamp received from the Satellite Agent, synced from the Control Plane API |
| `status.conditions` | []Condition | Standard Kubernetes-style conditions (see below) |

## 6. Conditions

Following the standard Kubernetes condition pattern (used by Crossplane, Cluster API, etc.):

| Condition Type | Meaning |
|---|---|
| `Ready` | The Satellite Agent has sent a recent heartbeat and the site is operational |
| `Synced` | The CRD spec matches the Control Plane API's record (no pending reconciliation) |
| `Error` | Set to `True` when the Operator fails to reach the Control Plane API (`Reason: APIUnavailable`) |

## 7. Finalizer

```text
fleetcontrol.io/finalizer
```

Added automatically on creation. Ensures that when a `Satellite` resource is deleted:
1. The Operator first calls `DELETE /satellites/{id}` on the Control Plane API.
2. Only after a successful API response does the Operator remove the finalizer, allowing Kubernetes to fully delete the CR.

This prevents orphaned records in the Control Plane database when a CRD is deleted.

## 8. Reconciliation Behavior Summary

| Trigger | Operator Action |
|---|---|
| CR created | Call `POST /satellites` with `managedBy: operator`, set `status.phase: Pending` |
| CR spec updated | Call `PATCH /satellites/{id}`, update `status.conditions[Synced]` |
| CR deleted | Call `DELETE /satellites/{id}` via finalizer, then remove finalizer |
| API unreachable | Set `status.conditions[Error] = True`, requeue with backoff (`RequeueAfter`) |
| Heartbeat received (via API polling or webhook) | Update `status.lastHeartbeat`, `status.phase: Ready` |

## 9. Relationship to Other Components

```text
Satellite CRD  ──(reconcile)──>  Fleet Operator  ──(REST call)──>  Control Plane API
                                                                          ↑
                                                                    heartbeat
                                                                          │
                                                                  Satellite Agent
```

The CRD never talks directly to the Agent — it only reflects the state that the Control Plane API already knows, keeping the Operator's responsibility limited to reconciliation (not data ownership).