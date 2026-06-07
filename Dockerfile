# syntax=docker/dockerfile:1.7

# Tahap build: compile binary statis dengan image Go resmi.
FROM golang:1.24-alpine AS build

# Install git agar go module bisa menarik dependency dari VCS.
RUN apk add --no-cache git ca-certificates

WORKDIR /src

# Cache layer dependency: copy go.mod/go.sum dulu, download, baru copy source.
# Ini memanfaatkan Docker layer cache agar build berikutnya cepat
# ketika hanya source code yang berubah.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build binary statis (CGO=0) untuk image runtime yang minimal.
# Strip debug info dan symbol table agar binary kecil.
ARG VERSION=dev
ARG COMMIT=unknown
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
      -trimpath \
      -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT}" \
      -o /out/app \
      .

# Tahap runtime: image Alpine minimal dengan CA certificates dan user non-root.
FROM alpine:3.20 AS runtime

# CA certificates penting untuk HTTPS outbound (viper, dll).
# tzdata untuk TIMEZONE-aware timestamp di logger.
# wget untuk healthcheck.
RUN apk add --no-cache ca-certificates tzdata wget && \
    addgroup -S app && \
    adduser -S -G app -h /app -s /sbin/nologin app

WORKDIR /app

# Salin binary dari tahap build.
COPY --from=build /out/app /app/app

# Salin default config. config.local.yaml TIDAK di-copy;注入 via env atau mount.
COPY config/config.yaml /app/config/config.yaml

# File dan direktori harus dimiliki user 'app' karena kita jalankan
# sebagai non-root. Config harus readable.
RUN chown -R app:app /app

USER app

WORKDIR /app

# Expose port default aplikasi (lihat config.yaml: server.port=3010).
EXPOSE 3010

ENV APP_ENV=production \
    CONFIG_PATH=/app/config/config.yaml

# Healthcheck: panggil /healthz. App harus handle SIGTERM dengan baik
# (lihat app/shutdown.go + main.go) agar docker stop --time tidak timeout.
HEALTHCHECK --interval=30s --timeout=5s --start-period=20s --retries=3 \
    CMD wget -qO- http://127.0.0.1:3010/healthz >/dev/null 2>&1 || exit 1

ENTRYPOINT ["/app/app"]
