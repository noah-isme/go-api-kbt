# go-api-kbt

API backend untuk Kedunglo Biking Team (KBT) dengan arsitektur modern, siap diskalakan, dan mengikuti praktik terbaik Go.

## Fitur utama

- **Struktur proyek standar** dengan `cmd/`, `internal/`, dan `pkg/` (jika diperlukan) untuk menjaga batasan dependency.
- **Layered architecture** (domain, repository, service, transport) untuk memudahkan testing dan pergantian infrastruktur.
- **Konfigurasi berbasis environment** menggunakan `.env` dan `envconfig`, dilengkapi profil environment.
- **Middleware komprehensif**: request ID, structured logging (slog), recovery, CORS, dan rate limiting.
- **Validasi request** memakai `go-playground/validator` dengan DTO terpisah dari entity GORM.
- **Observability** via OpenTelemetry stdout exporter (mudah diarahkan ke OTLP collector).
- **Docker & Compose** untuk pengembangan lokal (API + PostgreSQL + Redis).
- **CI GitHub Actions** menjalankan lint → test → build → docker build.

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
│   ├── domain/       # entity domain (user, event, location)
│   ├── middleware/   # middleware custom (structured logging)
│   ├── observability/# setup OpenTelemetry
│   ├── repository/   # implementasi repositori (GORM)
│   ├── service/      # business logic/usecase
│   └── transport/    # HTTP handlers & router (chi)
├── migrations/       # migrasi SQL (golang-migrate compatible)
├── docker-compose.yml
├── Dockerfile
├── Makefile
└── README.md
```

## Migrasi & seeding

Migrasi menggunakan format [golang-migrate](https://github.com/golang-migrate/migrate). Jalankan perintah berikut setelah mengisi variabel environment database:

```bash
make migrate
```

Aktifkan `DB_SEED_ADMIN=true` untuk membuat akun admin default (`admin@example.com`, password `ChangeMe123!`). Pastikan segera mengganti password di produksi.

## Testing

Unit test tersedia untuk service layer. Tambahkan integration test menggunakan `httptest` atau Compose sesuai kebutuhan proyek.

## Observability

Secara default tracing dikirim ke stdout dalam format OpenTelemetry sehingga mudah diinspeksi. Sesuaikan konfigurasi OTEL pada `.env` untuk mengarahkannya ke collector (misal OTLP/HTTP atau gRPC).

## Dokumentasi API

Gunakan Swagger/OpenAPI generator seperti [swaggo](https://github.com/swaggo/swag) untuk menghasilkan dokumentasi otomatis dari handler. Struktur DTO dan handler telah disiapkan agar mudah diintegrasikan.

## Deployment

Gunakan Docker image yang dihasilkan dari `Dockerfile` multi-stage. Target hosting yang disarankan: Railway, Render, Fly.io, atau platform container lain. Pastikan environment variable sudah terkonfigurasi dan endpoint health check (`/healthz`) digunakan untuk readiness probe.
