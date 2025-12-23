package websocket

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type marketClient struct {
	hub  *MarketHub
	conn *websocket.Conn
	send chan []byte
}

type MarketHub struct {
	upgrader   websocket.Upgrader
	clients    map[*marketClient]bool
	register   chan *marketClient
	unregister chan *marketClient
	broadcast  chan []byte
}

func NewMarketHub() *MarketHub {
	return &MarketHub{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		clients:    make(map[*marketClient]bool),
		register:   make(chan *marketClient),
		unregister: make(chan *marketClient),
		broadcast:  make(chan []byte, 32),
	}
}

func (h *MarketHub) Run() {
	for {
		select {
		case c := <-h.register:
			h.clients[c] = true
		case c := <-h.unregister:
			if _, ok := h.clients[c]; ok {
				delete(h.clients, c)
				close(c.send)
				_ = c.conn.Close()
			}
		case msg := <-h.broadcast:
			for c := range h.clients {
				select {
				case c.send <- msg:
				default:
					delete(h.clients, c)
					close(c.send)
					_ = c.conn.Close()
				}
			}
		}
	}
}

func (h *MarketHub) BroadcastMessage(message []byte) {
	payload, err := json.Marshal(message)

	if err != nil {
		log.Printf("WebSocket broadcast marshal error: %v", err)
		return
	}

	select {
	case h.broadcast <- payload:
	default:
		log.Println("WebSocket broadcast channel full, dropping message")
	}
}

func (h *MarketHub) HandleMarketWS(c *gin.Context) {
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)

	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &marketClient{
		hub:  h,
		conn: conn,
		send: make(chan []byte, 8),
	}

	h.register <- client
	go client.writePump()
	client.readPump()
}

func (c *marketClient) readPump() {
	defer func() {
		c.hub.unregister <- c
	}()

	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			break
		}
	}
}

func (c *marketClient) writePump() {
	for msg := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			break
		}
	}

	_ = c.conn.Close()
}
