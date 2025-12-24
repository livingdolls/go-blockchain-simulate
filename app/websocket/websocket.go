package websocket

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/livingdolls/go-blockchain-simulate/app/entity"
)

type Hub struct {
	clients       map[*ClientWS]bool
	subscriptions map[*ClientWS]map[entity.MessageType]bool
	register      chan *ClientWS
	unregister    chan *ClientWS
	broadcast     chan *Message

	mu sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients:       make(map[*ClientWS]bool),
		subscriptions: make(map[*ClientWS]map[entity.MessageType]bool),
		register:      make(chan *ClientWS),
		unregister:    make(chan *ClientWS),
		broadcast:     make(chan *Message, 256),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			h.mu.Lock()
			h.clients[c] = true
			h.subscriptions[c] = make(map[entity.MessageType]bool)
			h.mu.Unlock()
			log.Printf("Client registered, total :%d", len(h.clients))

		case c := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[c]; ok {
				delete(h.clients, c)
				delete(h.subscriptions, c)
				close(c.send)
				_ = c.conn.Close()
			}
			h.mu.Unlock()
			log.Printf("Client unregistered, total :%d", len(h.clients))
		case msg := <-h.broadcast:
			h.broadcastMessageToSubscribers(msg)
		}
	}
}

func (h *Hub) broadcastMessageToSubscribers(message *Message) {
	log.Printf("Broadcasting message to subscribers: %v", message.Type)
	payload, err := json.Marshal(message)
	if err != nil {
		log.Printf("==========================================")
		log.Printf("WebSocket broadcast marshal error: %v", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	log.Printf("COUNT CLIENT %v", len(h.clients))

	for client := range h.clients {
		log.Printf("SEND TO CLIENT: %v", len(h.clients))
		if subscribed, ok := h.subscriptions[client][message.Type]; ok && subscribed {
			select {
			case client.send <- payload:
				log.Println("Message sent to client")
			default:
				log.Println("WebSocket client send channel full, dropping message")
			}
		}
	}
}

func (h *Hub) Subscribe(client *ClientWS, evenType entity.MessageType) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if subs, ok := h.subscriptions[client]; ok {
		subs[evenType] = true
	}

	log.Printf("Client subscribed to %v", evenType)
}

func (h *Hub) Unsubscribe(client *ClientWS, eventType entity.MessageType) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if subs, ok := h.subscriptions[client]; ok {
		delete(subs, eventType)
	}

	log.Printf("Client unsubscribed from %v", eventType)
}

func (h *Hub) BroadCast(msgType entity.MessageType, data any) {
	message := &Message{
		Type: msgType,
		Data: data,
	}

	select {
	case h.broadcast <- message:
	default:
		log.Println("WebSocket broadcast channel full, dropping message")
	}
}
