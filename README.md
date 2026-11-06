# GeoTrack Backend — Geospatial Entity Management System

GeoTrack Backend adalah REST API berkinerja tinggi yang dirancang untuk mengelola entitas geospasial (*Vehicles, IoT Devices, Facilities, Infrastructure*) secara terpusat. Dibangun menggunakan bahasa **Golang** dengan menerapkan prinsip **Clean Architecture** (decoupling ketat antara domain bisnis, database, dan delivery layer).

> 🔗 **Repositori Terkait:** Antarmuka web pengguna (*frontend*) untuk sistem ini tersedia di [**Frontend-GeoTrack (TanStack Start & React 19)**](https://github.com/NadhiefIlmi/Frontend-GeoTrack).

---

## 🚀 Fitur Utama

- **Clean Architecture:** Pemisahan tanggung jawab yang jelas antara `domain`, `usecase`, `repository`, dan `delivery`.
- **Health & Lifecycle Management:** Graceful shutdown terkelola menangani sinyal OS (`SIGINT`, `SIGTERM`) dengan deadline context.
- **Entity Management:** CRUD lengkap dengan validasi geografis (*Latitude/Longitude*), pencarian teks bebas (*fuzzy search*), filter tipe/status, dan paginasi.
- **Audit Trail & Activity Logs:** Pencatatan otomatis setiap mutasi entitas untuk pelacakan linimasa aktivitas historis.
- **Dashboard Analytics:** Endpoint agregasi tunggal berkinerja tinggi untuk metrik operasional tanpa masalah *N+1 query*.
- **Security Hardening:** Whitelisted CORS (`http://localhost:8081`), Token Bucket Rate Limiting per IP client, proteksi panic recovery, dan pelacakan request ID (`X-Request-ID`).
- **Structured Logging:** Logging berbasis `rs/zerolog` dengan zero-memory allocation.

---

## 🛠️ Tech Stack

| Komponen | Pilihan Teknologi | Versi / Keterangan |
|---|---|---|
| **Bahasa Pemrograman** | Go (Golang) | `1.23+` (Tested on `go1.26.5 darwin/arm64`) |
| **HTTP Router** | `go-chi/chi/v5` | Router modular, ringan, dan standar idiomatic Go |
| **Database Driver** | `jackc/pgx/v5` | Driver PostgreSQL native berkinerja tinggi dengan connection pooling |
| **Structured Logger** | `rs/zerolog` | Logger cepat berformat JSON / Colored Console |
| **Config Loader** | `spf13/viper` | Manajemen environment variable dan file `.env` |
| **Migration Runner** | `golang-migrate/migrate/v4` | Migrasi skema database terkelola (up/down) |
| **Testing Suite** | `stretchr/testify` | Unit testing assertions & mocks |

---

## 📁 Struktur Direktori & Penjelasan File

Struktur proyek mengadopsi prinsip **Clean Architecture** (decoupling ketat antara domain bisnis, persistensi data, dan protokol pengiriman):

```text
Backend-Geotrack/
├── cmd/
│   ├── api/
│   │   └── main.go                         # Entrypoint HTTP server & dependency injection wiring
│   └── migrate/
│       └── main.go                         # CLI runner migrasi skema database manual
├── internal/
│   ├── config/
│   │   ├── config.go                       # Konfigurasi aplikasi via Viper & pemetaan .env
│   │   └── config_test.go                  # Unit test validasi parsing konfigurasi
│   ├── domain/
│   │   ├── entity.go                       # Model entitas geospasial, enum tipe/status & kontrak interface
│   │   ├── activity_log.go                 # Model riwayat aktivitas (audit trail) & kontrak interface
│   │   ├── dashboard.go                    # Struct agregasi metrik statistik analitik
│   │   └── errors.go                       # Definisi standar sentry error domain bisnis
│   ├── repository/
│   │   └── postgres/
│   │       ├── connection.go               # Manajemen connection pool PostgreSQL (pgxpool)
│   │       ├── migrator.go                 # Helper otomatis migrasi DDL via golang-migrate
│   │       ├── entity_repository.go        # Query SQL CRUD, pencarian teks bebas (fuzzy), & paginasi
│   │       ├── activity_log_repository.go  # Query SQL pencatatan & pembacaan riwayat aktivitas
│   │       └── dashboard_repository.go     # Query SQL agregasi metrik dashboard berkinerja tinggi
│   ├── usecase/
│   │   ├── entity_usecase.go               # Logika bisnis validasi entitas & trigger audit log otomatis
│   │   ├── activity_log_usecase.go         # Logika bisnis riwayat aktivitas
│   │   ├── dashboard_usecase.go            # Logika bisnis penyusunan analitik dashboard
│   │   └── benchmark_test.go               # Pengujian performa komputasi agregasi data
│   ├── delivery/
│   │   └── http/
│   │       ├── router.go                   # Pendaftaran routing Chi v5 & mounting middleware
│   │       ├── response.go                 # Standar pembungkus format respons JSON envelope
│   │       ├── entity_handler.go           # HTTP Handler CRUD entitas (/api/v1/entities)
│   │       ├── activity_log_handler.go     # HTTP Handler audit log (/api/v1/activities)
│   │       ├── dashboard_handler.go        # HTTP Handler metrik agregasi (/api/v1/dashboard/stats)
│   │       ├── health_handler.go           # HTTP Handler health check uptime (/api/v1/health)
│   │       └── swagger_handler.go          # HTTP Handler penyaji OpenAPI JSON & Swagger UI
│   └── middleware/
│       ├── cors.go                         # Whitelist CORS untuk origin frontend
│       ├── rate_limiter.go                 # Token bucket rate limiting per IP client
│       ├── recoverer.go                    # Panic recovery pencegah crash server
│       ├── request_id.go                   # Injeksi header pelacakan X-Request-ID
│       └── security.go                     # Penguatan header keamanan HTTP bawaan
├── pkg/
│   ├── logger/
│   │   ├── logger.go                       # Inisialisasi structured logging berbasis Zerolog
│   │   └── logger_test.go                  # Unit test format log JSON
│   └── validator/
│       ├── validator.go                    # Validasi koordinat spasial & payload request
│       └── validator_test.go               # Unit test batasan koordinat lat/lng
├── migrations/
│   ├── 000001_create_entities_table.up.sql # DDL pembuatan tabel entities & indeks
│   ├── 000001_create_entities_table.down.sql
│   ├── 000002_create_activity_logs_table.up.sql # DDL pembuatan tabel activity_logs
│   ├── 000002_create_activity_logs_table.down.sql
│   └── seeds/
│       └── 000001_seed_entities.sql        # Data awal geografis entitas Jakarta
├── tests/
│   └── mocks/                              # Mock interface repository & usecase untuk unit testing
├── docs/                                   # Spesifikasi OpenAPI (swagger.json) & Postman collection
├── .env.example                            # Template environment variable
├── .env                                    # Konfigurasi lokal lingkungan kerja
├── Makefile                                # Perintah automasi build, test, lint, dan migrasi
├── go.mod                                  # Definisi modul Go & dependensi pustaka
└── go.sum                                  # Checksum integritas dependensi
```

### 📑 Rincian Fungsi Per-Modul & File

#### 1. `cmd/` (Application Entrypoints)
* **`cmd/api/main.go`**: Titik masuk utama aplikasi server HTTP. Bertugas membaca konfigurasi lingkungan via Viper, menginisialisasi structured logger Zerolog, menyambungkan *connection pool* database PostgreSQL, merangkai dependency injection (`Repository` ➔ `Usecase` ➔ `Handler`), mendaftarkan rute Chi, serta mengelola *Graceful Shutdown* saat menerima sinyal OS (`SIGINT`, `SIGTERM`).
* **`cmd/migrate/main.go`**: Program CLI mandiri untuk mengeksekusi migrasi skema database DDL secara langsung (opsi *up* atau *down*) di luar proses utama HTTP server.

#### 2. `internal/domain/` (Core Business Rules & Models)
* **`entity.go`**: Mendefinisikan struct entitas geospasial (`Entity`), tipe enum jenis aset (`vehicle`, `iot_device`, `facility`, `infrastructure`), tipe status (`active`, `inactive`, `maintenance`), batas koordinat validasi (*Latitude: -90 s/d 90, Longitude: -180 s/d 180*), serta kontrak interface `EntityRepository` dan `EntityUsecase`.
* **`activity_log.go`**: Mendefinisikan struct jejak audit (`ActivityLog`), jenis mutasi aksi (`created`, `updated`, `status_changed`, `deleted`), rincian perubahan, serta kontrak interface `ActivityLogRepository` dan `ActivityLogUsecase`.
* **`dashboard.go`**: Mendefinisikan model agregasi statistik operasional (`DashboardStats`, `StatusSummary`, `TypeBreakdown`) yang dikonsumsi oleh dashboard frontend.
* **`errors.go`**: Kamus sentry error terstandarisasi untuk domain bisnis (misal: `ErrEntityNotFound`, `ErrDuplicateCode`, `ErrInvalidCoordinates`, `ErrInvalidStatus`).

#### 3. `internal/repository/postgres/` (Data Persistence Layer)
* **`connection.go`**: Mengelola siklus hidup *connection pool* PostgreSQL menggunakan driver native `pgxpool`, lengkap dengan konfigurasi batas koneksi (`MaxConns`, `MinConns`) dan health ping verification.
* **`migrator.go`**: Modul pembantu otomatis yang menjalankan berkas migrasi SQL secara terprogram saat server pertama kali dinyalakan.
* **`entity_repository.go`**: Implementasi query SQL murni untuk operasi CRUD entitas, pencarian teks bebas *case-insensitive* (*fuzzy search* `ILIKE`), pemfilteran multi-kriteria berbasis tipe/status, dan paginasi data aman dari SQL Injection.
* **`activity_log_repository.go`**: Implementasi persistensi riwayat audit trail ke tabel `activity_logs` dan pengambilan linimasa historis berdasarkan entitas atau filter waktu.
* **`dashboard_repository.go`**: Menjalankan query agregasi SQL tunggal berkinerja tinggi menggunakan ekspresi `COUNT(*) FILTER (...)` untuk menyajikan metrik tanpa beban query berulang (*N+1 problem*).

#### 4. `internal/usecase/` (Application Business Logic)
* **`entity_usecase.go`**: Orkestrasi logika bisnis entitas: melakukan sanitasi data, memvalidasi koordinat spasial, memverifikasi keunikan kode entitas (`code`), serta memicu pembuatan rekaman audit aktivitas secara otomatis saat terjadi penambahan, pengubahan, atau penghapusan entitas.
* **`activity_log_usecase.go`**: Logika bisnis untuk penyaringan, pencarian, dan penyajian data riwayat aktivitas pengguna/sistem.
* **`dashboard_usecase.go`**: Logika bisnis untuk memformat dan menghitung metrik persentase kesehatan entitas sebelum disajikan ke API delivery.

#### 5. `internal/delivery/http/` (API Delivery Layer)
* **`router.go`**: Mengonfigurasi router `go-chi/chi/v5`, memasang middleware global, dan memetakan endpoint REST API pada grup `/api/v1`.
* **`response.go`**: Helper utilitas untuk membungkus semua keluaran HTTP dalam format standar envelope JSON yang konsisten (`success`, `message`, `data`, `meta`, atau payload error terstruktur).
* **`entity_handler.go`**: Menerima request HTTP, mem-parsing query parameter/body JSON, memanggil `EntityUsecase`, dan mengembalikan respons entitas.
* **`activity_log_handler.go`**: Menangani permintaan riwayat aktivitas dan paginasi audit log.
* **`dashboard_handler.go`**: Menyajikan metrik analitik ringkasan operasional pada endpoint `/api/v1/dashboard/stats`.
* **`health_handler.go`**: Menyajikan status kesehatan server, status koneksi database, dan durasi aktif (*uptime*) pada `/api/v1/health`.
* **`swagger_handler.go`**: Menyajikan berkas spesifikasi `swagger.json` dan antarmuka visual dokumentasi Swagger UI interaktif pada `/swagger/index.html`.

#### 6. `internal/middleware/` (Security & Request Pipelines)
* **`cors.go`**: Mengatur header *Cross-Origin Resource Sharing* dengan daftar asal yang diizinkan (*whitelisted origins*: `http://localhost:8081` dan `http://localhost:5173`).
* **`rate_limiter.go`**: Mencegah serangan *Brute-Force* dan banjir request melalui algoritma *Token Bucket* berbasis alamat IP client.
* **`recoverer.go`**: Menangkap insiden *panic* yang tidak terduga pada thread handler agar proses server tidak mati (*crash*) dan merespons dengan HTTP 500 terstruktur.
* **`request_id.go`**: Menginjeksi dan meneruskan identitas penelusuran request unik (`X-Request-ID`) ke dalam header respons dan context logger.
* **`security.go`**: Memasang header perlindungan standar peramban web modern (`X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `X-XSS-Protection: 1; mode=block`).

#### 7. `pkg/` (Shared Utility Packages)
* **`logger/logger.go`**: Pembungkus terpusat untuk *structured logger* `rs/zerolog` yang mendukung format JSON di lingkungan produksi serta tampilan berwarna yang mudah dibaca di lingkungan pengembangan.
* **`validator/validator.go`**: Kumpulan fungsi utilitas validasi data geospasial murni tanpa ketergantungan layer database.

#### 8. `migrations/` (Database Migrations & Seeds)
* **`000001_create_entities_table.up.sql / down.sql`**: Skema DDL tabel `entities` lengkap dengan kolom koordinat numerik presisi tinggi dan indeks pencarian.
* **`000002_create_activity_logs_table.up.sql / down.sql`**: Skema DDL tabel `activity_logs` dengan foreign key berelasi ke `entities` serta indeks penanda waktu.
* **`seeds/000001_seed_entities.sql`**: Data bibit (*initial seed*) berisi 8 entitas geospasial realistis area DKI Jakarta untuk pengujian instan.

#### 9. Berkas Konfigurasi & Otomasi Root
* **`Makefile`**: Kumpulan perintah pintas automasi untuk kompilasi biner (`make build`), menjalankan server (`make run`), mengeksekusi test suite (`make test`), linting kode (`make lint`), dan migrasi database (`make migrate-up`).
* **`go.mod` & `go.sum`**: Berkas manifes modul Go yang mencatat seluruh dependensi pustaka pihak ketiga beserta checksum integritasnya.
* **`.env.example` & `.env`**: Berkas definisi variabel lingkungan (port server, kredensial PostgreSQL, dan CORS origins).

---

## ⚙️ Persyaratan Sistem

- **Go:** Versi 1.23 atau lebih baru
- **PostgreSQL:** Versi 14 atau lebih baru
- **Make:** Opsional (untuk mempermudah perintah automasi)

---

## 🚀 Panduan Menjalankan

### 1. Kloning & Masuk ke Direktori Backend
```bash
cd "Backend-Geotrack"
```

### 2. Konfigurasi Lingkungan (`.env`)
Salin file template `.env.example` menjadi `.env` dan sesuaikan nilainya:
```bash
cp .env.example .env
```
Isi konfigurasi standar:
```dotenv
# Application Configuration
APP_NAME=GeoTrack Backend
APP_PORT=8080
APP_ENV=development
CORS_ALLOWED_ORIGINS=http://localhost:8081,http://localhost:5173

# PostgreSQL Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=geotrack_db
DB_SSLMODE=disable
DB_MAX_CONNS=25
DB_MIN_CONNS=5
```

### 3. Menjalankan Server API

**Menggunakan Makefile:**
```bash
make run
```

**Menggunakan Go Langsung:**
```bash
go run cmd/api/main.go
```

Output log terminal akan menunjukkan server aktif:
```text
2026-11-05 11:21:48 INF cmd/api/main.go:34 > Configuration loaded and logger initialized app="GeoTrack Backend" env=development port=8080
2026-11-05 11:21:48 INF cmd/api/main.go:53 > GeoTrack API Server is listening address=:8080
```

### 4. Kompilasi Binary Produksi
```bash
make build
# Binary akan tersimpan di: bin/geotrack-api
```

---

## 🧪 Pengujian & Kualitas Kode

Eksekusi perintah berikut untuk memvalidasi integritas kode:

```bash
# Menjalankan unit tests
make test

# Menjalankan pengujian dengan Concurrency Race Detector
make test-race

# Menghasilkan laporan coverage HTML
make test-cover

# Membersihkan file build dan laporan coverage
make clean
```

---

## 📡 Daftar Endpoint API

Semua endpoint API disajikan dengan basis path `/api/v1`:

| Modul | Method | Endpoint | Deskripsi |
|---|:---:|---|---|
| **Health** | `GET` | `/api/v1/health` | Pemeriksaan status kesehatan server & uptime |
| **Entities** | `GET` | `/api/v1/entities` | Daftar entitas dengan filter (`search`, `type`, `status`) & paginasi |
| **Entities** | `POST` | `/api/v1/entities` | Menambahkan entitas geospasial baru |
| **Entities** | `GET` | `/api/v1/entities/{id}` | Mengambil detail 1 entitas berdasarkan ID |
| **Entities** | `PUT` | `/api/v1/entities/{id}` | Memperbarui data entitas |
| **Entities** | `DELETE` | `/api/v1/entities/{id}` | Menghapus entitas |
| **Activities** | `GET` | `/api/v1/activities` | Daftar riwayat aktivitas / audit trail |
| **Activities** | `GET` | `/api/v1/activities/{id}` | Rincian rekaman aktivitas |
| **Dashboard** | `GET` | `/api/v1/dashboard/stats` | Metrik agregasi operasional untuk dashboard frontend |
| **Swagger UI** | `GET` | `/swagger/index.html` | Dokumentasi interaktif OpenAPI 3.0 |

---

## 🔗 Ekosistem Repositori GeoTrack

| Komponen | Repositori GitHub | Deskripsi & Stack |
|---|---|---|
| **Backend REST API** | [Backend-GeoTrack](https://github.com/NadhiefIlmi/Backend-GeoTrack) | REST API berkinerja tinggi (Golang, Chi Router, PostgreSQL, pgxpool) |
| **Frontend Web App** | [Frontend-GeoTrack](https://github.com/NadhiefIlmi/Frontend-GeoTrack) | Antarmuka interaktif spasial (TanStack Start, React 19, Leaflet, Tailwind CSS) |

