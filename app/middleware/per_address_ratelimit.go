package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/livingdolls/go-blockchain-simulate/logger"
	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// PerAddressRateLimiter adalah middleware rate limit yang menggunakan
// field 'address' atau 'from_address' dari JSON body sebagai identifier.
//
// Catatan: middleware ini membaca body SEKALI untuk ekstrak address,
// lalu me-replace body dengan bytes.Buffer baru agar handler bisa
// membaca ulang. Implementasi aman untuk multiple instance jika
// Redis dipakai bersama.
//
// Jika body bukan JSON valid atau field address tidak ada, fallback
// ke client IP. Tujuannya agar endpoint tidak menjadi DoS amplifier
// ketika ada malformed request.
func PerAddressRateLimiter(redis *goredis.Client, keyPrefix string, limit int, window time.Duration) gin.HandlerFunc {
	// Fallback in-memory token bucket jika Redis down (sama dengan
	// RateLimitMiddleware utama). Rate = limit per window dalam detik.
	ratePerSec := float64(limit) / window.Seconds()
	fallback := NewInMemoryLimiter(ratePerSec, limit)

	return func(c *gin.Context) {
		ctx := c.Request.Context()
		addr := extractAddressFromBody(c)

		ident := addr
		if ident == "" {
			ident = c.ClientIP() // fallback ke IP jika address tidak ditemukan
		}

		key := keyPrefix + ":" + strings.ToLower(ident)

		count, err := redis.Incr(ctx, key).Result()
		if err != nil {
			// Redis down: pakai in-memory fallback. Fail-open dengan degradasi.
			logger.LogWarnCtx(ctx, "per-address rate limit redis error, pakai in-memory fallback",
				zap.String("key", key),
				zap.Error(err),
			)
			if !fallback.Allow(ident) {
				c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
					"success": false,
					"code":    http.StatusTooManyRequests,
					"error":   "Terlalu banyak request untuk address ini, coba lagi nanti",
				})
				return
			}
			c.Next()
			return
		}

		if count == 1 {
			if err := redis.Expire(ctx, key, window).Err(); err != nil {
				logger.LogWarnCtx(ctx, "per-address rate limit expire gagal",
					zap.String("key", key),
					zap.Error(err),
				)
			}
		}

		if count > int64(limit) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"code":    http.StatusTooManyRequests,
				"error":   "Terlalu banyak request untuk address ini, coba lagi nanti",
			})
			return
		}

		c.Next()
	}
}

// extractAddressFromBody membaca JSON body, mengambil field
// 'address' atau 'from_address' (lowercase), lalu me-restore body
// agar handler bisa membaca ulang.
//
// Pendekatan ini dipilih karena Gin's ShouldBindJSON akan membaca
// seluruh body, dan kita tidak ingin handler kena error 'EOF' setelah
// middleware ini consume body.
func extractAddressFromBody(c *gin.Context) string {
	if c.Request.Body == nil {
		return ""
	}

	// Baca body
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return ""
	}
	// Restore body untuk handler
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	if len(bodyBytes) == 0 {
		return ""
	}

	// Parse JSON longgar: hanya butuh field address
	var probe struct {
		Address     string `json:"address"`
		FromAddress string `json:"from_address"`
	}
	if err := json.Unmarshal(bodyBytes, &probe); err != nil {
		return ""
	}

	addr := strings.TrimSpace(probe.Address)
	if addr == "" {
		addr = strings.TrimSpace(probe.FromAddress)
	}
	return addr
}
