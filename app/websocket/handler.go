package websocket

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/livingdolls/go-blockchain-simulate/logger"
	"github.com/livingdolls/go-blockchain-simulate/security"
)

// newUpgrader membuat WebSocket upgrader dengan origin validation.
// allowedOrigins adalah whitelist dari config.ServerConfig.AllowedOrigins.
// Jika kosong, semua origin ditolak (fail-closed).
func newUpgrader(allowedOrigins []string) websocket.Upgrader {
	// Build lookup map untuk O(1) origin check.
	origins := make(map[string]bool, len(allowedOrigins))
	for _, o := range allowedOrigins {
		origins[strings.ToLower(strings.TrimSpace(o))] = true
	}

	return websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			if origin == "" {
				// Tanpa Origin header: tolak (bukan browser request,
				// atau curl/wget — tidak ada risiko CSWH).
				return false
			}
			// Normalize: hilangkan trailing slash, lowercase host.
			origin = strings.ToLower(strings.TrimRight(origin, "/"))
			return origins[origin]
		},
	}
}

func GinHandler(hub *Hub, jwt security.JWTService, allowedOrigins []string) gin.HandlerFunc {
	upgrader := newUpgrader(allowedOrigins)

	return func(c *gin.Context) {
		// get cookie
		token, err := c.Cookie("auth_token")

		if err != nil || strings.TrimSpace(token) == "" {
			http.Error(c.Writer, "Unauthorized: missing or invalid token", http.StatusUnauthorized)
			return
		}

		claims, err := jwt.ValidateToken(token)

		if err != nil {
			http.Error(c.Writer, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
			return
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)

		if err != nil {
			logger.LogError("WebSocket upgrade error", err)
			return
		}

		client := &ClientWS{
			address: strings.ToLower(claims.Address),
			hub:     hub,
			conn:    conn,
			send:    make(chan []byte, 256),
			done:    make(chan struct{}),
		}

		hub.register <- client
		go client.Write()
		go client.Read()
	}
}
