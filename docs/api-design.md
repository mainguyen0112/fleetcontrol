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