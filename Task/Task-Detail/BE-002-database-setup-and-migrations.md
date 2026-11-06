# BE-002 — Database Setup, Migrations & Connection Pool

> **Task ID:** BE-002  
> **Module:** Database / Infrastructure  
> **Priority:** High  
> **Dependencies:** BE-001  
> **Estimated Effort:** 1-2 Days  
> **Status:** COMPLETED (Implemented & Verified on 2026-11-05)

---

## 1. Objective
Merancang skema database relasional PostgreSQL untuk entitas GeoTrack (`entities` dan `activity_logs`), membangun mekanisme migrasi terkelola (up/down) menggunakan `golang-migrate`, mengimplementasikan connection pooling berkinerja tinggi menggunakan `jackc/pgx/v5` (`pgxpool`), serta menyediakan initial seeder data untuk mendukung pengujian lokal dan sinkronisasi frontend.

---

## 2. Scope of Work
1. Konfigurasi modul koneksi PostgreSQL `internal/repository/postgres/connection.go` dengan `pgxpool`.
2. Implementasi mekanisme migration runner berbasis CLI atau embedded `golang-migrate/migrate/v4`.
3. Penulisan file migration DDL:
   - `000001_create_entities_table.up.sql` & `.down.sql`
   - `000002_create_activity_logs_table.up.sql` & `.down.sql`
4. Pembuatan database index teroptimasi: B-Tree index pada kolom pencarian (`code`, `name`, `status`, `type`) dan composite index untuk koordinat spasial (`latitude`, `longitude`).
5. Pembuatan migration runner dan seeder data `migrations/seeds/` untuk mengisi mock data entitas awal yang identik dengan mock frontend GeoTrack.
6. Menyediakan fungsi ping/health check database untuk verifikasi koneksi saat aplikasi booting.

---

## 3. Technical Details

### File & Direktori yang Terlibat
```text
Backend-Geotrack/
├── migrations/
│   ├── 000001_create_entities_table.up.sql
│   ├── 000001_create_entities_table.down.sql
│   ├── 000002_create_activity_logs_table.up.sql
│   ├── 000002_create_activity_logs_table.down.sql
│   └── seeds/
│       └── 000001_seed_entities.sql
├── internal/
│   └── repository/
│       └── postgres/
│           ├── connection.go
│           └── migrator.go
├── cmd/
│   └── migrate/
│       └── main.go
└── Makefile
```

### Connection Pool Configuration (`pgxpool.Config`)
* `MaxConns`: 25 (menyesuaikan beban konkurensi produksi).
* `MinConns`: 5 (menjaga koneksi tetap hangat untuk menghindari overhead handshake TCP).
* `MaxConnLifetime`: 1 jam (mencegah memory leak koneksi dan stale socket).
* `MaxConnIdleTime`: 30 menit.
* `HealthCheckPeriod`: 1 menit (memeriksa integritas socket secara periodik).

---

## 4. Database Specification

### 1. Tabel: `entities`
Menyimpan seluruh data aset fisik berbasis geospasial.

| Kolom | Tipe Data | Constraint | Keterangan |
|---|---|---|---|
| `id` | `VARCHAR(36)` / `UUID` | `PRIMARY KEY` | Unique ID entitas (UUID v4) |
| `code` | `VARCHAR(30)` | `NOT NULL, UNIQUE` | Kode unik entitas (e.g. `VH-001`, `IOT-902`) |
| `name` | `VARCHAR(100)` | `NOT NULL` | Nama entitas (3-100 karakter) |
| `type` | `VARCHAR(20)` | `NOT NULL` | ENUM check: `'vehicle'`, `'iot_device'`, `'facility'`, `'infrastructure'` |
| `status` | `VARCHAR(20)` | `NOT NULL, DEFAULT 'active'` | ENUM check: `'active'`, `'inactive'`, `'maintenance'` |
| `latitude` | `DOUBLE PRECISION` | `NOT NULL` | Koordinat lintang (`-90.0` s/d `90.0`) |
| `longitude` | `DOUBLE PRECISION` | `NOT NULL` | Koordinat bujur (`-180.0` s/d `180.0`) |
| `description`| `TEXT` | `NULL` | Catatan/keterangan tambahan |
| `created_at` | `TIMESTAMPTZ` | `NOT NULL, DEFAULT NOW()` | Waktu pembuatan data (UTC) |
| `updated_at` | `TIMESTAMPTZ` | `NOT NULL, DEFAULT NOW()` | Waktu terakhir data diupdate (UTC) |

#### Indexes pada `entities`:
* `idx_entities_code` (`code` UNIQUE)
* `idx_entities_type` (`type`)
* `idx_entities_status` (`status`)
* `idx_entities_coordinates` (`latitude`, `longitude`) - composite index untuk filter koordinat bounding box
* `idx_entities_created_at` (`created_at` DESC) - sorting pagination default

### 2. Tabel: `activity_logs`
Menyimpan rekaman audit trail historis dari seluruh modifikasi data entitas.

| Kolom | Tipe Data | Constraint | Keterangan |
|---|---|---|---|
| `id` | `VARCHAR(36)` / `UUID` | `PRIMARY KEY` | Unique log ID (UUID v4) |
| `entity_id` | `VARCHAR(36)` | `NOT NULL` | Relasi ke entitas (`CASCADE` atau `SET NULL` on delete) |
| `entity_code`| `VARCHAR(30)` | `NOT NULL` | Snapshot kode entitas untuk jejak log |
| `entity_name`| `VARCHAR(100)` | `NOT NULL` | Snapshot nama entitas |
| `action` | `VARCHAR(30)` | `NOT NULL` | ENUM check: `'created'`, `'updated'`, `'status_changed'`, `'deleted'` |
| `details` | `TEXT` | `NOT NULL` | Ringkasan naratif perubahan dalam bahasa manusia |
| `actor` | `VARCHAR(50)` | `NOT NULL, DEFAULT 'System'` | User/sistem pelaku aktivitas |
| `created_at` | `TIMESTAMPTZ` | `NOT NULL, DEFAULT NOW()` | Waktu aktivitas tercatat (UTC) |

#### Foreign Key:
* `fk_activity_entity`: `FOREIGN KEY (entity_id) REFERENCES entities(id) ON DELETE CASCADE`

#### Indexes pada `activity_logs`:
* `idx_activity_entity_id` (`entity_id`)
* `idx_activity_action` (`action`)
* `idx_activity_created_at` (`created_at` DESC) - sorting timeline default

---

## 5. DDL Migrations

### `000001_create_entities_table.up.sql`
```sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS entities (
    id VARCHAR(36) PRIMARY KEY DEFAULT uuid_generate_v4()::text,
    code VARCHAR(30) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_entity_type CHECK (type IN ('vehicle', 'iot_device', 'facility', 'infrastructure')),
    CONSTRAINT chk_entity_status CHECK (status IN ('active', 'inactive', 'maintenance')),
    CONSTRAINT chk_latitude CHECK (latitude >= -90.0 AND latitude <= 90.0),
    CONSTRAINT chk_longitude CHECK (longitude >= -180.0 AND longitude <= 180.0)
);

CREATE INDEX IF NOT EXISTS idx_entities_code ON entities (code);
CREATE INDEX IF NOT EXISTS idx_entities_type ON entities (type);
CREATE INDEX IF NOT EXISTS idx_entities_status ON entities (status);
CREATE INDEX IF NOT EXISTS idx_entities_coordinates ON entities (latitude, longitude);
CREATE INDEX IF NOT EXISTS idx_entities_created_at ON entities (created_at DESC);
```

### `000001_create_entities_table.down.sql`
```sql
DROP TABLE IF EXISTS entities CASCADE;
```

### `000002_create_activity_logs_table.up.sql`
```sql
CREATE TABLE IF NOT EXISTS activity_logs (
    id VARCHAR(36) PRIMARY KEY DEFAULT uuid_generate_v4()::text,
    entity_id VARCHAR(36) NOT NULL,
    entity_code VARCHAR(30) NOT NULL,
    entity_name VARCHAR(100) NOT NULL,
    action VARCHAR(30) NOT NULL,
    details TEXT NOT NULL,
    actor VARCHAR(50) NOT NULL DEFAULT 'System',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_activity_action CHECK (action IN ('created', 'updated', 'status_changed', 'deleted')),
    CONSTRAINT fk_activity_entity FOREIGN KEY (entity_id) REFERENCES entities(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_activity_entity_id ON activity_logs (entity_id);
CREATE INDEX IF NOT EXISTS idx_activity_action ON activity_logs (action);
CREATE INDEX IF NOT EXISTS idx_activity_created_at ON activity_logs (created_at DESC);
```

### `000002_create_activity_logs_table.down.sql`
```sql
DROP TABLE IF EXISTS activity_logs CASCADE;
```

---

## 6. Security Requirements
* **Parameterized Queries Exclusively:** Seluruh interaksi database wajib menggunakan parameter placeholder (`$1, $2, ...`) dari driver `pgx`. Larangan mutlak terhadap string concatenation (`fmt.Sprintf`) dalam menyusun query SQL untuk memusnahkan risiko SQL Injection.
* **Least Privilege DB User:** User database aplikasi tidak boleh menggunakan superuser `postgres`. User aplikasi hanya memiliki izin `SELECT, INSERT, UPDATE, DELETE` pada tabel terkait.
* **TLS / SSL Mode:** Pada environment production, `DB_SSLMODE` wajib diset minimal `require` atau `verify-full` untuk mencegah eavesdropping lalu lintas data database.

---

## 7. Performance Requirements
* **Connection Pooling:** Membatasi koneksi maksimum agar PostgreSQL tidak mengalami resource exhaustion karena thread spawning yang berlebihan.
* **Proper Indexing:** Seluruh filter query (`type`, `status`, `code`) dan pengurutan (`created_at`) dijamin memiliki index pendukung untuk mencegah sequential table scan pada volume data ribuan record.
* **Context Cancellation:** Seluruh eksekusi query (`Query`, `Exec`, `BeginTx`) wajib menyertakan `context.Context` dengan deadline/timeout (default: 5 detik) agar query lambat tidak menggantung koneksi selamanya.

---

## 8. Acceptance Criteria
- [x] Database connection berhasil terhubung saat aplikasi booting dengan log indikator `Database connection pool initialized successfully`.
- [x] Script migrasi SQL dapat dieksekusi naik (`up`) dan turun (`down`) tanpa error sintaks.
- [x] Constraint tabel bekerja dengan baik: menolak `latitude` > 90, menolak invalid `type` atau `status`.
- [x] Unik constraint pada kolom `code` berhasil mencegah duplikasi kode.
- [x] Seeder data terpasang dengan minimal 8 record entitas awal untuk kebutuhan testing dashboard.
- [x] Database health check terintegrasi ke dalam endpoint `GET /api/v1/health`.

---

## 9. Testing Scenarios

| Skenario | Langkah Pengujian | Ekspektasi Hasil |
|---|---|---|
| **Koneksi Pool Valid** | Jalankan aplikasi dengan kredensial database yang benar. | Koneksi sukses, status `UP` di health check. |
| **Koneksi Gagal / Timeout** | Matikan service PostgreSQL lalu jalankan aplikasi. | Aplikasi mencatat log error dan keluar bersih (graceful fail) tanpa panik liar. |
| **Migrate Up & Down** | Jalankan `make migrate-up` diikuti `make migrate-down`. | Seluruh tabel terbuat lalu terhapus bersih tanpa menyisakan orphaned constraints. |
| **Check Constraint Type** | Coba INSERT entity dengan `type = 'airplane'`. | Database melempar error check constraint violation `chk_entity_type`. |
| **Check Constraint Status** | Coba INSERT entity dengan `status = 'broken'`. | Database melempar error check constraint violation `chk_entity_status`. |
| **Check Lat/Lng Range** | Coba INSERT entity dengan `latitude = 120.0`. | Database melempar error check constraint violation `chk_latitude`. |
| **Duplicate Code Unique** | Coba INSERT dua entity dengan kode `VH-001`. | Database melempar duplicate key value violates unique constraint `idx_entities_code`. |
| **Cascade Delete Entity** | Hapus satu record entity yang memiliki log aktivitas. | Seluruh baris di `activity_logs` dengan `entity_id` terkait otomatis terhapus bersih. |

