# BE-004 — Activity Logs & Audit Trail Module

> **Task ID:** BE-004  
> **Module:** Activity Logs  
> **Priority:** Medium  
> **Dependencies:** BE-002, BE-003  
> **Estimated Effort:** 1-2 Days  
> **Status:** COMPLETED (Implemented & Verified on 2026-11-05)

---

## 1. Objective
Mengembangkan modul audit trail Activity Logs yang bertugas mencatat secara otomatis setiap riwayat perubahan status, pembuatan data, pengubahan atribut, dan penghapusan entitas geospasial, serta menyediakan REST API bagi antarmuka frontend GeoTrack untuk menampilkan linimasa aktivitas (*activity timeline*) dengan dukungan pencarian teks, filter tipe aksi, dan paginasi.

---

## 2. Scope of Work
1. Mendefinisikan struct domain `ActivityLog`, konstanta `ActivityAction`, dan interface repository/usecase di `internal/domain/activity_log.go`.
2. Mengimplementasikan PostgreSQL repository di `internal/repository/postgres/activity_log_repository.go` (`Create`, `List`, `GetByID`, `GetByEntityID`).
3. Mengimplementasikan business logic di `internal/usecase/activity_log_usecase.go`:
   - Mekanisme logging sinkron yang dipanggil oleh `EntityUsecase` saat operasi mutasi data berhasil.
   - Pembangkitan ringkasan teks deskriptif otomatis (contoh: *"Status changed from active to maintenance"*, *"Entity VH-001 created"*).
   - Pengelolaan snapshot data entitas (`entity_code`, `entity_name`) agar riwayat audit tetap terbaca meskipun entitas aslinya telah dihapus.
4. Mengimplementasikan HTTP handler di `internal/delivery/http/activity_log_handler.go` untuk endpoint:
   - `GET /api/v1/activities` (pencarian, filter aksi, paginasi).
   - `GET /api/v1/activities/{id}` (rincian log).
5. Memastikan kontrak JSON cocok 100% dengan tipe TypeScript di frontend: `ActivityLog` (`id`, `entity_id`, `entity_code`, `entity_name`, `action`, `details`, `actor`, `created_at`).

---

## 3. Technical Details

### File & Direktori yang Terlibat
```text
Backend-Geotrack/
├── internal/
│   ├── domain/
│   │   └── activity_log.go
│   ├── repository/
│   │   └── postgres/
│   │       └── activity_log_repository.go
│   ├── usecase/
│   │   └── activity_log_usecase.go
│   └── delivery/
│       └── http/
│           ├── activity_log_handler.go
│           └── dto/
│               └── activity_log_dto.go
```

### Domain Model (`internal/domain/activity_log.go`)
```go
package domain

import (
	"context"
	"time"
)

type ActivityAction string
const (
	ActionCreated       ActivityAction = "created"
	ActionUpdated       ActivityAction = "updated"
	ActionStatusChanged ActivityAction = "status_changed"
	ActionDeleted       ActivityAction = "deleted"
)

type ActivityLog struct {
	ID         string         `json:"id"`
	EntityID   string         `json:"entity_id"`
	EntityCode string         `json:"entity_code"`
	EntityName string         `json:"entity_name"`
	Action     ActivityAction `json:"action"`
	Details    string         `json:"details"`
	Actor      string         `json:"actor"`
	CreatedAt  time.Time      `json:"created_at"`
}

type ActivityFilter struct {
	Search   string
	Action   *ActivityAction
	EntityID string
	Page     int
	Limit    int
}

type ActivityLogRepository interface {
	Create(ctx context.Context, log *ActivityLog) error
	List(ctx context.Context, filter ActivityFilter) ([]ActivityLog, int, error)
	GetByID(ctx context.Context, id string) (*ActivityLog, error)
	GetByEntityID(ctx context.Context, entityID string, limit int) ([]ActivityLog, error)
}

type ActivityLogUsecase interface {
	RecordActivity(ctx context.Context, entity Entity, action ActivityAction, details string, actor string) error
	ListActivities(ctx context.Context, filter ActivityFilter) ([]ActivityLog, MetaPagination, error)
	GetActivityByID(ctx context.Context, id string) (*ActivityLog, error)
}
```

---

## 4. API Specification

### 1. `GET /api/v1/activities` — List Activity Logs
* **Method:** `GET`
* **Query Parameters:**
  * `page` (integer, default: `1`)
  * `limit` (integer, default: `10`, max: `100`)
  * `search` (string, optional: mencari pada nama entitas, kode, atau rincian)
  * `action` (string, optional: `created`, `updated`, `status_changed`, `deleted`)
  * `entity_id` (string UUID, optional: menyaring riwayat milik satu entitas tertentu)

#### Response Success (`HTTP 200 OK`)
```json
{
  "success": true,
  "message": "Activity logs retrieved successfully",
  "data": [
    {
      "id": "a9b2c3d4-e5f6-7890-abcd-ef1234567890",
      "entity_id": "c1f7a091-8849-4bc2-9e96-a94f30ad8151",
      "entity_code": "VH-001",
      "entity_name": "Fleet Alpha - Truck 01",
      "action": "status_changed",
      "details": "Status changed from active to maintenance",
      "actor": "Admin Dispatcher",
      "created_at": "2026-11-05T09:15:00Z"
    },
    {
      "id": "b8c3d4e5-f6a7-8901-bcde-f12345678901",
      "entity_id": "c1f7a091-8849-4bc2-9e96-a94f30ad8151",
      "entity_code": "VH-001",
      "entity_name": "Fleet Alpha - Truck 01",
      "action": "created",
      "details": "New vehicle entity created in West Jakarta sector",
      "actor": "System",
      "created_at": "2026-11-04T10:00:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "limit": 10,
    "total_records": 2,
    "total_pages": 1
  }
}
```

---

### 2. `GET /api/v1/activities/{id}` — Get Activity Log Detail
* **Method:** `GET`
* **Path Parameter:** `id` (UUID string)

#### Response Success (`HTTP 200 OK`)
```json
{
  "success": true,
  "message": "Activity log retrieved successfully",
  "data": {
    "id": "a9b2c3d4-e5f6-7890-abcd-ef1234567890",
    "entity_id": "c1f7a091-8849-4bc2-9e96-a94f30ad8151",
    "entity_code": "VH-001",
    "entity_name": "Fleet Alpha - Truck 01",
    "action": "status_changed",
    "details": "Status changed from active to maintenance",
    "actor": "Admin Dispatcher",
    "created_at": "2026-11-05T09:15:00Z"
  }
}
```

#### Response Error (`HTTP 404 Not Found`)
```json
{
  "success": false,
  "error": {
    "code": "ACTIVITY_NOT_FOUND",
    "message": "Activity log with ID 'a9b2c3d4-e5f6-7890-abcd-ef1234567890' was not found"
  }
}
```

---

## 5. Security & Audit Integrity
* **Immutable Audit Trail:** Tabel `activity_logs` bersifat *append-only*. Tidak ada endpoint HTTP `PUT`, `PATCH`, atau `DELETE` yang diekspos untuk riwayat aktivitas. Log yang sudah dibuat tidak boleh diubah atau dimanipulasi oleh siapapun.
* **Actor Attribution:** Nilai aktor diperoleh dari header autentikasi masa depan atau default `"System"` / `"Admin User"`, menjamin kepastian siapa pelaku modifikasi.
* **SQL Injection Prevention:** Seluruh query filter pencarian teks bebas menggunakan `ILIKE $1` dengan binding parameter yang disanitasi.

---

## 6. Performance Requirements
* **Chronological Indexing:** Index `idx_activity_created_at (created_at DESC)` memastikan query pembacaan urutan waktu (*timeline*) tidak memicu full-table sorting di memori PostgreSQL.
* **Non-blocking Log Writes:** Penulisan log dapat dijalankan di dalam unit transaksi database yang sama dengan mutasi entity untuk menjamin ACID consistency, atau via asynchronous channel terproteksi jika throughput mutasi sangat masif.

---

## 7. Acceptance Criteria
- [x] Setiap pemanggilan `CreateEntity`, `UpdateEntity`, dan `DeleteEntity` menghasilkan 1 baris log baru di `activity_logs`.
- [x] Endpoint `GET /api/v1/activities` mengembalikan daftar aktivitas terurut dari yang paling mutakhir (`created_at DESC`).
- [x] Filter `action` berfungsi akurat (misal: hanya menampilkan `status_changed`).
- [x] Snapshot nama dan kode entitas tetap tersimpan di riwayat log meskipun entitas terkait dihapus.
- [x] Format JSON memenuhi standar TypeScript di frontend.

---

## 8. Testing Scenarios

| Skenario | Langkah Pengujian | Status Code | Ekspektasi |
|---|---|---|---|
| **Audit on Create** | Buat entitas baru via POST `/api/v1/entities`. | `201 Created` | Di tabel log muncul action `created` dengan actor `System`. |
| **Audit on Status Change** | Ubah status entitas dari `active` ke `maintenance`. | `200 OK` | Di tabel log muncul action `status_changed` dengan rincian perubahan. |
| **List Activities Default** | GET `/api/v1/activities`. | `200 OK` | Daftar log terurut descending berdasarkan timestamp `created_at`. |
| **Filter by Action** | GET `/api/v1/activities?action=deleted`. | `200 OK` | Hanya rekaman aksi `deleted` yang tampil. |
| **Search Log Keyword** | GET `/api/v1/activities?search=Alpha`. | `200 OK` | Seluruh aktivitas yang memuat kata "Alpha" ditemukan. |
| **Get Non-existent Log** | GET `/api/v1/activities/00000000-0000-0000-0000-000000000000`. | `404 Not Found` | Error `ACTIVITY_NOT_FOUND`. |

