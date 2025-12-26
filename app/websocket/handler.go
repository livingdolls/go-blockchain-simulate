package websocket

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/livingdolls/go-blockchain-simulate/security"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func GinHandler(hub *Hub, jwt security.JWTService) gin.HandlerFunc {
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

		log.Printf("WebSocket connection established for user: %s", claims.Address)

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)

		if err != nil {
			log.Printf("WebSocket upgrade error: %v", err)
			return
		}

		client := &ClientWS{
			address: strings.ToLower(claims.Address),
			hub:     hub,
			conn:    conn,
			send:    make(chan []byte, 256),
		}

		hub.register <- client
		go client.Write()
		go client.Read()
	}
}
