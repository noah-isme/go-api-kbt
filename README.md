# go-api-kbt

API backend untuk Kedunglo Biking Team (KBT) dengan arsitektur modern, siap diskalakan, dan mengikuti praktik terbaik Go.

## Fitur utama

- **Struktur proyek standar** dengan `cmd/`, `internal/`, dan `pkg/` (jika diperlukan) untuk menjaga batasan dependency.
- **Layered architecture** (domain, repository, service, transport) untuk memudahkan testing dan pergantian infrastruktur.
- **Konfigurasi berbasis environment** menggunakan `.env` dan `envconfig`, dilengkapi profil environment.
- **Middleware komprehensif**: request ID, structured logging (slog), recovery, CORS, dan rate limiting.
- **Validasi request** memakai `go-playground/validator` dengan DTO terpisah dari entity GORM.
- **Observability** via OpenTelemetry stdout exporter (mudah diarahkan ke OTLP collector) dan GORM instrumentation.
- **Docker & Compose** untuk pengembangan lokal (API + PostgreSQL + Redis) dengan healthcheck.
- **CI GitHub Actions** menjalankan lint → test → build → docker build.
- **Response JSON konsisten** dengan envelope `success`, `message`, dan `data`.

## Getting started

### Prasyarat
- Go 1.22+
- Docker & Docker Compose (opsional namun direkomendasikan)
- PostgreSQL & Redis (jika tidak memakai Compose)

### Konfigurasi

1. Salin file contoh environment:
   ```bash
   cp .env.example .env
   ```
2. Sesuaikan nilai variabel sesuai kebutuhan (lihat deskripsi variabel pada `.env.example`).

### Menjalankan secara lokal

```bash
make run
```

Atau gunakan Docker Compose:

```bash
docker-compose up --build
```

Aplikasi akan tersedia di `http://localhost:9090` dengan health check pada `/healthz` dan API versi pertama di `/api/v1`.

### Workflow pengembangan

```bash
make lint   # golangci-lint
make test   # go test ./...
make build  # go build ./...
```

Semua perintah tersebut dijalankan otomatis pada pipeline CI (`.github/workflows/ci.yml`).

## Struktur proyek

```
go-api-kbt/
├── cmd/
│   └── api/          # entrypoint aplikasi
├── internal/
│   ├── app/          # inisialisasi dependency & server
│   ├── config/       # loader konfigurasi (.env + env vars)
│   ├── database/     # koneksi postgres, migrasi, dan seed
│   ├── domain/       # entity domain (user, event, location, medaler, activity)
│   ├── middleware/   # middleware custom (structured logging)
│   ├── observability/# setup OpenTelemetry
│   ├── repository/   # implementasi repositori (GORM)
│   │   ├── event/
│   │   ├── location/
│   │   ├── medaler/
│   │   ├── user/
│   │   └── activity/
│   ├── service/      # business logic/usecase
│   │   ├── event/
│   │   ├── location/
│   │   ├── medaler/
│   │   ├── user/
│   │   └── activity/
│   └── transport/    # HTTP handlers & router (chi)
│       ├── http/
│       │   ├── handler/ # handler HTTP (user, event, location, medaler, activity)
├── migrations/       # migrasi SQL (golang-migrate compatible)
├── docker-compose.yml
├── Dockerfile
├── Makefile
└── README.md
```

## Migrasi & seeding

Skema DB otomatis dibuat saat aplikasi dijalankan menggunakan GORM AutoMigrate untuk entitas `User`, `Event`, `Location`, `Medaler`, dan `Activity`.

Migrasi menggunakan format [golang-migrate](https://github.com/golang-migrate/migrate). Jalankan perintah berikut setelah mengisi variabel environment database:

```bash
make migrate
```

Aktifkan `DB_SEED_ADMIN=true` untuk membuat akun admin default (`admin@example.com`, password `ChangeMe123!`). Pastikan segera mengganti password di produksi.

## Testing

Unit test tersedia untuk service layer dan handler layer (table-driven). Tambahkan integration test menggunakan `httptest` atau Compose sesuai kebutuhan proyek.

## Observability

Secara default tracing dikirim ke stdout dalam format OpenTelemetry sehingga mudah diinspeksi. Sesuaikan konfigurasi OTEL pada `.env` untuk mengarahkannya ke collector (misal OTLP/HTTP atau gRPC).

## Dokumentasi API

Gunakan Swagger/OpenAPI generator seperti [swaggo](https://github.com/swaggo/swag) untuk menghasilkan dokumentasi otomatis dari handler. Struktur DTO dan handler telah disiapkan agar mudah diintegrasikan.

Untuk pengalaman cepat, repositori menyertakan dokumentasi OpenAPI dan Swagger UI yang dapat diakses di runtime pada `/docs/`.

Install Swagger UI assets (lokal) dengan:

```bash
make docs-install
```

Kemudian jalankan aplikasi dan buka `http://localhost:9090/docs/`.

## Zero-Downtime Migration Guidelines

Untuk memastikan migrasi database tanpa downtime, ikuti pedoman berikut:

1.  **Additive Changes Only**: Hindari perubahan skema yang bersifat destruktif (misalnya, menghapus kolom atau tabel) dalam satu langkah migrasi. Lakukan perubahan secara bertahap.
2.  **Backward Compatibility**: Pastikan versi aplikasi yang lebih lama masih dapat bekerja dengan skema database yang baru setelah migrasi diterapkan.
3.  **Two-Phase Deployment**: Untuk perubahan yang lebih kompleks (misalnya, mengubah tipe kolom), lakukan dalam dua fase:
    *   **Fase 1**: Tambahkan kolom baru, migrasikan data dari kolom lama ke kolom baru, dan perbarui aplikasi untuk menulis ke kedua kolom. Deploy aplikasi baru.
    *   **Fase 2**: Hapus kolom lama dan perbarui aplikasi untuk hanya membaca dari kolom baru. Deploy aplikasi baru lagi.
4.  **Testing**: Selalu uji migrasi di lingkungan staging sebelum diterapkan ke produksi.
5.  **Rollback Plan**: Selalu siapkan rencana rollback jika terjadi masalah selama migrasi.

## Deployment

Gunakan Docker image yang dihasilkan dari `Dockerfile` multi-stage. Target hosting yang disarankan: Railway, Render, Fly.io, atau platform container lain. Pastikan environment variable sudah terkonfigurasi dan endpoint health check (`/healthz`) digunakan untuk readiness probe.

## Endpoint Samples

### Health Check

```bash
curl -v http://localhost:9090/healthz
```

### Medaler

#### Create Medaler

```bash
curl -v -X POST -H "Content-Type: application/json" -d '{"name":"John Doe","email":"john.doe@example.com"}' http://localhost:9090/api/v1/medalers
```

#### Get All Medaler

```bash
curl -v http://localhost:9090/api/v1/medalers
```

#### Get Medaler by ID

```bash
curl -v http://localhost:9090/api/v1/medalers/1
```

#### Update Medaler

```bash
curl -v -X PUT -H "Content-Type: application/json" -d '{"name":"Jane Doe","is_active":false}' http://localhost:9090/api/v1/medalers/1
```

#### Delete Medaler

```bash
curl -v -X DELETE http://localhost:9090/api/v1/medalers/1
```

### Activity

#### Get All Activity

```bash
curl -v http://localhost:9090/api/v1/activities
```

#### Get Activity by ID

```bash
curl -v http://localhost:9090/api/v1/activities/1
```
