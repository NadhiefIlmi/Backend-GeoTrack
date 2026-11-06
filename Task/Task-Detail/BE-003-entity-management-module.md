# BE-003 — Entity Management Module (CRUD, Search & Spatial Filters)

> **Task ID:** BE-003  
> **Module:** Entity Management  
> **Priority:** High  
> **Dependencies:** BE-001, BE-002  
> **Estimated Effort:** 2-3 Days  
> **Status:** COMPLETED (Implemented & Verified on 2026-11-05)

---

## 1. Objective
Mengimplementasikan modul inti Entity Management berbasis Clean Architecture yang menyediakan operasi lengkap CRUD (Create, Read, Update, Delete) entitas geospasial (`vehicles`, `iot_devices`, `facilities`, `infrastructures`), mendukung pencarian teks bebas (*fuzzy search*), filter multi-kriteria (`type`, `status`), pengurutan terkontrol (*whitelist sorting*), dan paginasi standar yang 100% kompatibel dengan frontend GeoTrack.

---

## 2. Scope of Work
1. Mendefinisikan struct domain `Entity` dan interface repository/usecase di `internal/domain/entity.go`.
2. Mengimplementasikan PostgreSQL repository di `internal/repository/postgres/entity_repository.go` dengan query dinamis berparameter aman.
3. Mengimplementasikan business logic di `internal/usecase/entity_usecase.go`:
   - Validasi keunikan `code`.
   - Validasi batas rentang geografis (`latitude`: -90 s/d 90, `longitude`: -180 s/d 180).
   - Otomatisasi pemicu pembuatan riwayat log aktivitas saat entitas dibuat, diubah statusnya, diedit, atau dihapus (koordinasi dengan modul Activity Logs).
4. Mengimplementasikan HTTP handler di `internal/delivery/http/entity_handler.go` menggunakan Chi router.
5. Mengintegrasikan parser query params dan validator DTO menggunakan `go-playground/validator/v10`.
6. Memastikan seluruh response dibungkus format envelope konsisten:
   - Success list: `{ success: true, message: "...", data: [...], meta: { page, limit, total_records, total_pages } }`
   - Success single: `{ success: true, message: "...", data: { ... } }`
   - Error: `{ success: false, error: { code: "...", message: "...", details: [...] } }`

---

## 3. Technical Details

### File & Direktori yang Terlibat
```text
Backend-Geotrack/
├── internal/
│   ├── domain/
│   │   ├── entity.go
│   │   └── errors.go
│   ├── repository/
│   │   └── postgres/
│   │       └── entity_repository.go
│   ├── usecase/
│   │   └── entity_usecase.go
│   └── delivery/
│       └── http/
│           ├── entity_handler.go
│           └── dto/
│               └── entity_dto.go
```

### Domain Model & DTO (`internal/domain/entity.go`)
```go
package domain

import (
	"context"
	"time"
)

type EntityType string
const (
	TypeVehicle        EntityType = "vehicle"
	TypeIoTDevice      EntityType = "iot_device"
	TypeFacility       EntityType = "facility"
	TypeInfrastructure EntityType = "infrastructure"
)

type EntityStatus string
const (
	StatusActive      EntityStatus = "active"
	StatusInactive    EntityStatus = "inactive"
	StatusMaintenance EntityStatus = "maintenance"
)

type Entity struct {
	ID          string       `json:"id"`
	Code        string       `json:"code"`
	Name        string       `json:"name"`
	Type        EntityType   `json:"type"`
	Status      EntityStatus `json:"status"`
	Latitude    float64      `json:"latitude"`
	Longitude   float64      `json:"longitude"`
	Description string       `json:"description,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

type EntityFilter struct {
	Search string
	Type   *EntityType
	Status *EntityStatus
	SortBy string // "created_at", "name", "code", "status"
	Order  string // "asc", "desc"
	Page   int
	Limit  int
}

type EntityRepository interface {
	Create(ctx context.Context, entity *Entity) error
	GetByID(ctx context.Context, id string) (*Entity, error)
	GetByCode(ctx context.Context, code string) (*Entity, error)
	List(ctx context.Context, filter EntityFilter) ([]Entity, int, error)
	Update(ctx context.Context, entity *Entity) error
	Delete(ctx context.Context, id string) error
}

type EntityUsecase interface {
	CreateEntity(ctx context.Context, req CreateEntityRequest) (*Entity, error)
	GetEntityByID(ctx context.Context, id string) (*Entity, error)
	ListEntities(ctx context.Context, filter EntityFilter) ([]Entity, MetaPagination, error)
	UpdateEntity(ctx context.Context, id string, req UpdateEntityRequest) (*Entity, error)
	DeleteEntity(ctx context.Context, id string) error
}
```

---

## 4. API Specification

### 1. `GET /api/v1/entities` — List Entities with Filters & Pagination
* **Method:** `GET`
* **Query Parameters:**
  * `page` (integer, default: `1`)
  * `limit` (integer, default: `10`, max: `100`)
  * `search` (string, optional: mencari pada nama atau kode)
  * `type` (string, optional: `vehicle`, `iot_device`, `facility`, `infrastructure`)
  * `status` (string, optional: `active`, `inactive`, `maintenance`)
  * `sort_by` (string, default: `created_at`, allowed: `created_at`, `name`, `code`, `status`)
  * `order` (string, default: `desc`, allowed: `asc`, `desc`)

#### Response Success (`HTTP 200 OK`)
```json
{
  "success": true,
  "message": "Entities retrieved successfully",
  "data": [
    {
      "id": "c1f7a091-8849-4bc2-9e96-a94f30ad8151",
      "code": "VH-001",
      "name": "Fleet Alpha - Truck 01",
      "type": "vehicle",
      "status": "active",
      "latitude": -6.2088,
      "longitude": 106.8456,
      "description": "Logistics delivery vehicle serving West Jakarta area",
      "created_at": "2026-11-04T10:00:00Z",
      "updated_at": "2026-11-04T10:00:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "limit": 10,
    "total_records": 1,
    "total_pages": 1
  }
}
```

---

### 2. `GET /api/v1/entities/{id}` — Get Entity Detail
* **Method:** `GET`
* **Path Parameter:** `id` (UUID string)

#### Response Success (`HTTP 200 OK`)
```json
{
  "success": true,
  "message": "Entity retrieved successfully",
  "data": {
    "id": "c1f7a091-8849-4bc2-9e96-a94f30ad8151",
    "code": "VH-001",
    "name": "Fleet Alpha - Truck 01",
    "type": "vehicle",
    "status": "active",
    "latitude": -6.2088,
    "longitude": 106.8456,
    "description": "Logistics delivery vehicle serving West Jakarta area",
    "created_at": "2026-11-04T10:00:00Z",
    "updated_at": "2026-11-04T10:00:00Z"
  }
}
```

#### Response Error (`HTTP 404 Not Found`)
```json
{
  "success": false,
  "error": {
    "code": "ENTITY_NOT_FOUND",
    "message": "Entity with ID 'c1f7a091-8849-4bc2-9e96-a94f30ad8151' was not found"
  }
}
```

---

### 3. `POST /api/v1/entities` — Create New Entity
* **Method:** `POST`
* **Request Headers:** `Content-Type: application/json`

#### Request Body
```json
{
  "code": "IOT-881",
  "name": "Weather Station Sensor 04",
  "type": "iot_device",
  "status": "active",
  "latitude": -6.1754,
  "longitude": 106.8272,
  "description": "Microclimate monitoring sensor in Central Jakarta"
}
```

#### Response Success (`HTTP 201 Created`)
```json
{
  "success": true,
  "message": "Entity created successfully",
  "data": {
    "id": "e81d4a90-7214-41d3-bc06-40713b145a19",
    "code": "IOT-881",
    "name": "Weather Station Sensor 04",
    "type": "iot_device",
    "status": "active",
    "latitude": -6.1754,
    "longitude": 106.8272,
    "description": "Microclimate monitoring sensor in Central Jakarta",
    "created_at": "2026-11-05T10:30:00Z",
    "updated_at": "2026-11-05T10:30:00Z"
  }
}
```

#### Response Error Validation (`HTTP 422 Unprocessable Entity`)
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Input validation failed",
    "details": [
      {
        "field": "code",
        "issue": "code must be alphanumeric without spaces (e.g. VH-001)"
      },
      {
        "field": "latitude",
        "issue": "latitude must be between -90 and 90"
      }
    ]
  }
}
```

#### Response Error Duplicate (`HTTP 409 Conflict`)
```json
{
  "success": false,
  "error": {
    "code": "DUPLICATE_CODE",
    "message": "Entity code 'IOT-881' is already in use by another entity"
  }
}
```

---

### 4. `PUT /api/v1/entities/{id}` — Update Entity
* **Method:** `PUT`
* **Path Parameter:** `id` (UUID string)
* **Request Body:** Sama dengan request `POST /api/v1/entities`. Jika `code` diubah, usecase akan memvalidasi apakah kode baru sudah dipakai entitas lain.

#### Response Success (`HTTP 200 OK`)
```json
{
  "success": true,
  "message": "Entity updated successfully",
  "data": {
    "id": "e81d4a90-7214-41d3-bc06-40713b145a19",
    "code": "IOT-881",
    "name": "Weather Station Sensor 04 - Updated",
    "type": "iot_device",
    "status": "maintenance",
    "latitude": -6.1754,
    "longitude": 106.8272,
    "description": "Sensor under maintenance for sensor calibration",
    "created_at": "2026-11-05T10:30:00Z",
    "updated_at": "2026-11-05T11:00:00Z"
  }
}
```

---

### 5. `DELETE /api/v1/entities/{id}` — Delete Entity
* **Method:** `DELETE`
* **Path Parameter:** `id` (UUID string)

#### Response Success (`HTTP 200 OK`)
```json
{
  "success": true,
  "message": "Entity deleted successfully"
}
```

---

## 5. Security Requirements
* **Input Sanitization & Whitelisting:** 
  - `SortBy` hanya menerima nilai yang terdaftar dalam whitelist (`created_at`, `name`, `code`, `status`). Jika user mengirim nilai asing (misal `sort_by=password`), fallback otomatis ke `created_at`.
  - `Order` hanya menerima `asc` atau `desc`.
* **SQL Injection Immunity:** Pembuatan query dinamis (`WHERE type = $1 AND status = $2`) menggunakan slice parameter `[]interface{}` dan argumen terindeks. Tidak ada string concatenation mentah.
* **Payload Size Limiter:** Request body dibatasi maksimal 1 MB oleh middleware HTTP untuk mencegah Denial of Service melalui large payload upload.

---

## 6. Performance Requirements
* **Composite Indexes:** Filter gabungan `WHERE type = $1 AND status = $2` dilayani oleh index B-tree yang efisien.
* **Pagination Hardcap:** `limit` dibatasi maksimal 100 record per request untuk melindungi memory footprint server dan serialization overhead.
* **Single Query Count Optimization:** Eksekusi `COUNT(*) OVER()` atau two-step query yang memanfaatkan query planner PostgreSQL dengan context timeout 5 detik.

---

## 7. Acceptance Criteria
- [x] Handler mengembalikan HTTP 200 dengan daftar entitas sesuai filter `search`, `type`, dan `status`.
- [x] Handler menolak input tidak valid dengan HTTP 422 dan pesan rincian field yang jelas.
- [x] Handler menolak kode ganda dengan HTTP 409 Conflict.
- [x] Setiap aksi CRUD sukses otomatis mencatat log pada tabel `activity_logs`.
- [x] Paginasi menghitung `total_records` dan `total_pages` secara akurat.
- [x] Kontrak response 100% cocok dengan interface TypeScript di frontend GeoTrack (`Entity`, `EntityType`, `EntityStatus`).

---

## 8. Testing Scenarios

| Skenario | Langkah Pengujian | Status Code | Ekspektasi |
|---|---|---|---|
| **Create Entity Valid** | POST body valid dengan data lengkap. | `201 Created` | Entity tersimpan di DB, log aktivitas `created` terbuat. |
| **Create Duplicate Code** | POST body dengan `code` yang sudah ada di DB. | `409 Conflict` | Error `DUPLICATE_CODE`, data tidak tersimpan. |
| **Validation Invalid Lat/Lng** | POST dengan `latitude: 95.0`. | `422 Unprocessable` | Error detail mencantumkan field `latitude`. |
| **Filter by Type & Status** | GET `/api/v1/entities?type=vehicle&status=active`. | `200 OK` | Hanya entitas vehicle aktif yang dikembalikan. |
| **Search Keyword** | GET `/api/v1/entities?search=truck`. | `200 OK` | Mengembalikan entitas yang memuat kata "truck" di nama/kode. |
| **Get Non-existent Entity** | GET `/api/v1/entities/invalid-uuid-000`. | `404 Not Found` | Error `ENTITY_NOT_FOUND`. |
| **Update Entity Valid** | PUT `/api/v1/entities/{id}` mengubah `status` ke `maintenance`. | `200 OK` | Status terupdate, log `status_changed` tercatat. |
| **Delete Entity Valid** | DELETE `/api/v1/entities/{id}`. | `200 OK` | Entity terhapus, log `deleted` tercatat. |
| **Sort Whitelist Injection** | GET `/api/v1/entities?sort_by=drop table&order=desc`. | `200 OK` | Query fallback aman ke sorting default `created_at desc`. |

