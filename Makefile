.PHONY: help run build clean test test-race vet docker-build docker-up docker-down docker-logs docker-restart docker-clean seed dev-up dev-down dev-logs dev-build dev-restart dev-shell

# Default target: tampilkan help.
help:
	@echo "Perintah yang tersedia:"
	@echo "  make run              - Jalankan app langsung (perlu Go terinstall)"
	@echo "  make build            - Build binary ke bin/app"
	@echo "  make test             - Jalankan semua unit test"
	@echo "  make test-race        - Jalankan test dengan race detector"
	@echo "  make vet              - Jalankan go vet"
	@echo "  make seed             - Seed data demo ke database"
	@echo "  make clean            - Hapus artifact build (bin/, tmp/)"
	@echo ""
	@echo "Docker (production image):"
	@echo "  make docker-build     - Build image app saja"
	@echo "  make docker-up        - Start stack lengkap (app + mysql + redis + rabbitmq)"
	@echo "  make docker-down      - Stop semua service"
	@echo "  make docker-logs      - Tail log semua service"
	@echo "  make docker-restart   - Rebuild + restart service app saja"
	@echo "  make docker-clean     - Stop + hapus volume (HATI-HATI: data hilang)"
	@echo ""
	@echo "Docker (dev dengan hot reload - pakai Dockerfile.dev + Air):"
	@echo "  make dev-up           - Start stack dev (source di-mount, hot reload aktif)"
	@echo "  make dev-down         - Stop stack dev"
	@echo "  make dev-logs         - Tail log service app (di container dev)"
	@echo "  make dev-build        - Rebuild image dev saja (mis. setelah ubah Dockerfile.dev)"
	@echo "  make dev-restart      - Kill & restart container app dev saja (volume cache tetap)"
	@echo "  make dev-shell        - Buka shell di dalam container app dev"

run:
	go run main.go

build:
	go build -o bin/app main.go

test:
	go test -timeout 60s ./...

test-race:
	go test -race -timeout 120s ./...

vet:
	go vet ./...

clean:
	rm -rf bin/ tmp/

# Seed data demo ke database (10 users, wallets, balances, 50 transactions).
# Berguna untuk demo/development tanpa harus manual input.
# Pastikan database sudah di-migrate sebelum menjalankan.
seed:
	go run cmd/seed/main.go

# Build image app saja (untuk verify Dockerfile tanpa start service lain).
docker-build:
	docker build -t blockchain-app:dev .

# Start stack lengkap. Tunggu healthcheck lulus sebelum exit.
docker-up:
	docker compose up -d --build
	@echo ""
	@echo "Stack berjalan. Cek:"
	@echo "  App         : http://localhost:3010/healthz"
	@echo "  RabbitMQ UI : http://localhost:15672 (guest/guest)"

# Stop semua service (volume tetap ada).
docker-down:
	docker compose down

# Tail log service app secara default. Override: make docker-logs svc=mysql
docker-logs:
	docker compose logs -f $(or $(svc),app)

# Rebuild image app dan restart kontainernya saja (tanpa sentuh DB/cache).
docker-restart:
	docker compose build app
	docker compose up -d --no-deps app

# Stop + hapus volume. DATA AKAN HILANG.
docker-clean:
	docker compose down -v
	@echo "Volume dihapus. Data MySQL/Redis/RabbitMQ sudah hilang."

# ===== Dev dengan hot reload (Air) =====
#
# Pakai -f docker-compose.yml -f docker-compose.dev.yml agar:
#   - Service mysql/redis/rabbitmq tetap dari base compose
#   - Service app di-override dengan Dockerfile.dev + bind mount
#
# Hot reload:
#   Edit file .go di host -> otomatis ke-rebuild & restart di container
#   dalam ~1-3 detik (incremental build) atau ~30-60 detik (cold start).

dev-up:
	docker compose -f docker-compose.yml -f docker-compose.dev.yml up -d --build
	@echo ""
	@echo "Dev stack berjalan. Hot reload aktif."
	@echo "  App         : http://localhost:3010/healthz"
	@echo "  RabbitMQ UI : http://localhost:15672 (guest/guest)"
	@echo "  Tail log    : make dev-logs"

dev-down:
	docker compose -f docker-compose.yml -f docker-compose.dev.yml down

dev-logs:
	docker compose -f docker-compose.yml -f docker-compose.dev.yml logs -f app

# Rebuild image dev. Hanya perlu kalau Dockerfile.dev atau .air.toml berubah.
# Untuk perubahan source code biasa, TIDAK perlu rebuild - Air handle.
dev-build:
	docker compose -f docker-compose.yml -f docker-compose.dev.yml build app

# Kill & restart container app dev. Cache (go build, go mod) tetap
# tersimpan di named volume, jadi start berikutnya cepat.
dev-restart:
	docker compose -f docker-compose.yml -f docker-compose.dev.yml restart app

# Shell ke dalam container app dev. Useful untuk:
#   - Manual `go build` untuk debug error
#   - Inspect tmp/ folder (hasil build Air)
#   - Cek go.mod / installed tools
dev-shell:
	docker compose -f docker-compose.yml -f docker-compose.dev.yml exec app sh
