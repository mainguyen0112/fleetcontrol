# FleetControl - Architecture

## 1. High-Level Architecture

```text
Developer / Platform Engineer
            ↓
      Git Repository (fleet-configs)
            ↓
          ArgoCD
            ↓
      Fleet Operator (CRD watcher)
            ↓
     Control Plane API  ←──── fleetctl (dev/debug only)
            ↓
        PostgreSQL
            ↓
     Satellite Agent (runs at each edge site, heartbeats back to the API)
```

## 2. Source of Truth (key decision)

- **GitOps/CRD is the official path** for creating/updating Satellites in a production-like environment.
- `fleetctl satellite create/update/delete` is intended for **dev/test/debug only**, not for managing the real fleet.
- Each Satellite has a `managedBy` field:
  - `operator` — created via CRD/GitOps
  - `manual` — created directly via the CLI
- If the CLI attempts to modify a Satellite with `managedBy: operator`, the system will warn or reject the operation.

## 3. Domain Model

```go
type Satellite struct {
    ID         string
    Name       string
    Region     string
    Status     string // Pending, Ready, Error, Unreachable
    ManagedBy  string // "operator" | "manual"
    LastSeenAt *time.Time
}

type User struct {
    ID       string
    Username string
    Role     string // admin | viewer
}
```

## 4. Core Components

| Component | Role | Tech Stack |
|---|---|---|
| Control Plane API | Stores state, exposes REST API | Go + PostgreSQL |
| Fleet CLI (fleetctl) | Manual/dev interaction with the API | Go + Cobra |
| Fleet Operator | Reconciles CRD ↔ API | Kubebuilder + controller-runtime |
| Satellite Agent | A real process running at the edge site, sends heartbeats | Go |

## 5. Open Questions Answered (Phase 0 self-check)

- Who is allowed to modify a Satellite? → The Operator (via GitOps) is the official path; the CLI is for dev/test only.
- What happens if the CLI and Git both try to modify the same Satellite? → `managedBy` prevents conflicts; the CLI warns when attempting to edit a resource managed by the Operator.