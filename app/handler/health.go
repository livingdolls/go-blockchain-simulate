package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/livingdolls/go-blockchain-simulate/database"
	"github.com/livingdolls/go-blockchain-simulate/logger"
	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// HealthHandler menyediakan endpoint liveness dan readiness probe.
// Dipakai untuk load balancer, Kubernetes liveness/readiness probe,
// atau monitoring eksternal (uptime check).
type HealthHandler struct {
	db          database.Database
	redis       *goredis.Client
	serviceName string
	version     string
}

// NewHealthHandler membuat health handler baru. redisClient boleh nil
// jika redis tidak dipakai; readiness check akan skip komponen yang nil.
func NewHealthHandler(db database.Database, redisClient *goredis.Client, serviceName, version string) *HealthHandler {
	return &HealthHandler{
		db:          db,
		redis:       redisClient,
		serviceName: serviceName,
		version:     version,
	}
}

// ComponentHealth mewakili status satu komponen.
type ComponentHealth struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// HealthResponse adalah payload respons health check.
type HealthResponse struct {
	Status     string                     `json:"status"`
	Service    string                     `json:"service"`
	Version    string                     `json:"version"`
	Timestamp  time.Time                  `json:"timestamp"`
	Components map[string]ComponentHealth `json:"components,omitempty"`
}

// Liveness selalu mengembalikan 200 OK selama proses masih hidup.
// Cocok untuk Kubernetes livenessProbe agar pod tidak di-restart
// karena dependency external yang down.
func (h *HealthHandler) Liveness(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{
		Status:    "ok",
		Service:   h.serviceName,
		Version:   h.version,
		Timestamp: time.Now().UTC(),
	})
}

// Readiness memeriksa dependency utama (database, redis) dan mengembalikan
// 503 jika ada yang tidak reachable. Cocok untuk Kubernetes readinessProbe
// agar pod hanya menerima traffic saat dependency siap.
func (h *HealthHandler) Readiness(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	components := make(map[string]ComponentHealth)
	overallOK := true

	if h.db != nil {
		if err := h.db.Ping(ctx); err != nil {
			components["database"] = ComponentHealth{Status: "down", Error: err.Error()}
			overallOK = false
		} else {
			components["database"] = ComponentHealth{Status: "up"}
		}
	}

	if h.redis != nil {
		if err := h.redis.Ping(ctx).Err(); err != nil {
			components["redis"] = ComponentHealth{Status: "down", Error: err.Error()}
			overallOK = false
		} else {
			components["redis"] = ComponentHealth{Status: "up"}
		}
	}

	status := "ok"
	code := http.StatusOK
	if !overallOK {
		status = "degraded"
		code = http.StatusServiceUnavailable
		logger.LogWarnCtx(c.Request.Context(), "readiness check gagal",
			zap.String("components", componentsString(components)),
		)
	}

	c.JSON(code, HealthResponse{
		Status:     status,
		Service:    h.serviceName,
		Version:    h.version,
		Timestamp:  time.Now().UTC(),
		Components: components,
	})
}

func componentsString(m map[string]ComponentHealth) string {
	out := ""
	for k, v := range m {
		out += k + "=" + v.Status + ";"
	}
	return out
}
