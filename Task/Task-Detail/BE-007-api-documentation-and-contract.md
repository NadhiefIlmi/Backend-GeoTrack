# BE-007 — API Documentation, Swagger UI & Contract Verification

> **Task ID:** BE-007  
> **Module:** API Documentation / Contract  
> **Priority:** Medium  
> **Dependencies:** BE-001, BE-003, BE-004, BE-005  
> **Estimated Effort:** 1 Day  
> **Status:** COMPLETED (Implemented & Verified on 2026-11-05)

---

## 1. Objective
Menyediakan dokumentasi OpenAPI 3.0 interaktif melalui Swagger UI terintegrasi (`GET /swagger/index.html`), serta memverifikasi keselarasan 100% kontrak antarmuka API (*API Contract Parity*) antara backend Golang dengan tipe data TypeScript dan model Axios client yang digunakan pada aplikasi Frontend GeoTrack.

---

## 2. Scope of Work
1. Menyusun spesifikasi OpenAPI 3.0 dalam format YAML (`docs/openapi.yaml`) dan JSON (`docs/swagger.json`).
2. Menghubungkan Swagger UI handler ke dalam router Chi menggunakan `swaggo/http-swagger` atau static asset embedded (`embed.FS`).
3. Mendokumentasikan seluruh skema entitas:
   - `Entity`, `CreateEntityRequest`, `UpdateEntityRequest`
   - `ActivityLog`
   - `DashboardStats`
   - `ApiResponse<T>`, `ApiListResponse<T>`, `ApiErrorResponse`
4. Melakukan audit komparasi kontrak antara backend DTO dengan model frontend:
   - `src/types/entity.ts`
   - `src/types/activity.ts`
   - `src/types/dashboard.ts`
5. Menyediakan Postman / Bruno Collection (`docs/geotrack_postman_collection.json`) untuk kemudahan pengujian manual tim developer.

---

## 3. Technical Details

### File & Direktori yang Terlibat
```text
Backend-Geotrack/
├── docs/
│   ├── openapi.yaml
│   ├── swagger.json
│   ├── docs.go
│   └── geotrack_postman_collection.json
├── internal/
│   └── delivery/
│       └── http/
│           └── swagger_handler.go
└── Makefile (target: make swagger)
```

### Swagger UI Integration Router
```go
package http

import (
	"net/http"
	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
	_ "geotrack/backend/docs" // register generated docs
)

func RegisterSwaggerRoutes(r chi.Router) {
	// Akses dokumentasi interaktif pada: http://localhost:8080/swagger/index.html
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))
}
```

---

## 4. Frontend-Backend Contract Parity Matrix

Verifikasi keselarasan tipe data dan penamaan field (Backend JSON vs Frontend TypeScript):

### 1. Entitas (`Entity`)
| Field JSON Backend | Tipe Backend Go | Tipe Frontend TypeScript | Status Parity |
|---|---|---|---|
| `id` | `string` (UUID) | `string` | Sesuai |
| `code` | `string` | `string` | Sesuai |
| `name` | `string` | `string` | Sesuai |
| `type` | `"vehicle"` \| `"iot_device"` \| `"facility"` \| `"infrastructure"` | `EntityType` | Sesuai |
| `status` | `"active"` \| `"inactive"` \| `"maintenance"` | `EntityStatus` | Sesuai |
| `latitude` | `float64` | `number` | Sesuai |
| `longitude` | `float64` | `number` | Sesuai |
| `description` | `string` (opsional) | `string?` | Sesuai |
| `created_at` | `string` (ISO 8601 UTC) | `string` | Sesuai |
| `updated_at` | `string` (ISO 8601 UTC) | `string` | Sesuai |

### 2. Log Aktivitas (`ActivityLog`)
| Field JSON Backend | Tipe Backend Go | Tipe Frontend TypeScript | Status Parity |
|---|---|---|---|
| `id` | `string` (UUID) | `string` | Sesuai |
| `entity_id` | `string` (UUID) | `string` | Sesuai |
| `entity_code` | `string` | `string` | Sesuai |
| `entity_name` | `string` | `string` | Sesuai |
| `action` | `"created"` \| `"updated"` \| `"status_changed"` \| `"deleted"` | `ActivityAction` | Sesuai |
| `details` | `string` | `string` | Sesuai |
| `actor` | `string` | `string` | Sesuai |
| `created_at` | `string` (ISO 8601 UTC) | `string` | Sesuai |

### 3. Paginasi Meta (`meta`)
| Field JSON Backend | Tipe Backend Go | Tipe Frontend TypeScript | Status Parity |
|---|---|---|---|
| `page` | `int` | `number` | Sesuai |
| `limit` | `int` | `number` | Sesuai |
| `total_records` | `int` | `number` | Sesuai |
| `total_pages` | `int` | `number` | Sesuai |

---

## 5. Security Requirements
* **Production Swagger Access Control:** Pada environment `production`, rute `/swagger/*` dapat dinonaktifkan atau diproteksi menggunakan HTTP Basic Authentication agar spesifikasi internal API tidak terekspos ke publik bebas.
* **No Sensitive Secrets in Examples:** Parameter contoh (*example values*) di dalam dokumentasi Swagger dilarang keras memuat API key, password, token, atau informasi infrastruktur privat nyata.

---

## 6. Performance Requirements
* **Embedded Static Assets:** Asset HTML, CSS, dan JavaScript Swagger UI dibungkus menggunakan `embed.FS` bawaan Golang (Go 1.16+) sehingga server tidak melakukan pembacaan disk I/O berulang setiap kali antarmuka Swagger dibuka.

---

## 7. Acceptance Criteria
- [x] Swagger UI dapat diakses dengan membuka `http://localhost:8080/swagger/index.html` di browser.
- [x] Fitur "Try it out" pada Swagger UI berfungsi untuk seluruh endpoint (`health`, `entities`, `activities`, `dashboard`).
- [x] Format JSON response dan error response yang dicontohkan di Swagger identik dengan response aktual server.
- [x] File Postman Collection `docs/geotrack_postman_collection.json` tersedia dan siap diimpor.
- [x] Tidak ada perbedaan nama kolom antara model backend dengan frontend (misalnya tidak ada inkonsistensi camelCase vs snake_case).

---

## 8. Testing Scenarios

| Skenario | Langkah Pengujian | Status Code | Ekspektasi |
|---|---|---|---|
| **Akses Swagger UI** | Buka `GET /swagger/index.html` via browser. | `200 OK` | Halaman interaktif Swagger UI termuat lengkap dengan skema GeoTrack. |
| **Akses OpenAPI JSON** | Buka `GET /swagger/doc.json`. | `200 OK` | Dokumen JSON OpenAPI 3.0 valid dan dapat diparsing oleh linter OpenAPI. |
| **Try it out Entity List** | Klik "Execute" pada `GET /api/v1/entities` di UI Swagger. | `200 OK` | Respons aktual diterima dan skema sesuai dengan contoh. |
| **Try it out Create Entity** | Kirim POST payload baru melalui Swagger UI. | `201 Created` | Entity berhasil terbuat dan respons terformat rapi. |

