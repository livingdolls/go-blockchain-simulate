package websocket

import (
	"encoding/json"
	"sync"

	"github.com/livingdolls/go-blockchain-simulate/app/entity"
	"github.com/livingdolls/go-blockchain-simulate/logger"
)

type Hub struct {
	clients       map[*ClientWS]bool
	address       map[string]map[*ClientWS]bool
	subscriptions map[*ClientWS]map[entity.MessageType]bool
	register      chan *ClientWS
	unregister    chan *ClientWS
	broadcast     chan *Message

	stopChan chan struct{}
	// closeOnce mencegah close(stopChan) kedua -> panic.
	// Defensive: meski main hanya panggil Close() sekali, handler SIGTERM
	// ganda atau race antara cleanup path berbeda bisa trigger double-close.
	closeOnce sync.Once
	mu        sync.RWMutex

	// ready di-close oleh Run() saat loop sudah mulai consume channel.
	// Producer (BroadCast, register client) bisa wait di ready untuk
	// guarantee bahwa Run() sudah aktif. Tanpa ini, ada race window
	// di mana publish pertama bisa block selamanya (register channel
	// unbuffered).
	ready chan struct{}
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
		ready:         make(chan struct{}),
	}
}

// Ready mengembalikan channel yang di-close saat Hub.Run() loop sudah
// aktif. Caller bisa <-h.Ready() untuk guarantee bahwa register/unregister
// sudah dijamin akan di-consume.
func (h *Hub) Ready() <-chan struct{} {
	return h.ready
}

func (h *Hub) Run() {
	// Tandai ready SEBELUM masuk loop agar producer yang wait di
	// Ready() mendapat signal segera. close(ready) idempotent-safe.
	close(h.ready)

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
			logger.LogInfo("Client registered")

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
			logger.LogInfo("Client unregistered")
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
		logger.LogError("gagal marshal broadcast", err)
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

	logger.LogInfo("Client subscribed")
}

func (h *Hub) Unsubscribe(client *ClientWS, eventType entity.MessageType) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if subs, ok := h.subscriptions[client]; ok {
		delete(subs, eventType)
	}

	logger.LogInfo("Client unsubscribed")
}

func (h *Hub) BroadCast(msgType entity.MessageType, data any) {
	message := &Message{
		Type: msgType,
		Data: data,
	}

	select {
	case h.broadcast <- message:
	case <-h.stopChan:
		logger.LogWarn("Hub is closing, broadcast message dropped")
	default:
		logger.LogWarn("WebSocket broadcast channel full, dropping message")
	}
}

func (h *Hub) SendToAddress(address string, msgType entity.MessageType, data any) {
	message := &Message{
		Type: msgType,
		Data: data,
	}

	payload, err := json.Marshal(message)
	if err != nil {
		logger.LogError("gagal marshal pesan ke address", err)
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

	logger.LogInfo("Closing all client connections")

	for client := range h.clients {
		safeCloseClient(client)
	}

	// clear maps
	h.clients = make(map[*ClientWS]bool)
	h.address = make(map[string]map[*ClientWS]bool)
	h.subscriptions = make(map[*ClientWS]map[entity.MessageType]bool)

	logger.LogInfo("All client connections closed")
}

func (h *Hub) Close() {
	logger.LogInfo("shutting down websocket hub")
	// Idempotent: aman dipanggil beberapa kali (mis. dari shutdown handler
	// DAN dari test cleanup). Sebelumnya double-close akan panic.
	h.closeOnce.Do(func() {
		close(h.stopChan)
	})
	logger.LogInfo("websocket hub stopped")
}
