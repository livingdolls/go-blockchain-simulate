package websocket

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/livingdolls/go-blockchain-simulate/app/entity"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

type ClientWS struct {
	address string
	conn    *websocket.Conn
	send    chan []byte
	hub     *Hub
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
				log.Printf("WebSocket read error: %v", err)
			}
			break
		}

		// Here you can handle incoming messages from the client if needed
		log.Printf("Received message from client: %s", message)
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
				log.Printf("WebSocket write error: %v", err)
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
		log.Printf("WebSocket handleMessage unmarshal error: %v", err)
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
				log.Printf("WebSocket handleMessage subscribe unmarshal error: %v", err)
			}
		} else {
			log.Printf("WebSocket handleMessage subscribe marshal error: %v", err)
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
			log.Println("WebSocket client send channel full, dropping response")
		}
	}
}
