# FleetControl

> GitOps-based Edge Fleet Management Platform built with Go, Kubernetes Operators, PostgreSQL, and a real heartbeat-reporting Agent.

[Architecture](docs/architecture.md) · [API Design](docs/api-design.md) · [CRD Design](docs/crd-design.md)

**Status:** 🚧 Active development — **Phase 3 of 10**

Phase 2 is complete: the Control Plane API now includes JWT authentication, RBAC, CRUD APIs, heartbeat reporting, health endpoints, structured logging, and integration tests.

---

# Why FleetControl?

Managing hundreds or thousands of edge devices manually doesn't scale. Configuration drift, unclear ownership, and lack of visibility quickly become operational problems.

FleetControl is designed as a Kubernetes-inspired control plane that manages edge satellites declaratively instead of imperatively.

The system follows the same philosophy Kubernetes uses for Pods:

- A Kubernetes **Operator** watches `Satellite` Custom Resources and continuously reconciles them against the backend Control Plane.
- A **Satellite Agent** runs independently on each edge node and periodically reports its health through heartbeat requests.
- **GitOps** (via ArgoCD) becomes the single source of truth for production fleet state.
- An imperative CLI (`fleetctl`) exists only for development and debugging.

Most importantly, a Satellite is **not** just a database row.

Every Satellite represents a real running process that can become unreachable, recover, or drift away from the desired state. FleetControl continuously detects and reports those situations rather than simply storing metadata.

That is what makes FleetControl an infrastructure platform instead of a CRUD application.

---

# Features

- ✅ Kubernetes Custom Resource (`Satellite`)
- ✅ Kubernetes Operator foundation (Kubebuilder)
- ✅ Control Plane REST API (Go)
- ✅ PostgreSQL persistence
- ✅ JWT Authentication
- ✅ Role-Based Access Control (Admin / Viewer)
- ✅ Satellite CRUD
- ✅ User CRUD
- ✅ Business rule enforcement (`managed_by`)
- ✅ Satellite heartbeat endpoint
- ✅ Health & Version endpoints
- ✅ Structured JSON logging (Zap)
- ✅ Integration tests covering authentication and CRUD flow
- ⏳ OpenAPI-first contract generation
- ⏳ `fleetctl` CLI
- ⏳ Declarative Apply Engine
- ⏳ Real Satellite Agent
- ⏳ Full GitOps integration
- ⏳ Prometheus metrics

---

# Architecture

```text
               Developer / Platform Engineer
                           │
                           ▼
                    Git Repository
                           │
                           ▼
                        ArgoCD
                           │
                           ▼
              Fleet Operator (Reconciliation)
                           │
                           ▼
                 Control Plane REST API
                           ▲
                           │
        fleetctl (Development / Debugging Only)
                           │
                           ▼
                     PostgreSQL Database
                           ▲
                           │
              Satellite Agent (Heartbeat)
```

---

# Source of Truth

One of the earliest architectural decisions made in this project is defining **who owns fleet state**.

FleetControl intentionally separates declarative and imperative workflows.

- **GitOps + Kubernetes CRDs** are the official way to manage production Satellites.
- **fleetctl** is intended only for local development, testing, and debugging.
- Satellites created by the Operator are marked as `managed_by = operator`.
- Manual updates to Operator-managed Satellites are rejected by the Control Plane API to prevent configuration drift.

Making this ownership decision before implementation greatly simplifies later development.

---

# Design Principles

FleetControl intentionally follows several engineering principles.

### Layered Architecture

```
HTTP Request
    │
    ▼
Router
    │
    ▼
Middleware
    │
    ▼
Handler
    │
    ▼
Service
    │
    ▼
Repository
    │
    ▼
PostgreSQL
```

Each layer has exactly one responsibility.

| Layer | Responsibility |
|--------|----------------|
| Router | Route incoming requests |
| Middleware | Authentication, authorization, logging |
| Handler | HTTP parsing and response generation |
| Service | Business logic |
| Repository | Database access |
| PostgreSQL | Data persistence |

---

### Dependency Injection

Objects are created in `main.go` using constructor injection.

```
main
 ├── Repository
 ├── Service(Repository)
 └── Handler(Service)
```

Business logic depends on interfaces instead of concrete implementations, making the code easier to test and extend.

---

### Separation of Concerns

FleetControl deliberately separates:

- HTTP handling
- Business rules
- Database access
- Infrastructure

No SQL appears inside handlers.

No HTTP details appear inside repositories.

Business rules stay inside the Service layer.

---

### GitOps First

The project treats Git as the source of truth.

Instead of issuing imperative commands against production systems, desired state is declared in Git and continuously reconciled by the Kubernetes Operator.

---

### Stateless API

Authentication is implemented with JWT.

The server stores no session state, allowing multiple API instances to be deployed behind a load balancer without session synchronization.

---

# Components

| Component | Description | Status |
|-----------|-------------|--------|
| `api/` | Control Plane REST API | ✅ Phase 2 Complete |
| `operator/` | Kubernetes Operator | ✅ Phase 1 Complete |
| `agent/` | Edge Heartbeat Agent | ⏳ Phase 6.5 |
| `fleetctl/` | CLI Tool | ⏳ Phase 4 |
| PostgreSQL | Metadata Storage | ✅ |

---

# Technology Stack

| Technology | Why it was chosen |
|------------|-------------------|
| **Go** | Native language of the Kubernetes ecosystem. Produces static binaries, excellent concurrency model, and strong standard library. |
| **Chi Router** | Lightweight, idiomatic router built on `net/http`. Easier to understand and test than heavier frameworks while remaining highly extensible. |
| **PostgreSQL** | Reliable relational database with excellent consistency, indexing, and mature tooling. Fleet metadata naturally fits relational modeling. |
| **pgx** | Native PostgreSQL driver offering better PostgreSQL support and performance than generic `database/sql` drivers. |
| **JWT** | Stateless authentication suitable for APIs and future Agent communication without server-side sessions. |
| **bcrypt** | Industry-standard password hashing algorithm resistant to brute-force attacks. Passwords are never stored in plaintext. |
| **Zap** | High-performance structured JSON logging designed for production services and centralized log aggregation. |
| **Docker Compose** | Simple reproducible local development environment for API and PostgreSQL. |
| **Kubebuilder + controller-runtime** | Standard toolkit for building Kubernetes Operators using reconciliation patterns. |
| **OpenAPI + oapi-codegen** *(Phase 3)* | Contract-first API development with generated server/client types to eliminate documentation drift. |
| **ArgoCD** *(Planned)* | Implements GitOps by continuously reconciling Git with the Kubernetes cluster. |

---

# Project Structure

```text
fleetcontrol/
├── api/              # Control Plane REST API
├── operator/         # Kubernetes Operator
├── agent/            # Edge Agent (upcoming)
├── fleetctl/         # CLI (upcoming)
├── deploy/           # Helm & ArgoCD manifests
├── observability/    # Prometheus & Grafana
├── docs/
│   ├── architecture.md
│   ├── api-design.md
│   └── crd-design.md
└── go.work
```

---

# Example Workflow (Target End State)

1. A Platform Engineer commits a new `Satellite` resource into Git.
2. ArgoCD synchronizes the resource into the Kubernetes cluster.
3. The Fleet Operator reconciles the CRD against the Control Plane API.
4. The Satellite is created with `managed_by = operator`.
5. A real Satellite Agent registers itself and begins sending heartbeat requests.
6. If the Agent stops reporting, the Control Plane automatically marks it as `Unreachable`.
7. Fleet status can be viewed through either `kubectl` or `fleetctl`.

---

# Getting Started

```bash
# Start PostgreSQL
docker compose up -d postgres

# Start API
cd api
go run cmd/server/main.go

# Start local Kubernetes cluster
kind create cluster

# Install CRD
cd operator
make manifests
make install

# Run Operator
make run
```

The CLI, Agent, and GitOps integration are still under development.

---

# Progress

- ✅ **Phase 0** — Architecture, API Design, CRD Design
- ✅ **Phase 1** — Kubernetes Operator Spike
- ✅ **Phase 2** — Control Plane API
  - JWT Authentication
  - RBAC
  - CRUD APIs
  - Heartbeat endpoint
  - Health endpoint
  - Structured logging
  - Integration tests
- ⏳ **Phase 3** — OpenAPI-first contract
- ⏳ **Phase 4** — fleetctl CLI
- ⏳ **Phase 5** — Declarative Apply Engine
- ⏳ **Phase 6** — Full Operator lifecycle
- ⏳ **Phase 6.5** — Real Satellite Agent
- ⏳ **Phase 7** — Operator ↔ API integration
- ⏳ **Phase 8** — GitOps with ArgoCD
- ⏳ **Phase 9** — Observability
- ⏳ **Phase 10** — Productionization

---

# Lessons Learned

## Architecture

- Designing software around layers (Handler → Service → Repository) results in cleaner code and easier maintenance than organizing around endpoints.
- Constructor-based Dependency Injection keeps dependencies explicit and greatly improves testability.
- Depending on interfaces instead of concrete implementations follows the Dependency Inversion Principle and enables mocking during testing.

## Authentication & Authorization

- Authentication (JWT) and authorization (RBAC) solve different problems and should remain separate.
- Stateless JWT authentication enables horizontal scaling without server-side session storage.
- Passwords should never be stored directly; bcrypt provides adaptive hashing that significantly increases resistance to brute-force attacks.

## Business Logic

- Business rules belong in the Service layer, not in SQL queries or HTTP handlers.
- Rules such as `managed_by = operator` prevent configuration drift between GitOps-managed resources and manual API operations.
- Default values (`Pending`, `manual`) are business decisions rather than database concerns.

## Persistence

- Database migrations make schema evolution reproducible, version-controlled, and safe across environments.
- The Repository pattern isolates persistence logic from business logic, allowing future database changes with minimal impact.

## API Design

- HTTP handlers should translate HTTP requests into domain objects and nothing more.
- Integration tests are essential because they validate the complete request lifecycle, including middleware, routing, business logic, and database interactions.

## Observability

- Structured JSON logging is significantly more useful than plain text logs because it can be indexed and queried by centralized logging systems.
- Health endpoints should verify dependencies (such as database connectivity) instead of merely confirming that the process is alive.

## Project Design

- Defining ownership (GitOps vs CLI) before implementation prevents configuration drift later.
- FleetControl is intentionally developed as an application platform. A future companion project, **CloudBase**, will provide the production infrastructure that deploys and operates FleetControl, mirroring the separation between application teams and platform teams in real-world organizations.

---

# Future Work

- Multi-cluster support
- Multi-region scheduling
- Web dashboard
- Fine-grained RBAC
- Agent auto-registration
- Event streaming
- Horizontal API deployment
- Cloud-native production deployment on AWS

---

# License

MIT