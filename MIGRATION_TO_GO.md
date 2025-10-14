
> 📘 **“Dokumentasi Migrasi dari Python (Django) ke Go (Golang)”**

Dokumen ini mempertahankan gaya dan kelengkapan struktur file aslinya, tetapi diarahkan untuk **migrasi backend Python → Go**, sesuai deskripsi yang kamu berikan:

---

# 🐍➡️🐹 Dokumentasi Migrasi dari Python (Django) ke Golang

Panduan lengkap untuk migrasi **KBT App Backend** dari **Django (Python)** ke **Golang** dengan arsitektur modern dan maintainable.

---

## 📋 Daftar Isi

1. [Gambaran Umum](#gambaran-umum)
2. [Target Arsitektur Go](#target-arsitektur-go)
3. [Strategi Migrasi](#strategi-migrasi)
4. [Struktur Proyek Go](#struktur-proyek-go)
5. [Konfigurasi & Environment](#konfigurasi--environment)
6. [Arsitektur Layered (Domain–Repository–Service–Transport)](#arsitektur-layered)
7. [Middleware & Infrastruktur](#middleware--infrastruktur)
8. [Validasi & DTO](#validasi--dto)
9. [Observability (OpenTelemetry)](#observability-opentelemetry)
10. [Docker & Compose Setup](#docker--compose-setup)
11. [Continuous Integration (GitHub Actions)](#continuous-integration-github-actions)
12. [Checklist Migrasi](#checklist-migrasi)

---

## 🎯 Gambaran Umum

### Arsitektur Saat Ini

* **Backend**: Django 4.2.9 (Python)
* **Frontend**: React.js (hasil migrasi sebelumnya)
* **Database**: PostgreSQL
* **Cache/Queue**: Redis
* **Deployment**: Docker + Nginx reverse proxy

### Tujuan Migrasi

Migrasi backend dari **Python Django** ke **Go** dengan tujuan:

* Performa dan efisiensi resource yang lebih tinggi
* Konkurensi native (goroutines) untuk beban tinggi (tracking, streaming, analytics)
* Arsitektur modular, mudah di-maintain dan diuji
* Observability bawaan untuk tracing dan metrics

---

## 🧱 Target Arsitektur Go

### Prinsip

* **Layered architecture**: `domain` → `repository` → `service` → `transport`
* **Dependency boundaries** terjaga (`internal` dan `pkg`)
* **Configuration by environment**
* **Middleware yang komprehensif**
* **Observability siap produksi**

---

## 🚀 Strategi Migrasi

| Fase       | Deskripsi                                           | Durasi     |
| ---------- | --------------------------------------------------- | ---------- |
| **Fase 1** | Setup project Go & struktur dasar                   | 1 minggu   |
| **Fase 2** | Implement domain & repository layer                 | 1–2 minggu |
| **Fase 3** | Porting logic dari Django ke Go service layer       | 2–3 minggu |
| **Fase 4** | Tambah transport layer (REST API)                   | 1 minggu   |
| **Fase 5** | Integrasi middleware, observability, Docker, dan CI | 1–2 minggu |

Migrasi dilakukan **inkremental**, mulai dari modul sederhana (misal: `Medaler`, `Activity`) sebelum modul kompleks.

---

## 🗂️ Struktur Proyek Go

```bash
kbt-go/
├── cmd/
│   └── api/
│       ├── main.go           # Entry point
│       └── config.go         # Load environment configs
├── internal/
│   ├── domain/               # Entity models (pure Go)
│   ├── repository/           # Database access (GORM)
│   ├── service/              # Business logic
│   ├── transport/
│   │   ├── http/             # REST API handlers
│   │   └── middleware/       # Logging, recovery, etc.
│   ├── config/               # Env loader & struct
│   └── observability/        # OpenTelemetry setup
├── pkg/
│   ├── validator/            # Request validation utilities
│   └── logger/               # Structured logging (slog)
├── .env
├── Dockerfile
├── docker-compose.yml
├── go.mod
└── go.sum
```

---

## ⚙️ Konfigurasi & Environment

Gunakan package [`kelseyhightower/envconfig`](https://github.com/kelseyhightower/envconfig).

### `.env`

```bash
APP_NAME=kbt-api
APP_ENV=development
SERVER_PORT=8080
DATABASE_URL=postgres://user:password@db:5432/kbt_db?sslmode=disable
REDIS_URL=redis://redis:6379
RATE_LIMIT=100
```

### `internal/config/config.go`

```go
package config

import (
	"log"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	AppName     string `envconfig:"APP_NAME" default:"kbt-api"`
	AppEnv      string `envconfig:"APP_ENV" default:"development"`
	ServerPort  int    `envconfig:"SERVER_PORT" default:"8080"`
	DatabaseURL string `envconfig:"DATABASE_URL" required:"true"`
	RedisURL    string `envconfig:"REDIS_URL"`
	RateLimit   int    `envconfig:"RATE_LIMIT" default:"100"`
}

func Load() *Config {
	var c Config
	if err := envconfig.Process("", &c); err != nil {
		log.Fatalf("Failed to load env: %v", err)
	}
	return &c
}
```

---

## 🧩 Arsitektur Layered

### 1. Domain Layer (`internal/domain`)

```go
type Medaler struct {
	ID       uint
	Name     string
	Email    string
	IsActive bool
}
```

### 2. Repository Layer (`internal/repository`)

Gunakan **GORM** untuk ORM.

```go
type MedalerRepository interface {
	FindAll() ([]domain.Medaler, error)
	FindByID(id uint) (*domain.Medaler, error)
	Create(m *domain.Medaler) error
}
```

Implementasi PostgreSQL:

```go
type medalerRepo struct { db *gorm.DB }

func (r *medalerRepo) FindAll() ([]domain.Medaler, error) {
	var medalers []domain.Medaler
	err := r.db.Find(&medalers).Error
	return medalers, err
}
```

### 3. Service Layer (`internal/service`)

```go
type MedalerService struct {
	repo repository.MedalerRepository
}

func (s *MedalerService) ListMedalers() ([]domain.Medaler, error) {
	return s.repo.FindAll()
}
```

### 4. Transport Layer (`internal/transport/http`)

Gunakan **Chi Router** untuk handler.

```go
r := chi.NewRouter()
r.Use(middleware.RequestID)
r.Use(middleware.Logger)
r.Use(middleware.Recoverer)
r.Use(cors.Handler(cors.Options{
    AllowedOrigins: []string{"*"},
    AllowedMethods: []string{"GET","POST","PUT","DELETE"},
}))
```

Handler:

```go
func (h *Handler) ListMedalers(w http.ResponseWriter, r *http.Request) {
	medalers, err := h.svc.ListMedalers()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	json.NewEncoder(w).Encode(medalers)
}
```

---

## 🛠️ Middleware & Infrastruktur

### Middleware yang Digunakan

* **Request ID** → `chi/middleware.RequestID`
* **Structured Logging** → `log/slog`
* **Recovery** → `chi/middleware.Recoverer`
* **CORS** → `github.com/go-chi/cors`
* **Rate Limiting** → `golang.org/x/time/rate`

---

## ✅ Validasi & DTO

Pisahkan **DTO (request/response)** dari entity.

```go
type CreateMedalerRequest struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}
```

Gunakan validator:

```go
validate := validator.New()
if err := validate.Struct(req); err != nil {
	http.Error(w, err.Error(), 400)
	return
}
```

---

## 🔭 Observability (OpenTelemetry)

Gunakan exporter stdout agar mudah diarahkan ke collector.

```go
import "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"

exp, _ := stdouttrace.New(stdouttrace.WithPrettyPrint())
tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(exp))
otel.SetTracerProvider(tp)
```

---

## 🐳 Docker & Compose Setup

### `Dockerfile`

```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod tidy && go build -o api ./cmd/api

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/api .
CMD ["./api"]
```

### `docker-compose.yml`

```yaml
version: "3.9"
services:
  api:
    build: .
    ports:
      - "8080:8080"
    env_file: .env
    depends_on:
      - db
      - redis
  db:
    image: postgres:15
    environment:
      POSTGRES_DB: kbt_db
      POSTGRES_USER: user
      POSTGRES_PASSWORD: password
    ports:
      - "5432:5432"
  redis:
    image: redis:7
    ports:
      - "6379:6379"
```

---

## ⚙️ Continuous Integration (GitHub Actions)

### `.github/workflows/ci.yml`

```yaml
name: Go CI

on: [push, pull_request]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: 1.22
      - name: Lint
        run: go vet ./...
      - name: Test
        run: go test ./... -v
      - name: Build
        run: go build -v ./cmd/api
      - name: Docker Build
        run: docker build -t kbt-api .
```

---

## 🧾 Checklist Migrasi

### Persiapan

* [ ] Audit logic Django (models, views, serializers)
* [ ] Tentukan domain entities
* [ ] Siapkan .env & struktur project Go

### Implementasi

* [ ] Domain layer selesai
* [ ] Repository layer (GORM + PostgreSQL)
* [ ] Service layer (logic business)
* [ ] HTTP transport (Chi)
* [ ] Middleware lengkap

### Infrastruktur

* [ ] OpenTelemetry setup
* [ ] Docker & Compose berjalan
* [ ] GitHub Actions pipeline jalan

### Uji Coba

* [ ] Endpoint CRUD medaler
* [ ] Auth JWT
* [ ] Observability log/traces muncul

---

## 📚 Referensi

* [Go Project Layout](https://github.com/golang-standards/project-layout)
* [Chi Router](https://github.com/go-chi/chi)
* [GORM ORM](https://gorm.io)
* [Envconfig](https://github.com/kelseyhightower/envconfig)
* [Validator](https://github.com/go-playground/validator)
* [OpenTelemetry-Go](https://opentelemetry.io/docs/instrumentation/go/)

---
