# BE-005 — Dashboard Analytics & Aggregated Metrics API

> **Task ID:** BE-005  
> **Module:** Dashboard / Analytics  
> **Priority:** Medium  
> **Dependencies:** BE-002, BE-003, BE-004  
> **Estimated Effort:** 1-2 Days  
> **Status:** COMPLETED (Implemented & Verified on 2026-11-05)

---

## 1. Objective
Menyediakan REST API analitik dashboard berkinerja tinggi (`GET /api/v1/dashboard/stats`) yang mengumpulkan, mengagregasi, dan meringkas metrik operasional seluruh entitas geospasial (total aset, distribusi per status, distribusi per tipe aset) serta melampirkan daftar aktivitas terbaru (*recent activities*) dalam satu kali pemanggilan jaringan tunggal untuk kebutuhan visualisasi di frontend GeoTrack.

---

## 2. Scope of Work
1. Mendefinisikan struct domain `DashboardStats`, `StatusBreakdown`, `TypeBreakdown`, dan interface usecase/repository di `internal/domain/dashboard.go`.
2. Mengimplementasikan PostgreSQL repository di `internal/repository/postgres/dashboard_repository.go` dengan query agregasi teroptimasi menggunakan `COUNT(*) FILTER (WHERE ...)` atau CTE untuk menghindari round-trip database berulang.
3. Mengimplementasikan business logic di `internal/usecase/dashboard_usecase.go` yang mengombinasikan metrik agregasi entitas dengan 5 rekaman aktivitas paling mutakhir dari modul `activity_logs`.
4. Mengimplementasikan HTTP handler di `internal/delivery/http/dashboard_handler.go` untuk rute `GET /api/v1/dashboard/stats`.
5. Memastikan kontrak data sesuai dengan kebutuhan komponen `Dashboard` dan kartu statistik di frontend GeoTrack.

---

## 3. Technical Details

### File & Direktori yang Terlibat
```text
Backend-Geotrack/
├── internal/
│   ├── domain/
│   │   └── dashboard.go
│   ├── repository/
│   │   └── postgres/
│   │       └── dashboard_repository.go
│   ├── usecase/
│   │   └── dashboard_usecase.go
│   └── delivery/
│       └── http/
│           └── dashboard_handler.go
```

### Domain Model (`internal/domain/dashboard.go`)
```go
package domain

import "context"

type StatusBreakdown struct {
	Active      int `json:"active"`
	Inactive    int `json:"inactive"`
	Maintenance int `json:"maintenance"`
}

type TypeBreakdown struct {
	Vehicles        int `json:"vehicles"`
	IoTDevices      int `json:"iot_devices"`
	Facilities      int `json:"facilities"`
	Infrastructures int `json:"infrastructures"`
}

type DashboardStats struct {
	TotalEntities    int             `json:"total_entities"`
	StatusSummary    StatusBreakdown `json:"status_summary"`
	TypeSummary      TypeBreakdown   `json:"type_summary"`
	RecentActivities []ActivityLog   `json:"recent_activities"`
}

type DashboardRepository interface {
	GetAggregateCounts(ctx context.Context) (total int, status StatusBreakdown, types TypeBreakdown, err error)
}

type DashboardUsecase interface {
	GetStats(ctx context.Context) (*DashboardStats, error)
}
```

### Optimized Aggregation SQL Query
Daripada mengeksekusi 8 kali query `SELECT COUNT(*)` terpisah yang membebani database, repository menggunakan satu query tunggal berbasis conditional aggregation:
```sql
SELECT 
    COUNT(*) AS total_entities,
    COUNT(*) FILTER (WHERE status = 'active') AS active_count,
    COUNT(*) FILTER (WHERE status = 'inactive') AS inactive_count,
    COUNT(*) FILTER (WHERE status = 'maintenance') AS maintenance_count,
    COUNT(*) FILTER (WHERE type = 'vehicle') AS vehicle_count,
    COUNT(*) FILTER (WHERE type = 'iot_device') AS iot_count,
    COUNT(*) FILTER (WHERE type = 'facility') AS facility_count,
    COUNT(*) FILTER (WHERE type = 'infrastructure') AS infra_count
FROM entities;
```

---

## 4. API Specification

### `GET /api/v1/dashboard/stats` — Get Dashboard Operational Statistics
* **Method:** `GET`
* **Endpoint:** `/api/v1/dashboard/stats`
* **Authentication:** None (Public / App Key)
* **Headers:** `Accept: application/json`

#### Response Success (`HTTP 200 OK`)
```json
{
  "success": true,
  "message": "Dashboard statistics retrieved successfully",
  "data": {
    "total_entities": 8,
    "status_summary": {
      "active": 5,
      "inactive": 1,
      "maintenance": 2
    },
    "type_summary": {
      "vehicles": 3,
      "iot_devices": 2,
      "facilities": 2,
      "infrastructures": 1
    },
    "recent_activities": [
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
        "entity_id": "e81d4a90-7214-41d3-bc06-40713b145a19",
        "entity_code": "IOT-881",
        "entity_name": "Weather Station Sensor 04",
        "action": "created",
        "details": "New IoT device registered",
        "actor": "System",
        "created_at": "2026-11-05T08:30:00Z"
      }
    ]
  }
}
```

---

## 5. Security Requirements
* **Context Timeout Enforcement:** Eksekusi query agregasi wajib dibatasi dengan `context.WithTimeout` maksimal 3 detik agar beban komputasi analitik tidak memblokir worker pool database saat lonjakan trafik.
* **Read-Only Transaction Isolation:** Transaksi atau koneksi yang membaca agregasi dashboard dapat diset pada level `READ COMMITTED` untuk mencegah overhead locking.

---

## 6. Performance Requirements
* **Zero N+1 Query:** Seluruh perhitungan metrik diakumulasikan dalam 1 round-trip SQL, disusul dengan 1 query limit 5 untuk log aktivitas terbaru. Total hanya 2 query terisolasi.
* **Response Payload Lightweight:** Ukuran payload JSON dipertahankan < 5 KB untuk memastikan halaman utama GeoTrack memuat instan (< 100 ms latensi server).
* **In-Memory Cache (Opsional/Masa Depan):** Jika volume data entitas mencapai > 100.000 record, disiapkan interface caching (in-memory go-cache atau Redis) dengan TTL 30 detik.

---

## 7. Acceptance Criteria
- [x] Endpoint mengembalikan data `status_summary` dengan total yang sama persis dengan `total_entities`.
- [x] Endpoint mengembalikan data `type_summary` dengan total yang sama persis dengan `total_entities`.
- [x] `recent_activities` memuat maksimal 5 rekaman aktivitas terbaru terurut secara descending.
- [x] Response terbungkus rapi dalam standard response envelope GeoTrack.
- [x] Perubahan status atau pembuatan entitas baru langsung tercermin pada respons endpoint dashboard berikutnya.

---

## 8. Testing Scenarios

| Skenario | Langkah Pengujian | Status Code | Ekspektasi |
|---|---|---|---|
| **Ambil Statistik Lengkap** | GET `/api/v1/dashboard/stats` saat DB berisi data seeder. | `200 OK` | Jumlah total entitas dan rincian status terhitung konsisten. |
| **Kondisi Database Kosong** | Kosongkan seluruh tabel `entities` dan `activity_logs`, lalu panggil endpoint. | `200 OK` | Mengembalikan total 0, seluruh breakdown bernilai 0, dan array aktivitas kosong `[]` (bukan null). |
| **Konsistensi Real-time** | Tambahkan 1 entitas baru, lalu langsung panggil stats endpoint. | `200 OK` | `total_entities` bertambah 1, tipe terkait bertambah 1, aktivitas `created` muncul di posisi teratas. |
| **Koneksi Database Timeout** | Simulasikan delay network pada query agregasi. | `504 Gateway Timeout` / `500 Internal Error` | Handler memotong proses sesuai timeout context tanpa membocorkan stack trace. |

