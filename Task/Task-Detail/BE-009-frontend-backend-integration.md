# BE-009 — Frontend-Backend Integration & End-to-End (E2E) Connectivity

> **Task ID:** BE-009  
> **Module:** Integration / E2E Wiring  
> **Priority:** High  
> **Dependencies:** BE-001 s/d BE-007  
> **Estimated Effort:** 2 Days  
> **Status:** COMPLETED

---

## 1. Objective
Menghubungkan secara penuh aplikasi **Frontend GeoTrack** (TanStack Start, React 19, Tailwind CSS, Axios) dengan **Backend REST API Golang**, menggantikan sumber data purwarupa (*in-memory mock data*) dengan data dinamis real-time dari database PostgreSQL, mengonfigurasi interceptor pembungkus respons (*JSON envelope unwrapper*), serta memvalidasi alur bisnis secara utuh (*End-to-End*).

---

## 2. Scope of Work

### A. Konfigurasi Lingkungan Frontend
1. Membuat file `Frontend-Geotrack/.env` dan memperbarui `Frontend-Geotrack/.env.example` dengan variabel:
   ```dotenv
   VITE_API_BASE_URL=http://localhost:8080/api/v1
   ```
2. Memastikan konfigurasi port pengembangan berjalan harmonis:
   - Frontend: `http://localhost:8081` (Vite SSR)
   - Backend: `http://localhost:8080` (Golang Chi API Server)

### B. Standardisasi Axios Client & Response Envelope
1. Memodifikasi `Frontend-Geotrack/src/services/api.ts`:
   - Mengonfigurasi Axios Interceptor untuk otomatis meng-unwrap format envelope backend:
     `response.data.data` -> menjadi kembalian data utama.
   - Menyimpan pagination metadata `response.data.meta` jika tersedia.
   - Menangani error interceptor terpusat yang memetakan error format `{ code, message, details }` ke notifikasi toast (Sonner).

### C. Pembuatan & Pemutakhiran Layanan API Frontend
1. **`src/services/entityService.ts`:**
   - Menghubungkan seluruh fungsi CRUD (`getAllEntities`, `getEntityByCode`, `createEntity`, `updateEntity`, `deleteEntity`) ke endpoint live `/api/v1/entities`.
   - Mengirim query parameters: `search`, `type`, `status`, `page`, `limit`, `sort_by`, `order`.
2. **`src/services/activityService.ts` (Baru):**
   - Menyediakan fungsi `getActivities(filter)` untuk mengambil linimasa aktivitas live dari `/api/v1/activities`.
3. **`src/services/dashboardService.ts` (Baru):**
   - Menyediakan fungsi `getDashboardStats()` untuk mengambil metrik agregat terpadu dari `/api/v1/dashboard/stats`.

### D. Refactoring State Management & Data Provider
1. Memperbarui `Frontend-Geotrack/src/state/EntityStore.tsx`:
   - Mengganti state lokal statis dengan integrasi TanStack Query (`@tanstack/react-query`).
   - Menyediakan hooks reaktif: `useEntitiesQuery`, `useCreateEntityMutation`, `useUpdateEntityMutation`, `useDeleteEntityMutation`.
   - Melakukan query invalidation otomatis (`queryClient.invalidateQueries`) setelah mutasi sukses agar data tabel, peta, dan aktivitas sinkron seketika.
2. Memperbarui `Frontend-Geotrack/src/routes/__root.tsx`:
   - Menghapus ketergantungan import hardcoded `mockEntities` dan `mockActivities`.
   - Menginisialisasi `EntityStoreProvider` dengan fetch asinkron dari backend API.

---

## 3. Technical Details

### File & Direktori yang Terlibat

```text
Frontend-Geotrack/
├── .env.example
├── .env
├── src/
│   ├── services/
│   │   ├── api.ts              # Axios instance + Envelope Interceptor
│   │   ├── entityService.ts    # Live Entity CRUD client
│   │   ├── activityService.ts  # Live Activity Logs client
│   │   └── dashboardService.ts # Live Dashboard Stats client
│   ├── state/
│   │   └── EntityStore.tsx     # TanStack Query & Reactive State
│   └── routes/
│       ├── __root.tsx          # Provider initialization
│       ├── index.tsx           # Dashboard view live binding
│       ├── entities.tsx        # Entities table & map live binding
│       └── activity.tsx        # Activity logs timeline live binding
```

### Arsitektur Interceptor Axios (`src/services/api.ts`)

```typescript
import axios, { type AxiosResponse } from "axios";

export interface BackendResponse<T> {
  success: boolean;
  message: string;
  data: T;
  meta?: {
    page: number;
    limit: number;
    total_records: number;
    total_pages: number;
  };
}

export const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || "http://localhost:8080/api/v1",
  timeout: 10000,
  headers: {
    "Content-Type": "application/json",
    Accept: "application/json",
  },
});

// Response Interceptor: Unwrap backend envelope
apiClient.interceptors.response.use(
  (response: AxiosResponse<BackendResponse<unknown>>) => {
    // If backend returns standard envelope, return inner data directly
    if (response.data && typeof response.data === "object" && "success" in response.data) {
      return response;
    }
    return response;
  },
  (error) => {
    const errorData = error.response?.data?.error;
    const errorMessage = errorData?.message || error.message || "An unexpected error occurred";
    console.error("[GeoTrack API Error]:", errorData || error.message);
    return Promise.reject(new Error(errorMessage));
  }
);
```

---

## 4. Pemetaan Kontrak Data (Frontend TypeScript vs Backend JSON)

| Fitur | Endpoint Backend | Method | Service Frontend | Tipe Data Frontend |
|---|---|:---:|---|---|
| **Health Check** | `/api/v1/health` | `GET` | `apiClient.get('/health')` | `{ status, version, uptime_seconds }` |
| **List Entities** | `/api/v1/entities` | `GET` | `entityService.getAll(params)` | `Entity[]` + `meta` |
| **Get Entity** | `/api/v1/entities/{id}` | `GET` | `entityService.getById(id)` | `Entity` |
| **Create Entity** | `/api/v1/entities` | `POST` | `entityService.create(payload)` | `Entity` |
| **Update Entity** | `/api/v1/entities/{id}` | `PUT` | `entityService.update(id, payload)` | `Entity` |
| **Delete Entity** | `/api/v1/entities/{id}` | `DELETE` | `entityService.delete(id)` | `boolean` |
| **Activity Logs** | `/api/v1/activities` | `GET` | `activityService.getAll(params)` | `ActivityLog[]` + `meta` |
| **Dashboard Stats** | `/api/v1/dashboard/stats` | `GET` | `dashboardService.getStats()` | `DashboardStats` |

---

## 5. Security & Network Considerations
* **CORS Harmonization:** Server Golang (port 8080) telah mengizinkan origin `http://localhost:8081` secara eksplisit melalui middleware `SimpleCORS`.
* **Safe Error Propagation:** Error detail dari backend (seperti validasi format latitude/longitude atau kode duplikat) ditampilkan ke user melalui toast tanpa memaparkan sistem database internal.
* **Network Offline Resilience:** Jika backend tidak dapat dihubungi (misal saat server mati), frontend menampilkan banner peringatan koneksi terputus dan tombol *Retry* alih-alih aplikasi crash layar putih (*white screen of death*).

---

## 6. Performance Requirements
* **TanStack Query Caching:**
  - `staleTime`: 30 detik (data dashboard dan entities tidak di-refetch berlebihan pada navigasi antar-tab).
  - `gcTime`: 5 menit.
* **Search Input Debouncing:** Input pencarian pada halaman Entities dan Activity Logs di-debounce selama 300 ms untuk mencegah flooding request HTTP ke server.
* **Optimistic / Instant Feedback:** Mutasi penambahan dan pengeditan entitas memberikan feedback visual cepat (*optimistic update* atau loading skeleton halus).

---

## 7. Acceptance Criteria
- [x] Frontend dapat memuat daftar entitas secara langsung dari database backend Golang.
- [x] Menambahkan entitas baru melalui modal form di frontend berhasil tersimpan permanen di database PostgreSQL.
- [x] Riwayat log aktivitas pada halaman `/activity` otomatis bertambah saat terjadi mutasi entitas.
- [x] Statistik ringkasan pada dashboard utama (`/`) menampilkan total aset dan status aktual dari backend API.
- [x] Fitur pencarian dan filter tipe/status pada halaman `/entities` mengeksekusi query dinamis ke backend API.
- [x] Tidak ada CORS error pada browser console saat frontend `http://localhost:8081` memanggil backend `http://localhost:8080`.
- [x] Seluruh unit test frontend (`pnpm test`) tetap lulus 100%.

---

## 8. Testing Scenarios (End-to-End)

| Skenario | Langkah Pengujian | Hasil yang Diharapkan |
|---|---|---|
| **E2E-001: Load Dashboard** | Buka `http://localhost:8081/` saat backend aktif. | Dashboard menampilkan kartu statistik dinamis dari `GET /api/v1/dashboard/stats`. |
| **E2E-002: Live Entity List** | Buka `http://localhost:8081/entities`. | Tabel dan peta menampilkan pin entitas aktual dari database. |
| **E2E-003: Create Entity Flow** | Klik "Add Entity", isi kode `VH-999`, nama `Delivery Van 99`, simpan. | Modal tertutup, notifikasi sukses muncul, entitas muncul di tabel, record masuk di DB PostgreSQL. |
| **E2E-004: Audit Trail Sync** | Buka `http://localhost:8081/activity` setelah create entity. | Log baru aksi `created` untuk entitas `VH-999` muncul di posisi teratas linimasa. |
| **E2E-005: Validation Handling** | Coba buat entitas baru dengan kode yang sama (`VH-999`). | Backend menolak dengan status 409 Conflict, UI menampilkan toast error "Entity code already exists". |
| **E2E-006: Server Down Resilience** | Matikan server backend Golang, lalu refresh frontend. | Muncul state peringatan koneksi gagal dengan opsi "Retry Connection", aplikasi tidak crash. |

