# FleetControl - API Design

## 1. Design Principles

- The API follows an **OpenAPI-first** approach (Phase 3) — the content below is a hand-written draft that will later become the authoritative `openapi.yaml`.
- All routes except `/health`, `/version`, and `/auth/login` require a valid JWT.
- The `managedBy` field determines edit permissions: a client attempting to modify a Satellite with `managedBy: operator` through a manual route will be rejected (403) unless the request comes from the Operator's service account.

## 2. Auth

### POST /auth/login
Request:
```json
{ "username": "admin", "password": "..." }
```
Response:
```json
{ "token": "eyJ...", "expiresIn": 3600 }
```

### POST /auth/refresh
Refreshes the token before it expires.

## 3. Satellite

### POST /satellites
Creates a new Satellite. Defaults to `managedBy: manual` when called via CLI/direct API; the Operator calls this with a dedicated header to set `managedBy: operator`.

Request:
```json
{ "name": "hcm-edge", "region": "hcm" }
```
Response: `201 Created`
```json
{
  "id": "uuid",
  "name": "hcm-edge",
  "region": "hcm",
  "status": "Pending",
  "managedBy": "manual",
  "lastSeenAt": null
}
```

### GET /satellites
Lists all Satellites. Supports query params `?region=` and `?status=` for filtering.

### GET /satellites/{id}
Retrieves details of a single Satellite.

### PATCH /satellites/{id}
Updates fields (region, etc.). Returns `403` if `managedBy: operator` and the request lacks the Operator's service-account header.

### DELETE /satellites/{id}
Deletes a Satellite. Same `managedBy` rule as PATCH.

### POST /satellites/{id}/heartbeat
Called periodically by the Agent (every 30s — see Phase 6.5).

Request:
```json
{ "status": "Ready" }
```
Effect: updates `lastSeenAt = now()`, `status = Ready`.
If no heartbeat is received for > 90s → a background job sets `status = Unreachable`.

### GET /satellites/{id}/status
Returns current status + lastSeenAt — used by `fleetctl health` or a dashboard.

## 4. User

### POST /users
Admin only.
```json
{ "username": "viewer1", "password": "...", "role": "viewer" }
```

### GET /users
Lists users (admin only).

### DELETE /users/{id}
Admin only.

## 5. Health & Version

### GET /health
```json
{ "status": "ok", "db": "connected" }
```
No auth required — used for liveness probes.

### GET /version
```json
{ "version": "v0.1.0", "commit": "abc1234" }
```

## 6. Standard Error Format (applies to the entire API)

```json
{
  "error": {
    "code": "SATELLITE_MANAGED_BY_OPERATOR",
    "message": "This satellite is managed by GitOps and cannot be edited manually."
  }
}
```

## 7. Status Codes Used Consistently

| Code | Case |
|---|---|
| 200 | Successful GET/PATCH |
| 201 | Successful creation via POST |
| 204 | Successful DELETE |
| 400 | Validation error |
| 401 | Missing/invalid token |
| 403 | Valid permission but blocked by `managedBy` rule or role |
| 404 | Resource not found |
| 409 | Conflict (e.g. duplicate Satellite name) |