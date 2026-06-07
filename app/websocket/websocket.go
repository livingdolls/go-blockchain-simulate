package websocket

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/livingdolls/go-blockchain-simulate/app/entity"
)

type Hub struct {
	clients       map[*ClientWS]bool
	address       map[string]map[*ClientWS]bool
	subscriptions map[*ClientWS]map[entity.MessageType]bool
	register      chan *ClientWS
	unregister    chan *ClientWS
	broadcast     chan *Message

	stopChan chan struct{}
	mu       sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients:       make(map[*ClientWS]bool),
		address:       make(map[string]map[*ClientWS]bool),
		subscriptions: make(map[*ClientWS]map[entity.MessageType]bool),
		register:      make(chan *ClientWS),
		unregister:    make(chan *ClientWS),
		broadcast:     make(chan *Message, 256),
		stopChan:      make(chan struct{}),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			h.mu.Lock()
			h.clients[c] = true
			h.subscriptions[c] = make(map[entity.MessageType]bool)

			// Register client by address
			if h.address[c.address] == nil {
				h.address[c.address] = make(map[*ClientWS]bool)
			}

			h.address[c.address][c] = true
			h.mu.Unlock()
			log.Printf("Client registered user=%s, total :%d", c.address, len(h.clients))

		case c := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[c]; ok {
				delete(h.clients, c)
				delete(h.subscriptions, c)

				//unregister client by address
				if address, ok := h.address[c.address]; ok {
					delete(address, c)
					if len(address) == 0 {
						delete(h.address, c.address)
					}
				}

				safeCloseClient(c)
			}
			h.mu.Unlock()
			log.Printf("Client unregistered user=%s total=%d", c.address, len(h.clients))
		case msg := <-h.broadcast:
			h.broadcastMessageToSubscribers(msg)
		case <-h.stopChan:
			h.closeAllConnections()
			return
		}
	}
}

// safeCloseClient menutup send channel dan done channel secara idempotent.
// Memicu dua hal sekaligus: producer yang sedang `select` di `done` akan
// keluar, dan consumer (Write goroutine) akan melihat channel closed dan
// keluar setelah mengirim close frame ke client.
func safeCloseClient(c *ClientWS) {
	c.closeOnce.Do(func() {
		close(c.done)
		close(c.send)
		_ = c.conn.Close()
	})
}

func (h *Hub) broadcastMessageToSubscribers(message *Message) {
	payload, err := json.Marshal(message)
	if err != nil {
		log.Printf("[WS] gagal marshal broadcast: %v", err)
		return
	}

	// Pakai Lock (write) bukan RLock agar tidak terjadi race
	// dengan unregister yang menutup client.send. RLock memungkinkan
	// dua broadcast berjalan paralel + unregister di tengahnya,
	// yang bisa menyebabkan "send on closed channel" panic.
	h.mu.Lock()
	defer h.mu.Unlock()

	for client := range h.clients {
		if subscribed, ok := h.subscriptions[client][message.Type]; ok && subscribed {
			// Select terhadap client.done memastikan kita tidak panic
			// "send on closed channel" bila unregister terjadi bersamaan
			// dengan broadcast. default menjaga agar producer tidak blocking
			// saat channel penuh.
			select {
			case client.send <- payload:
			case <-client.done:
				// client sudah di-unregister, skip
			default:
				// channel penuh, drop pesan agar tidak blocking
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
	case <-h.stopChan:
		log.Println("Hub is closing, broadcast message dropped")
	default:
		log.Println("WebSocket broadcast channel full, dropping message")
	}
}

func (h *Hub) SendToAddress(address string, msgType entity.MessageType, data any) {
	message := &Message{
		Type: msgType,
		Data: data,
	}

	payload, err := json.Marshal(message)
	if err != nil {
		log.Printf("[WS] gagal marshal pesan ke address: %v", err)
		return
	}

	// Pakai write lock agar konsisten dengan broadcastMessageToSubscribers
	// dan tidak terjadi race dengan close pada saat unregister.
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, ok := h.address[address]; ok {
		for client := range clients {
			if subscribed, ok := h.subscriptions[client][message.Type]; ok && subscribed {
				select {
				case client.send <- payload:
				case <-client.done:
					// client sudah di-unregister, skip
				case <-h.stopChan:
					return
				default:
					// channel penuh, drop
				}
			}
		}
	}
}

func (h *Hub) closeAllConnections() {
	h.mu.Lock()
	defer h.mu.Unlock()

	log.Printf("Closing all %d client connections...\n", len(h.clients))

	for client := range h.clients {
		safeCloseClient(client)
	}

	// clear maps
	h.clients = make(map[*ClientWS]bool)
	h.address = make(map[string]map[*ClientWS]bool)
	h.subscriptions = make(map[*ClientWS]map[entity.MessageType]bool)

	log.Println("All client connections closed.")
}

func (h *Hub) Close() {
	log.Println("shutting down websocket hub...")
	close(h.stopChan)
	log.Println("websocket hub stopped")
}
