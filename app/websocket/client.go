package websocket

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/livingdolls/go-blockchain-simulate/app/entity"
	"github.com/livingdolls/go-blockchain-simulate/logger"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

// ClientWS mewakili satu koneksi WebSocket aktif. Field done dan closeOnce
// digunakan untuk mencegah race "send on closed channel" antara producer
// (Hub broadcast/send) dan unregister handler.
type ClientWS struct {
	address string
	conn    *websocket.Conn
	send    chan []byte
	hub     *Hub

	// done ditutup saat client sudah tidak valid. Producer harus select
	// terhadap done agar tidak terjadi send-on-closed-channel panic.
	done chan struct{}
	// closeOnce memastikan close() pada send/done dipanggil hanya sekali,
	// mencegah double-close panic.
	closeOnce sync.Once
}

func (c *ClientWS) Read() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(appData string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.LogError("WebSocket read error", err)
			}
			break
		}

		// Here you can handle incoming messages from the client if needed
		logger.LogDebug("Received message from client")
		c.handleMessage(message)
	}
}

func (c *ClientWS) Write() {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Write the message
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				logger.LogError("WebSocket write error", err)
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *ClientWS) handleMessage(data []byte) {
	var msg Message

	if err := json.Unmarshal(data, &msg); err != nil {
		logger.LogError("WebSocket handleMessage unmarshal error", err)
		return
	}

	switch msg.Type {
	case entity.EventTypeSubscribe:
		var subReq SubscribeRequest
		if dataBytes, err := json.Marshal(msg.Data); err == nil {
			if err := json.Unmarshal(dataBytes, &subReq); err == nil {
				for _, eventType := range subReq.Events {
					c.hub.Subscribe(c, eventType)
				}

				// send confirmation
				c.sendResponse(entity.EventTypeSubscribe, SubscribeResponse{
					Success: true,
					Events:  subReq.Events,
				})
			} else {
				logger.LogError("WebSocket subscribe unmarshal error", err)
			}
		} else {
			logger.LogError("WebSocket subscribe marshal error", err)
		}

	case entity.EventTypeUnsubscribe:
		var subReq SubscribeRequest
		if dataBytes, err := json.Marshal(msg.Data); err == nil {
			if err := json.Unmarshal(dataBytes, &subReq); err == nil {
				for _, eventType := range subReq.Events {
					c.hub.Unsubscribe(c, eventType)
				}
			}
		}
	}
}

func (c *ClientWS) sendResponse(msgType entity.MessageType, data any) {
	response := Message{
		Type: msgType,
		Data: data,
	}

	if payload, err := json.Marshal(response); err == nil {
		select {
		case c.send <- payload:
		default:
			logger.LogWarn("WebSocket client send channel full, dropping response")
		}
	}
}
