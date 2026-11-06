# BE-001 — Backend Project Initialization & Base Configuration

> **Task ID:** BE-001  
> **Module:** Core / Foundation  
> **Priority:** High  
> **Dependencies:** None  
> **Estimated Effort:** 1 Day  
> **Status:** COMPLETED (Implemented & Verified on 2026-11-05)

---

## 1. Objective
Menginisialisasi fondasi project backend Golang menggunakan Clean Architecture, menyiapkan konfigurasi lingkungan berbasis Viper, menyediakan structured logger Zerolog, mengatur router Chi v5 dengan graceful shutdown, dan menyediakan health check endpoint sebagai verifikasi awal sistem.

---

## 2. Scope of Work
1. Inisialisasi Go Module dengan nama module `geotrack/backend` (Go 1.23+).
2. Membangun struktur direktori Clean Architecture (`cmd/`, `internal/config`, `internal/domain`, `internal/delivery/http`, `internal/middleware`, `pkg/logger`, `pkg/validator`).
3. Membuat modul konfigurasi `internal/config/config.go` yang membaca environment variables dari `.env` dengan fallback default values.
4. Membuat wrapper structured logger berbasis `rs/zerolog` di `pkg/logger/logger.go`.
5. Mengonfigurasi router Chi v5 dan HTTP server dengan OS signal listener (SIGINT, SIGTERM) untuk mekanisme Graceful Shutdown.
6. Menyediakan endpoint Health Check: `GET /api/v1/health`.
7. Menyiapkan `.env.example`, `.golangci.yml`, dan `Makefile`.

---

## 3. Technical Details

### File & Direktori yang Terlibat
```text
Backend-Geotrack/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   └── delivery/
│       └── http/
│           ├── router.go
│           ├── response.go
│           └── health_handler.go
├── pkg/
│   └── logger/
│       └── logger.go
├── .env.example
├── .golangci.yml
├── Makefile
└── go.mod
```

### Business & Technical Logic
1. **Config Loader (`internal/config/config.go`):**
   * Menggunakan Viper untuk membaca konfigurasi:
     * `APP_NAME` (default: "GeoTrack Backend")
     * `APP_PORT` (default: 8080)
     * `APP_ENV` (default: "development", values: "development", "staging", "production")
     * `CORS_ALLOWED_ORIGINS` (default: "http://localhost:8081,http://localhost:5173")
     * `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE`
     * `DB_MAX_CONNS` (default: 25), `DB_MIN_CONNS` (default: 5)
2. **Graceful Shutdown (`cmd/api/main.go`):**
   * Menjalankan server HTTP di dalam goroutine terpisah.
   * Menangkap sinyal OS `syscall.SIGINT` dan `syscall.SIGTERM`.
   * Memberikan timeout grace period 10 detik menggunakan `context.WithTimeout` untuk menyelesaikan in-flight requests sebelum proses keluar bersih.
3. **Structured Logger (`pkg/logger/logger.go`):**
   * Pada mode development: menggunakan `zerolog.ConsoleWriter` berwarna agar mudah dibaca di terminal.
   * Pada mode production: menghasilkan log JSON terstruktur dengan timestamp UTC dan level log.

---

## 4. API Specification: Health Check

* **Method:** `GET`
* **Endpoint:** `/api/v1/health`
* **Authentication:** None (Public)
* **Headers:** `Accept: application/json`

### Success Response (`HTTP 200 OK`)
```json
{
  "success": true,
  "message": "Service is healthy and operational",
  "data": {
    "status": "UP",
    "version": "1.0.0",
    "environment": "development",
    "uptime_seconds": 342,
    "timestamp": "2026-11-05T10:00:00Z"
  }
}
```

---

## 5. Security Requirements
* **Environment Secret Isolation:** File `.env` wajib masuk dalam `.gitignore` dan tidak boleh pernah di-commit ke repository. Template variabel rahasia didefinisikan di `.env.example`.
* **Zero Leakage Error Handling:** Jika server mengalami internal failure, HTTP handler tidak boleh membocorkan direktori sistem atau stack trace.

---

## 6. Performance Requirements
* **Startup Time:** Waktu inisialisasi server hingga siap menerima koneksi HTTP < 500 ms.
* **Low Memory Footprint:** Penggunaan memory saat server idle < 25 MB.
* **Zero-Allocation Logging:** Zerolog menghindari alokasi heap berlebih pada setiap request log.

---

## 7. Acceptance Criteria
- [x] Go module terinisialisasi dengan Go 1.23+ dan dependencies minimal yang disetujui.
- [x] Server berhasil dijalankan pada port yang ditentukan (`APP_PORT=8080`).
- [x] Endpoint `GET /api/v1/health` mengembalikan HTTP 200 dengan format JSON envelope standar.
- [x] Sinyal `Ctrl+C` (SIGINT) memicu graceful shutdown dengan log penutupan yang tertib tanpa pemutusan koneksi paksa.
- [x] `go test -race ./...` dan pengujian unit lulus 100%.

---

## 8. Testing Scenario

| Test Case | Langkah Pengujian | Expected Result | Status |
| :--- | :--- | :--- | :---: |
| **TC-001: Health Endpoint Success** | Kirim `GET /api/v1/health` | HTTP 200, `"status": "UP"`, format JSON standar | **PASSED** |
| **TC-002: Config Fallback Defaults** | Jalankan app tanpa file `.env` | App tetap berjalan dengan default port 8080 & env development | **PASSED** |
| **TC-003: Graceful Shutdown Handling** | Kirim sinyal `kill -SIGTERM <pid>` saat server running | Server menyelesaikan in-flight request, menutup listener, dan log shutdown muncul | **PASSED** |
| **TC-004: Race & Concurrency Check** | Eksekusi `go test -race ./...` | 0 data race detected, all packages pass | **PASSED** |

---

## 9. Testing Evidence Report
Laporan lengkap hasil pengujian dicatat pada [BE-001-test-report.md](file:///Users/ptsiagaabdiutama/Test%20Software/Backend-Geotrack/docs/testing/BE-001-test-report.md).

