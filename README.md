# FleetControl

> GitOps-based Edge Fleet Management Platform — built with Go, Kubernetes Operators, PostgreSQL, and a real heartbeat-reporting Agent.

[Architecture](docs/architecture.md) · [API Design](docs/api-design.md) · [CRD Design](docs/crd-design.md)

**Status:** 🚧 Active development — Phase 2 of 10 (see [Progress](#progress) below)

---

## Why FleetControl?

Managing a fleet of edge devices ("satellites") by hand doesn't scale: configuration drift, no clear ownership, no visibility into which nodes are actually alive.

FleetControl is a control plane that lets you manage edge satellites **declaratively**, the same way Kubernetes manages Pods:

- A Kubernetes **Operator** watches `Satellite` custom resources and reconciles them against a backend API.
- A **Satellite Agent** — a real, independently running process — registers itself and sends heartbeats back to the control plane.
- **GitOps** (via ArgoCD) becomes the official way to change fleet state in production; an imperative CLI (`fleetctl`) exists only for local dev/debug.

The detail that matters most here: a `Satellite` is not just a row in Postgres. It's backed by an actual running Agent process that can go offline, and the system can detect and report that. This is what separates FleetControl from "a CRUD app wearing a Kubernetes costume."

## Features

- [x] Custom Resource Definition (`Satellite`) + Kubernetes Operator (Kubebuilder / controller-runtime)
- [x] Control Plane REST API (Go) backed by PostgreSQL
- [ ] JWT authentication (admin / viewer roles) — in progress
- [ ] Satellite heartbeat monitoring with liveness detection (`Ready` → `Unreachable`)
- [ ] OpenAPI-first contract + generated client
- [ ] `fleetctl` CLI (imperative, dev/debug only)
- [ ] Declarative apply engine (`fleetctl apply -f fleet.yaml`)
- [ ] Real Satellite Agent process (heartbeat every 30s)
- [ ] Full GitOps loop via ArgoCD
- [ ] Observability: Prometheus metrics, structured logging

## Architecture

```text
Developer / Platform Engineer
            ↓
      Git Repository
            ↓
          ArgoCD
            ↓
      Fleet Operator (CRD watcher)
            ↓
     Control Plane API  ←──── fleetctl (dev/debug only)
            ↓
        PostgreSQL
            ↓
     Satellite Agent (runs at each edge site, heartbeats to API)
```

### Source of Truth

A core design decision made before any code was written (see [`docs/architecture.md`](docs/architecture.md)):

- **GitOps/CRD is the official path** to create or modify Satellites in any real environment.
- **`fleetctl satellite create/update/delete` is imperative and dev/debug-only** — never used against production fleet state.
- Any Satellite created via CRD is marked `managedBy: operator` in the Control Plane, so manual edits can be flagged or rejected and don't silently cause drift.

## Components

| Component | Description | Status |
|---|---|---|
| `api/` | Control Plane REST API | ✅ MVP done |
| `operator/` | Kubernetes Operator (CRD reconciliation) | ✅ Spike done |
| `agent/` | Edge-node heartbeat agent | ⏳ Phase 6.5 |
| `fleetctl/` | CLI tool | ⏳ Phase 4 |
| PostgreSQL | Metadata storage | ✅ |

## Tech Stack

- Go
- Kubernetes, Kubebuilder, controller-runtime
- PostgreSQL
- JWT (auth)
- Docker / Docker Compose
- ArgoCD (GitOps)
- OpenAPI / `oapi-codegen`
- Zap (structured logging)
- GitHub Actions, Trivy, Gosec (planned — Phase 10)

## Project Structure

```text
fleetcontrol/
├── api/          # Control Plane REST API
├── operator/     # Kubernetes Operator
├── agent/        # Satellite Agent (heartbeat process) — upcoming
├── fleetctl/     # CLI tool — upcoming
├── docs/
│   ├── architecture.md
│   ├── api-design.md
│   └── crd-design.md
└── go.work
```

## Example Workflow (target end state)

1. Platform engineer commits a `Satellite` resource to `fleet-configs/`.
2. ArgoCD syncs the resource into the cluster.
3. The Fleet Operator reconciles it, creating the Satellite in the Control Plane API (`managedBy: operator`).
4. A Satellite Agent at the edge site registers itself and starts sending heartbeats every 30s.
5. If the Agent goes silent, the Control Plane marks the Satellite `Unreachable` — visible immediately via `fleetctl satellite list` or `kubectl get satellite`.

## Getting Started

```bash
# Control Plane API + PostgreSQL
docker compose up

# Operator (local kind cluster)
kind create cluster
make manifests && make install
make run
```

> CLI, Agent, and full GitOps setup are not available yet — see [Progress](#progress).

## Progress

Tracking against the full project plan (10 phases + Agent + GitOps):

- [x] **Phase 0** — Architecture, API design, CRD design, Source-of-Truth decision
- [x] **Phase 1** — Operator spike (PoC): CRD, Reconcile loop, Status
- [ ] **Phase 2** — Control Plane API + JWT auth + tests *(in progress)*
- [ ] **Phase 3** — OpenAPI-first contract + generated client
- [ ] **Phase 4** — `fleetctl` CLI
- [ ] **Phase 5** — Declarative apply engine
- [ ] **Phase 6** — Operator MVP (full lifecycle: create/update/delete/retry/finalizer)
- [ ] **Phase 6.5** — Satellite Agent (real heartbeat process)
- [ ] **Phase 7** — Fleet Operator ↔ API integration
- [ ] **Phase 8** — Full GitOps integration (ArgoCD)
- [ ] **Phase 9** — Observability (metrics, structured logging)
- [ ] **Phase 10** — Productionization (CI/CD, security scanning, release)

## Lessons Learned (so far)

- Kubernetes reconciliation patterns and the role of `status` vs `spec`
- Designing a `Satellite ↔ Agent` relationship modeled after Crossplane's Composite/Managed Resource pattern
- Deciding *upfront* who owns writes to a resource (CLI vs GitOps) to avoid drift — easier to design than to retrofit

## Future Work

- Multi-cluster support
- Web UI / dashboard
- Fine-grained RBAC
- Multi-region scheduling

## License

MIT