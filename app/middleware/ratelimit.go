package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/livingdolls/go-blockchain-simulate/logger"
	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RateLimiter menyimpan konfigurasi untuk satu aturan rate limit.
type RateLimiter struct {
	// KeyPrefix dipakai untuk membedakan aturan di Redis, contoh "ratelimit:login".
	KeyPrefix string
	// Limit adalah jumlah maksimum request yang diizinkan dalam Window.
	Limit int
	// Window adalah jangka waktu penghitungan request.
	Window time.Duration
	// IdentifierExtractor mengekstrak identifier (biasanya IP) dari request.
	// Default: client IP.
	IdentifierExtractor func(*gin.Context) string
}

// RateLimitMiddleware membuat middleware rate limit berbasis Redis.
//
// Algoritma: fixed window counter. Setiap request:
//   1. Hitung key = prefix + ":" + identifier
//   2. INCR key (atomic). Jika hasil = 1, set EXPIRE = window.
//   3. Jika hasil > limit, kembalikan 429 Too Many Requests.
//
// Implementasi ini terdistribusi (aman untuk multiple instance) selama
// Redis dipakai bersama. Trade-off: fixed window bisa burst 2x di
// batas window (akhir window 1 + awal window 2); untuk akurasi lebih
// tinggi gunakan sliding window atau token bucket di Redis.
func RateLimitMiddleware(redis *goredis.Client, limiters ...RateLimiter) gin.HandlerFunc {
	if len(limiters) == 0 {
		// No-op: tidak ada rule, return passthrough.
		return func(c *gin.Context) { c.Next() }
	}

	identifiers := make([]func(*gin.Context) string, len(limiters))
	for i, l := range limiters {
		ident := l.IdentifierExtractor
		if ident == nil {
			ident = defaultIdentifier
		}
		identifiers[i] = ident
	}

	return func(c *gin.Context) {
		ctx := c.Request.Context()
		for i, l := range limiters {
			ident := identifiers[i](c)
			if ident == "" {
				continue
			}

			key := l.KeyPrefix + ":" + ident
			count, err := redis.Incr(ctx, key).Result()
			if err != nil {
				// Gagal Redis: fail-open (lewati rate limit, log warning).
				// Alternatif: fail-closed, tapi ini bisa DoS sendiri saat Redis down.
				logger.LogWarnCtx(ctx, "rate limit redis error, fail-open",
					zap.String("key", key),
					zap.Error(err),
				)
				continue
			}

			if count == 1 {
				// Set expire hanya pada INCR pertama dalam window.
				if err := redis.Expire(ctx, key, l.Window).Err(); err != nil {
					logger.LogWarnCtx(ctx, "rate limit expire gagal",
						zap.String("key", key),
						zap.Error(err),
					)
				}
			}

			// Set header informatif untuk client.
			c.Header("X-RateLimit-Limit", strconv.Itoa(l.Limit))
			remaining := int64(l.Limit) - count
			if remaining < 0 {
				remaining = 0
			}
			c.Header("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))

			if count > int64(l.Limit) {
				retryAfter := l.Window.Seconds()
				c.Header("Retry-After", strconv.FormatInt(int64(retryAfter), 10))
				c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
					"error":   "too many requests",
					"limit":   l.Limit,
					"window":  l.Window.String(),
					"retryIn": l.Window.String(),
				})
				return
			}
		}

		c.Next()
	}
}

func defaultIdentifier(c *gin.Context) string {
	return c.ClientIP()
}

// IPKeyExtractor adalah identifier extractor standar yang menggunakan IP client.
// Disimpan sebagai package-private function helper agar tidak duplikat di tiap call site.
var IPKeyExtractor = defaultIdentifier

// InMemoryLimiter adalah rate limiter sederhana berbasis sync.Map untuk
// deployment tanpa Redis. Tidak terdistribusi (per-instance counter),
// cocok untuk development atau single-instance deployment.
//
// Algoritma: token bucket dengan rate (request per second) dan burst.
// Pakai cleanup goroutine untuk evict entry yang tidak dipakai > 10 menit
// agar map tidak grow tanpa batas.
type InMemoryLimiter struct {
	mu      sync.Mutex
	buckets map[string]*bucket
	rate    float64 // tokens per second
	burst   int     // bucket capacity
	ttl     time.Duration
}

type bucket struct {
	tokens   float64
	lastFill time.Time
}

// NewInMemoryLimiter membuat rate limiter dengan rate (req/s) dan burst (max burst).
func NewInMemoryLimiter(rate float64, burst int) *InMemoryLimiter {
	l := &InMemoryLimiter{
		buckets: make(map[string]*bucket),
		rate:    rate,
		burst:   burst,
		ttl:     10 * time.Minute,
	}
	go l.cleanupLoop()
	return l
}

// Allow mengembalikan true jika identifier diizinkan satu request lagi.
func (l *InMemoryLimiter) Allow(ident string) bool {
	if ident == "" {
		return true
	}
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()

	b, ok := l.buckets[ident]
	if !ok {
		b = &bucket{tokens: float64(l.burst), lastFill: now}
		l.buckets[ident] = b
	}

	// Refill: tambahkan token sebanyak rate * elapsed
	elapsed := now.Sub(b.lastFill).Seconds()
	b.tokens += elapsed * l.rate
	if b.tokens > float64(l.burst) {
		b.tokens = float64(l.burst)
	}
	b.lastFill = now

	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

func (l *InMemoryLimiter) cleanupLoop() {
	t := time.NewTicker(l.ttl)
	defer t.Stop()
	for range t.C {
		cutoff := time.Now().Add(-l.ttl)
		l.mu.Lock()
		for k, b := range l.buckets {
			if b.lastFill.Before(cutoff) {
				delete(l.buckets, k)
			}
		}
		l.mu.Unlock()
	}
}
