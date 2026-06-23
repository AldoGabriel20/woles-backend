# Woles Backend

REST API backend untuk **Woles** — asisten administrasi kehidupan berbasis WhatsApp.

Dibangun dengan **Go 1.25**, **Fiber v2**, arsitektur Hexagonal (Ports & Adapters).

---

## Prasyarat

| Tool | Versi minimum | Keterangan |
|---|---|---|
| Go | 1.25 | [golang.org/dl](https://golang.org/dl/) |
| Docker | 24+ | Menjalankan Postgres, Redis, RabbitMQ, MinIO |
| Docker Compose | v2 | Sudah termasuk di Docker Desktop |
| `make` | — | Tersedia di Linux/macOS |
| `openssl` | — | Untuk generate RSA key pair |

---

## Quickstart (Infrastruktur via Docker + Server via Go)

Cara paling umum untuk development: jalankan infrastruktur via Docker, lalu server Go langsung di host.

### 1. Clone dan install dependencies

```bash
git clone https://github.com/AldoGabriel20/woles-backend.git
cd woles-backend
go mod download
```

### 2. Salin dan isi file environment

```bash
cp .env.example .env
```

Edit `.env` sesuai kebutuhan. Untuk local development, nilai default sudah cukup kecuali beberapa field di bawah ini yang **wajib diisi**:

| Variable | Yang perlu diisi |
|---|---|
| `JWT_PRIVATE_KEY_PATH` | Path ke file RSA private key (lihat langkah 3) |
| `JWT_PUBLIC_KEY_PATH` | Path ke file RSA public key (lihat langkah 3) |
| `WHATSAPP_WEBHOOK_SECRET` | Secret dari provider WhatsApp (bisa dikosongkan saat dev) |
| `AI_API_KEY` | API key OpenAI / provider AI yang digunakan |
| `PAYMENT_SECRET_KEY` | Secret key Midtrans / payment provider |

### 3. Generate RSA key pair untuk JWT (RS256)

```bash
mkdir -p keys
openssl genrsa -out keys/private.pem 2048
openssl rsa -in keys/private.pem -pubout -out keys/public.pem
```

Pastikan path di `.env` sudah sesuai:

```env
JWT_PRIVATE_KEY_PATH=./keys/private.pem
JWT_PUBLIC_KEY_PATH=./keys/public.pem
```

> **Catatan keamanan:** Folder `keys/` sudah ada di `.gitignore`. Jangan pernah commit private key ke repository.

### 4. Jalankan infrastruktur (Postgres, Redis, RabbitMQ, MinIO)

```bash
docker compose up postgres redis rabbitmq minio -d
```

Tunggu hingga semua service healthy (sekitar 15–30 detik):

```bash
docker compose ps
```

Output yang diharapkan:

```
NAME                STATUS
woles-postgres      Up (healthy)
woles-redis         Up (healthy)
woles-rabbitmq      Up (healthy)
woles-minio         Up
```

### 5. Jalankan database migration

```bash
make migrate-up
```

Perintah ini menjalankan semua file SQL di `internal/migration/` secara berurutan menggunakan [goose](https://github.com/pressly/goose).

### 6. Jalankan API server

```bash
make run
```

Server berjalan di `http://localhost:8080`.

Untuk memverifikasi:

```bash
curl http://localhost:8080/api/v1/auth/me
# → 401 Unauthorized (normal — belum login)
```

---

## Menjalankan Semua Service via Docker Compose (Full Stack)

Cocok untuk staging atau production-like environment.

```bash
# Build image terlebih dahulu
docker compose build

# Jalankan semua service
docker compose up -d

# Lihat log
docker compose logs -f api
```

Saat menggunakan cara ini, environment variable koneksi sudah otomatis di-override oleh `docker-compose.yml` (hostname service Docker, bukan `localhost`).

---

## Perintah Make yang Tersedia

```bash
make run           # Jalankan API server (go run)
make build         # Compile binary ke bin/woles-backend
make test          # Jalankan semua unit test
make migrate-up    # Apply semua migration yang pending
make migrate-down  # Rollback migration terakhir
make lint          # Jalankan golangci-lint
make generate-mock # Generate mock untuk semua port interface
```

---

## Menjalankan Worker dan Scheduler (Opsional)

Worker dan scheduler adalah binary terpisah. Untuk development, bisa dijalankan di terminal berbeda:

```bash
# Intent worker — memproses pesan masuk WhatsApp
go run ./cmd/intent_worker/main.go

# Notification worker — mengirim notifikasi via WhatsApp
go run ./cmd/notification_worker/main.go

# Scheduler — men-claim notifikasi yang jatuh tempo setiap 60 detik
go run ./cmd/scheduler/main.go
```

---

## Menjalankan Test

### Unit test (tidak butuh infrastruktur)

```bash
go test ./tests/unit/... -v
```

### Integration test (butuh Postgres dan Redis)

Pastikan infrastruktur sudah berjalan (langkah 4), lalu:

```bash
TEST_DATABASE_URL="postgres://woles:woles@localhost:5432/woles?sslmode=disable" \
TEST_REDIS_URL="redis://localhost:6379/0" \
go test ./tests/integration/... -v
```

Tanpa environment variable tersebut, integration test akan otomatis di-skip.

---

## Struktur Port dan Koneksi

| Service | Port | URL default |
|---|---|---|
| API Server | 8080 | http://localhost:8080 |
| PostgreSQL | 5432 | `postgres://woles:woles@localhost:5432/woles` |
| Redis | 6379 | `redis://localhost:6379/0` |
| RabbitMQ AMQP | 5672 | `amqp://woles:woles@localhost:5672/` |
| RabbitMQ Management UI | 15672 | http://localhost:15672 (user: `woles`, pass: `woles`) |
| MinIO API | 9000 | http://localhost:9000 |
| MinIO Console | 9001 | http://localhost:9001 (user: `minioadmin`, pass: `minioadmin`) |

---

## Dokumentasi API

| Format | Path |
|---|---|
| OpenAPI 3.1 YAML | [`docs/openapi.yaml`](docs/openapi.yaml) |
| Postman Collection | [`docs/postman/woles.postman_collection.json`](docs/postman/woles.postman_collection.json) |
| Postman Environment | [`docs/postman/woles.postman_environment.json`](docs/postman/woles.postman_environment.json) |

Untuk membuka di Postman: **Import → Upload Files** → pilih kedua file di `docs/postman/`.

Untuk melihat OpenAPI spec secara visual, bisa gunakan [Swagger Editor](https://editor.swagger.io) atau [Redocly](https://redocly.github.io/redoc/) dengan paste isi `docs/openapi.yaml`.

---

## Troubleshooting

**`go: cannot find crypto/pbkdf2`**

Gunakan toolchain yang spesifik jika terjadi masalah dengan Go versi sistem:

```bash
GOROOT=$(go env GOROOT) go build ./...
```

Atau install Go 1.25 dari [golang.org/dl](https://golang.org/dl/).

---

**Port sudah terpakai**

```bash
# Cek proses yang memakai port 8080
lsof -i :8080

# Stop semua container Docker
docker compose down
```

---

**Migration gagal karena tabel sudah ada**

```bash
# Lihat status migration
go run ./cmd/migrate/main.go status

# Rollback jika diperlukan
make migrate-down
```

---

**RabbitMQ belum siap saat server start**

Server akan retry koneksi ke RabbitMQ secara otomatis. Pastikan container `woles-rabbitmq` berstatus `healthy` sebelum menjalankan `make run`.

```bash
docker compose ps rabbitmq
```
