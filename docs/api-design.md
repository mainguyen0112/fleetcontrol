# FleetControl - API Design

## 1. Nguyên tắc thiết kế

- API được thiết kế theo hướng **OpenAPI-first** (Phase 3) — nội dung dưới đây là bản nháp tay trước, sẽ chuyển thành `openapi.yaml` chính thức sau.
- Tất cả route trừ `/health`, `/version`, `/auth/login` đều yêu cầu JWT hợp lệ.
- Field `managedBy` quyết định ai có quyền sửa: client cố sửa Satellite có `managedBy: operator` qua route thủ công sẽ bị từ chối (403) trừ khi request đến từ Operator's service account.

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
Làm mới token khi gần hết hạn.

## 3. Satellite

### POST /satellites
Tạo Satellite mới. Mặc định `managedBy: manual` nếu gọi qua CLI/API trực tiếp; Operator gọi với header riêng để set `managedBy: operator`.

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
List toàn bộ Satellite. Hỗ trợ query param `?region=` và `?status=` để filter.

### GET /satellites/{id}
Lấy chi tiết 1 Satellite.

### PATCH /satellites/{id}
Cập nhật field (region, v.v.). Trả `403` nếu `managedBy: operator` và request không có header service-account của Operator.

### DELETE /satellites/{id}
Xóa Satellite. Cùng rule `managedBy` như PATCH.

### POST /satellites/{id}/heartbeat
Agent gọi định kỳ (mỗi 30s — xem Phase 6.5).

Request:
```json
{ "status": "Ready" }
```
Hiệu ứng: cập nhật `lastSeenAt = now()`, `status = Ready`.
Nếu không nhận heartbeat trong > 90s → background job set `status = Unreachable`.

### GET /satellites/{id}/status
Trả về status + lastSeenAt hiện tại — dùng cho `fleetctl health` hoặc dashboard.

## 4. User

### POST /users
Chỉ `admin` được gọi.
```json
{ "username": "viewer1", "password": "...", "role": "viewer" }
```

### GET /users
List user (admin only).

### DELETE /users/{id}
Admin only.

## 5. Health & Version

### GET /health
```json
{ "status": "ok", "db": "connected" }
```
Không cần auth — dùng cho liveness probe.

### GET /version
```json
{ "version": "v0.1.0", "commit": "abc1234" }
```

## 6. Error format chuẩn (áp dụng toàn bộ API)

```json
{
  "error": {
    "code": "SATELLITE_MANAGED_BY_OPERATOR",
    "message": "This satellite is managed by GitOps and cannot be edited manually."
  }
}
```

## 7. Status codes dùng nhất quán

| Code | Trường hợp |
|---|---|
| 200 | GET/PATCH thành công |
| 201 | POST tạo mới thành công |
| 204 | DELETE thành công |
| 400 | Validation lỗi |
| 401 | Thiếu/sai token |
| 403 | Đúng quyền nhưng bị chặn bởi rule `managedBy` hoặc role |
| 404 | Resource không tồn tại |
| 409 | Conflict (ví dụ tạo Satellite trùng name) |