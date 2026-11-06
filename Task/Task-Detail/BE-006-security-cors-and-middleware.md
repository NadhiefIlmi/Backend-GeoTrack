# BE-006 — Security, CORS & Middleware Hardening

> **Task ID:** BE-006  
> **Module:** Security / Middleware  
> **Priority:** High  
> **Dependencies:** BE-001  
> **Estimated Effort:** 1 Day  
> **Status:** COMPLETED (Implemented & Verified on 2026-11-05)

---

## 1. Objective
Membangun lapisan middleware keamanan HTTP berstandar enterprise yang melindungi seluruh endpoint GeoTrack dari serangan umum web (CORS abuse, DDoS/brute force melalui rate limiting, server crash akibat unhandled panic, payload bombardment, dan MIME sniffing), serta menginjeksi traceability context (`Request ID`) pada setiap transaksi API.

---

## 2. Scope of Work
1. **CORS Middleware (`internal/middleware/cors.go`):**
   - Mengizinkan origin resmi frontend GeoTrack: `http://localhost:8081` (Vite SSR port saat ini), `http://localhost:5173`, dan origin produksi yang dapat dikonfigurasi melalui `.env` (`CORS_ALLOWED_ORIGINS`).
   - Mengizinkan HTTP methods: `GET`, `POST`, `PUT`, `DELETE`, `OPTIONS`.
   - Mengizinkan headers: `Content-Type`, `Authorization`, `X-Request-ID`, `Accept`.
   - Mengizinkan preflight caching (`MaxAge: 300`).
2. **Rate Limiting Middleware (`internal/middleware/rate_limiter.go`):**
   - Membatasi kuota request berbasis IP client menggunakan algoritma *Token Bucket* (`golang.org/x/time/rate`).
   - Limit: 60 requests/menit per IP dengan burst capacity 20 requests.
   - Mengembalikan response `HTTP 429 Too Many Requests` saat kuota terlampaui beserta header `Retry-After`.
3. **Panic Recovery Middleware (`internal/middleware/recoverer.go`):**
   - Menangkap unhandled runtime panic secara elegan tanpa mematikan proses server.
   - Mencatat detail stack trace secara aman ke structured logger Zerolog.
   - Mengirim respons `HTTP 500 Internal Server Error` standar ke client tanpa membocorkan trace kode.
4. **Request Tracing Middleware (`internal/middleware/request_id.go`):**
   - Membaca header `X-Request-ID` yang dikirim client atau menghasilkan UUID v4 baru jika tidak ada.
   - Menyematkan request ID ke dalam Go `context.Context` dan header response.
5. **Security Headers & Body Limit Middleware (`internal/middleware/security.go`):**
   - Menyetel headers: `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `X-XSS-Protection: 1; mode=block`, `Referrer-Policy: strict-origin-when-cross-origin`.
   - Membatasi ukuran body request maksimal 1 MB menggunakan `http.MaxBytesReader`.
   - Memvalidasi header `Content-Type: application/json` pada mutasi data (`POST`, `PUT`).

---

## 3. Technical Details

### File & Direktori yang Terlibat
```text
Backend-Geotrack/
└── internal/
    └── middleware/
        ├── cors.go
        ├── rate_limiter.go
        ├── recoverer.go
        ├── request_id.go
        └── security.go
```

### Urutan Middleware Execution Pipeline (Router Chain)
Urutan eksekusi middleware di `internal/delivery/http/router.go` disusun secara deterministik:
1. `RequestID`: Pasang tracking ID pertama kali.
2. `RealIP`: Ekstraksi IP asli client dari `X-Forwarded-For` / `X-Real-IP`.
3. `StructuredLogger`: Logging durasi request, method, path, IP, dan status code.
4. `Recoverer`: Proteksi panic level terluar agar log terformat rapi.
5. `SecurityHeaders`: Set headers keamanan HTTP.
6. `CORS`: Verifikasi origin dan tangani preflight `OPTIONS`.
7. `RateLimiter`: Periksa batas kuota pemanggilan per IP.
8. `BodySizeLimiter`: Batasi 1 MB payload untuk mutasi data.

---

## 4. Middleware Implementation Blueprint

### 1. Rate Limiter Blueprint
```go
package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type clientVisitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	visitors = make(map[string]*clientVisitor)
	mu       sync.Mutex
)

func RateLimit(requestsPerMin int, burst int) func(http.Handler) http.Handler {
	// Background cleanup routine for stale visitor records every 5 minutes
	go cleanupVisitors()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, _, _ := net.SplitHostPort(r.RemoteAddr)
			
			mu.Lock()
			v, exists := visitors[ip]
			if !exists {
				limiter := rate.NewLimiter(rate.Every(time.Minute/time.Duration(requestsPerMin)), burst)
				visitors[ip] = &clientVisitor{limiter: limiter, lastSeen: time.Now()}
				v = visitors[ip]
			}
			v.lastSeen = time.Now()
			mu.Unlock()

			if !v.limiter.Allow() {
				w.Header().Set("Retry-After", "60")
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"success":false,"error":{"code":"RATE_LIMIT_EXCEEDED","message":"Too many requests. Please retry in 60 seconds."}}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
```

### 2. Panic Recoverer Blueprint
```go
package middleware

import (
	"net/http"
	"runtime/debug"

	"geotrack/backend/pkg/logger"
)

func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				reqID := r.Header.Get("X-Request-ID")
				logger.Error().
					Str("request_id", reqID).
					Interface("panic", rvr).
					Bytes("stack", debug.Stack()).
					Msg("Unhandled panic recovered in HTTP pipeline")

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"success":false,"error":{"code":"INTERNAL_SERVER_ERROR","message":"An unexpected internal server error occurred."}}`))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
```

---

## 5. Security Evaluation Matrix

| Risiko Keamanan | Status | Mekanisme Pertahanan |
|---|---|---|
| **CORS Wildcard (`*`) Misconfiguration** | Wajib | Origin dibatasi hanya ke whitelist yang didefinisikan secara eksplisit di environment configuration. |
| **API Denial of Service (DoS)** | Wajib | Token Bucket Rate Limiting 60 req/min/IP + payload hardcap 1 MB. |
| **Server Crash via Nil Pointer** | Wajib | Panic Recoverer menangkap exception dan mengembalikan HTTP 500 JSON. |
| **Information Leakage (Stack Trace)** | Wajib | Response error di-masking; trace hanya dicatat di level log server internal. |
| **MIME Type Confusion Attack** | Wajib | Header `X-Content-Type-Options: nosniff` mencegah browser mengesampingkan deklarasi MIME JSON. |
| **Clickjacking** | Wajib | Header `X-Frame-Options: DENY` mencegah endpoint dibingkai dalam iframe pihak ketiga. |

---

## 6. Performance Requirements
* **In-memory Rate Limiter Garbage Collection:** Visitor map dibersihkan secara berkala setiap 5 menit agar IP address yang sudah tidak aktif dihapus dari RAM dan tidak memicu memory leak.
* **CORS Preflight Cache:** Header `Access-Control-Max-Age: 300` membebaskan browser dari keharusan mengirim request `OPTIONS` berulang untuk setiap panggilan API selama 5 menit.

---

## 7. Acceptance Criteria
- [x] Frontend di `http://localhost:8081` dapat melakukan request `GET`, `POST`, `PUT`, `DELETE` tanpa hambatan CORS.
- [x] Preflight request `OPTIONS` mengembalikan status HTTP 204 atau HTTP 200 dengan header CORS lengkap.
- [x] Request body yang melebihi 1 MB langsung ditolak dengan status HTTP 413 Payload Too Large.
- [x] Request yang melebihi batas kuota (misal > 60 req/menit) menerima HTTP 429 Too Many Requests.
- [x] Ketika terjadi panic di handler, server tetap hidup dan client menerima HTTP 500 JSON standar.
- [x] Header `X-Request-ID` selalu ada pada setiap HTTP response header.

---

## 8. Testing Scenarios

| Skenario | Langkah Pengujian | Status Code | Ekspektasi |
|---|---|---|---|
| **CORS Preflight Valid** | Kirim `OPTIONS /api/v1/entities` dengan `Origin: http://localhost:8081`. | `200 OK` / `204 No Content` | Header `Access-Control-Allow-Origin: http://localhost:8081` tersemat. |
| **CORS Disallowed Origin** | Kirim request dengan `Origin: http://evil-site.com`. | Origin ditolak | Header `Access-Control-Allow-Origin` tidak dikirimkan. |
| **Rate Limit Exceeded** | Kirim 70 request cepat berturut-turut dari 1 IP. | `429 Too Many Requests` | Request ke-61+ ditolak dengan error `RATE_LIMIT_EXCEEDED` dan header `Retry-After`. |
| **Panic Protection** | Panggil endpoint uji yang sengaja melakukan `panic("null pointer simulated")`. | `500 Internal Error` | Server tidak mati, log server mencatat stack trace, response berisi JSON standar. |
| **Large Payload Defense** | Kirim POST request dengan body berukuran 2 MB teks dummy. | `413 Payload Too Large` | Request langsung diputus sebelum dialokasikan ke memori. |
| **Invalid Content-Type** | Kirim POST `/api/v1/entities` dengan `Content-Type: text/plain`. | `415 Unsupported Media Type` | Handler menolak request yang bukan JSON. |

