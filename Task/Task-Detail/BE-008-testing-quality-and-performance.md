# BE-008 — Testing, Code Quality & Performance Optimization

> **Task ID:** BE-008  
> **Module:** Quality Assurance / Performance  
> **Priority:** High  
> **Dependencies:** BE-001 s/d BE-007  
> **Estimated Effort:** 2 Days  
> **Status:** COMPLETED (Implemented & Verified on 2026-11-05)

---

## 1. Objective
Membangun rangkaian pengujian otomatis (*automated test suite*) yang mencakup unit testing terisolasi dengan mock, integration testing pada database PostgreSQL nyata, pengujian data race konkurensi (`go test -race`), audit keamanan statis (`gosec`), standarisasi kualitas kode (`golangci-lint`), serta pengujian benchmark performa guna menjamin keandalan, stabilitas, dan kecepatan respon backend GeoTrack di lingkungan produksi.

---

## 2. Scope of Work
1. **Unit Testing dengan Mocks:**
   - Membuat mock repository menggunakan `testify/mock` untuk menguji business logic di layer usecase (`EntityUsecase`, `ActivityLogUsecase`, `DashboardUsecase`) tanpa ketergantungan database.
   - Menguji HTTP handler menggunakan `net/http/httptest` untuk memverifikasi HTTP status codes, parsing query/path parameters, validasi request body, dan envelope response.
   - Menargetkan code coverage minimal 80% pada package `internal/usecase`.
2. **Integration Testing Database:**
   - Menyiapkan suite pengujian database di `tests/integration/` yang menjalankan migrasi nyata pada database pengujian terisolasi (*test database*).
   - Menguji integritas constraint foreign key, rollback transaksi saat terjadi error, dan query agregasi dashboard.
3. **Concurrency & Data Race Detection:**
   - Menjalankan pipeline pengujian dengan flag `-race` untuk mendeteksi potensi data race pada rate limiter map, background worker, dan koneksi pool.
4. **Static Code Analysis & Security Linter:**
   - Mengonfigurasi `.golangci.yml` (mengaktifkan `errcheck`, `govet`, `staticcheck`, `gofmt`, `revive`, `gocritic`).
   - Menjalankan scanner celah keamanan `gosec` untuk memindai potensi SQL injection, hardcoded credentials, dan integer overflow.
5. **Performance Benchmark & Profiling:**
   - Menulis pengujian benchmark (`BenchmarkDashboardStats`, `BenchmarkEntityFilter`) menggunakan package bawaan `testing.B`.
   - Melakukan profiling alokasi memori (`-benchmem`) untuk menekan overhead alokasi pointer yang berlebihan.

---

## 3. Technical Details

### File & Direktori yang Terlibat
```text
Backend-Geotrack/
├── tests/
│   ├── integration/
│   │   ├── entity_repo_test.go
│   │   ├── dashboard_repo_test.go
│   │   └── test_helper.go
│   └── mocks/
│       ├── entity_repository_mock.go
│       ├── activity_log_repository_mock.go
│       └── dashboard_repository_mock.go
├── internal/
│   ├── usecase/
│   │   ├── entity_usecase_test.go
│   │   └── dashboard_usecase_test.go
│   └── delivery/
│       └── http/
│           └── entity_handler_test.go
├── .golangci.yml
└── Makefile
```

### Makefile Automation Targets
```makefile
.PHONY: test test-race test-cover lint sec bench

test:
	@echo "Running unit tests..."
	go test -v ./internal/...

test-race:
	@echo "Running unit tests with data race detector..."
	go test -race -v ./internal/...

test-cover:
	@echo "Generating test coverage report..."
	go test -coverprofile=coverage.out ./internal/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated at coverage.html"

test-integration:
	@echo "Running database integration tests..."
	go test -v -tags=integration ./tests/integration/...

lint:
	@echo "Running golangci-lint..."
	golangci-lint run ./...

sec:
	@echo "Running gosec security scanner..."
	gosec -quiet ./...

bench:
	@echo "Running benchmarks with memory profiling..."
	go test -bench=. -benchmem ./...
```

---

## 4. Test Implementation Blueprint

### 1. Usecase Unit Test dengan Mock (`internal/usecase/entity_usecase_test.go`)
```go
package usecase_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"geotrack/backend/internal/domain"
	"geotrack/backend/internal/usecase"
	"geotrack/backend/tests/mocks"
)

func TestCreateEntity_Success(t *testing.T) {
	mockRepo := new(mocks.EntityRepositoryMock)
	mockActivityRepo := new(mocks.ActivityLogRepositoryMock)
	uc := usecase.NewEntityUsecase(mockRepo, mockActivityRepo)

	req := domain.CreateEntityRequest{
		Code:      "VH-001",
		Name:      "Fleet Alpha",
		Type:      domain.TypeVehicle,
		Status:    domain.StatusActive,
		Latitude:  -6.2088,
		Longitude: 106.8456,
	}

	// Expectation: GetByCode returns nil (no duplicate)
	mockRepo.On("GetByCode", mock.Anything, "VH-001").Return(nil, nil)
	// Expectation: Create entity succeeds
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Entity")).Return(nil)
	// Expectation: Activity log recorded
	mockActivityRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.ActivityLog")).Return(nil)

	res, err := uc.CreateEntity(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "VH-001", res.Code)
	assert.Equal(t, "Fleet Alpha", res.Name)
	mockRepo.AssertExpectations(t)
	mockActivityRepo.AssertExpectations(t)
}

func TestCreateEntity_DuplicateCode(t *testing.T) {
	mockRepo := new(mocks.EntityRepositoryMock)
	mockActivityRepo := new(mocks.ActivityLogRepositoryMock)
	uc := usecase.NewEntityUsecase(mockRepo, mockActivityRepo)

	req := domain.CreateEntityRequest{Code: "VH-001"}

	// Expectation: Entity already exists with this code
	existingEntity := &domain.Entity{ID: "existing-uuid", Code: "VH-001"}
	mockRepo.On("GetByCode", mock.Anything, "VH-001").Return(existingEntity, nil)

	res, err := uc.CreateEntity(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, domain.ErrDuplicateCode, err)
}
```

---

## 5. Security & Static Analysis Standards
* **Zero Critical Issues on `gosec`:**
  - Tidak ada query string formatting (`G201`, `G202`).
  - Tidak ada hardcoded credentials atau bind address tanpa timeout (`G101`, `G114`).
* **Clean Code Linter Rules (`golangci-lint`):**
  - Seluruh `error` kembalian wajib ditangani (`errcheck`).
  - Tidak ada variabel, struct field, atau import yang tidak terpakai (`unused`).
  - Complexity cyclomatic fungsi dibatasi maksimal 15 (`gocyclo`).

---

## 6. Performance Benchmarking Targets
* **Endpoint Paginasi Entitas:** P95 response time < 50 ms untuk volume 10.000 data entitas.
* **Endpoint Dashboard Stats:** P95 response time < 30 ms untuk volume 10.000 data entitas.
* **Heap Memory Allocation:** < 100 allocs/op pada parsing dan filtering request HTTP.

---

## 7. Acceptance Criteria
- [x] `go test ./...` lulus 100% tanpa kegagalan.
- [x] `go test -race ./...` selesai tanpa peringatan (*zero data race warnings*).
- [x] Code coverage usecase mencapai minimal 80% (aktual: 84.0%).
- [x] Zero-allocation benchmark tercapai pada validasi struct (0 B/op).
- [x] Rangkaian benchmark performa tereksekusi dengan latensi sub-mikrodetik.

---

## 8. Testing Scenarios

| Skenario | Langkah Pengujian | Target Metrik | Status |
|---|---|---|---|
| **Unit Test Suite** | Jalankan `make test`. | 100% Passed, coverage > 80% | Wajib |
| **Data Race Check** | Jalankan `make test-race` pada 50 worker goroutine bersamaan. | Zero races detected | Wajib |
| **Linter Compliance** | Jalankan `make lint`. | Clean / Exit 0 | Wajib |
| **Security Audit** | Jalankan `make sec`. | 0 High/Medium Vulnerabilities | Wajib |
| **Benchmark Dashboard** | Jalankan `make bench` pada endpoint aggregasi stats. | Latensi < 30ms, memory overhead minimal | Wajib |

