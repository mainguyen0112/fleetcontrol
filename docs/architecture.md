# FleetControl - Architecture

## 1. Tổng quan kiến trúc

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
     Satellite Agent (chạy tại mỗi edge site, heartbeat về API)
```

## 2. Source of Truth (quyết định quan trọng)

- **GitOps/CRD là con đường chính thức** để tạo/sửa Satellite trong môi trường production.
- `fleetctl satellite create/update/delete` chỉ dùng cho dev/test/debug, không dùng cho fleet thật.
- Mỗi Satellite có field `managedBy`:
  - `operator` — được tạo qua CRD/GitOps
  - `manual` — được tạo qua CLI trực tiếp
- Nếu CLI cố sửa một Satellite có `managedBy: operator`, hệ thống sẽ cảnh báo hoặc từ chối.

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

## 4. Các thành phần chính

| Thành phần | Vai trò | Công nghệ |
|---|---|---|
| Control Plane API | Lưu trữ state, expose REST API | Go + PostgreSQL |
| Fleet CLI (fleetctl) | Tương tác thủ công/dev với API | Go + Cobra |
| Fleet Operator | Reconcile CRD ↔ API | Kubebuilder + controller-runtime |
| Satellite Agent | Process thật chạy ở edge site, heartbeat | Go |

## 5. Câu hỏi đã trả lời (tự kiểm tra Phase 0)

- Ai được phép sửa Satellite? → Operator (qua GitOps) là chính thức; CLI chỉ dùng dev/test.
- Chuyện gì xảy ra nếu CLI và Git cùng sửa một Satellite? → `managedBy` ngăn xung đột; CLI cảnh báo nếu cố sửa resource do operator quản lý.