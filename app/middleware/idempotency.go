package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/livingdolls/go-blockchain-simulate/logger"
	"github.com/livingdolls/go-blockchain-simulate/redis"
	"go.uber.org/zap"
)

// IdempotencyConfig mengatur perilaku middleware Idempotency-Key.
type IdempotencyConfig struct {
	// MemoryAdapter digunakan untuk menyimpan cache respons.
	Memory redis.MemoryAdapter
	// TTL adalah jangka waktu cache respons. Default 24 jam.
	TTL time.Duration
	// KeyPrefix untuk membedakan penggunaan (mis. "idempotency:tx").
	KeyPrefix string
	// RequiredScope opsional. Jika di-set, hanya path di dalam scope
	// (mis. "POST /transaction/send") yang akan di-cache. Method lain
	// tidak kena middleware ini.
	RequiredScope []string
}

const defaultIdempotencyTTL = 24 * time.Hour

// responseRecorder membungkus http.ResponseWriter untuk menangkap body
// yang ditulis handler, sehingga dapat di-cache untuk replay pada
// request berikutnya dengan Idempotency-Key yang sama.
type responseRecorder struct {
	gin.ResponseWriter
	body *bytes.Buffer
	code int
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

func (r *responseRecorder) WriteString(s string) (int, error) {
	r.body.WriteString(s)
	return r.ResponseWriter.WriteString(s)
}

func (r *responseRecorder) WriteHeader(code int) {
	r.code = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseRecorder) Status() int {
	if r.code != 0 {
		return r.code
	}
	return r.ResponseWriter.Status()
}

// IdempotencyMiddleware menerapkan pola Stripe-like Idempotency-Key:
//
//   - Client mengirim POST dengan header "Idempotency-Key: <uuid>"
//   - Pertama kali: handler dijalankan, response di-cache
//   - Berikutnya dengan key sama: response di-replay (tidak ada handler call)
//
// Tanpa header, middleware passthrough (tetap call handler).
// Hanya response 2xx yang di-cache; 4xx/5xx tidak di-cache sehingga
// client dapat retry setelah memperbaiki request.
//
// Header: Idempotency-Key
// Header balasan: Idempotent-Replayed: true (saat cache hit)
func IdempotencyMiddleware(cfg IdempotencyConfig) gin.HandlerFunc {
	if cfg.TTL <= 0 {
		cfg.TTL = defaultIdempotencyTTL
	}
	if cfg.KeyPrefix == "" {
		cfg.KeyPrefix = "idempotency:default"
	}
	if cfg.Memory == nil {
		// Tanpa store, return passthrough.
		return func(c *gin.Context) { c.Next() }
	}

	// Pre-build scope map untuk lookup O(1).
	scopeSet := make(map[string]struct{}, len(cfg.RequiredScope))
	for _, s := range cfg.RequiredScope {
		scopeSet[s] = struct{}{}
	}

	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// Filter scope: skip jika method+path tidak ada di scope.
		if len(scopeSet) > 0 {
			key := c.Request.Method + " " + c.Request.URL.Path
			if _, ok := scopeSet[key]; !ok {
				c.Next()
				return
			}
		}

		idemKey := strings.TrimSpace(c.GetHeader("Idempotency-Key"))
		if idemKey == "" {
			c.Next()
			return
		}

		// Batas panjang key untuk mencegah DoS via key sangat panjang.
		if len(idemKey) > 200 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "Idempotency-Key terlalu panjang (maks 200 karakter)",
			})
			return
		}

		// Composite key: prefix + user identity (jika ada) + raw key.
		// User identity mencegah tabrakan antara dua user yang kebetulan
		// menggunakan key yang sama.
		ident := idempotencyUserID(c)
		compositeKey := cfg.KeyPrefix + ":" + ident + ":" + idemKey

		// Cek cache: jika ada, replay response.
		if cached := getCachedResponse(ctx, cfg.Memory, compositeKey); cached != nil {
			logger.LogInfoCtx(ctx, "idempotency cache hit, replay response",
				zap.String("key", idemKey),
				zap.Int("status", cached.StatusCode),
			)
			c.Header("Idempotent-Replayed", "true")
			c.Data(cached.StatusCode, cached.ContentType, cached.Body)
			c.Abort()
			return
		}

		// Passthrough handler dengan response recorder.
		recorder := &responseRecorder{ResponseWriter: c.Writer, body: &bytes.Buffer{}}
		c.Writer = recorder

		c.Next()

		// Hanya cache response 2xx. 4xx/5xx di-skip agar client
		// dapat retry setelah memperbaiki request.
		status := recorder.Status()
		if status >= 200 && status < 300 {
			contentType := recorder.Header().Get("Content-Type")
			if contentType == "" {
				contentType = "application/json; charset=utf-8"
			}
			entry := cachedResponse{
				StatusCode:  status,
				ContentType: contentType,
				Body:        recorder.body.Bytes(),
			}
			storeCachedResponse(ctx, cfg.Memory, compositeKey, entry, cfg.TTL)
		}
	}
}

type cachedResponse struct {
	StatusCode  int    `json:"status"`
	ContentType string `json:"content_type"`
	Body        []byte `json:"body"`
}

func getCachedResponse(ctx context.Context, m redis.MemoryAdapter, key string) *cachedResponse {
	val, ok := m.Get(ctx, key)
	if !ok || len(val) == 0 {
		return nil
	}
	var entry cachedResponse
	if err := json.Unmarshal(val, &entry); err != nil {
		return nil
	}
	return &entry
}

func storeCachedResponse(ctx context.Context, m redis.MemoryAdapter, key string, entry cachedResponse, ttl time.Duration) {
	val, err := json.Marshal(entry)
	if err != nil {
		return
	}
	m.Set(ctx, key, val, ttl)
}

// idempotencyUserID mengekstrak identifier user dari request.
// Order: X-User-ID header > authenticated user dari context > client IP.
// Dipakai untuk namespace idempotency key per user, mencegah tabrakan
// antara user berbeda yang mengirim key sama.
func idempotencyUserID(c *gin.Context) string {
	if v := c.GetHeader("X-User-ID"); v != "" {
		return v
	}
	if v, ok := c.Get("address"); ok {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	return c.ClientIP()
}
