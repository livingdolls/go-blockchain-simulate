package rabbitmq

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/livingdolls/go-blockchain-simulate/logger"

	"github.com/livingdolls/go-blockchain-simulate/app/models"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQConn struct {
	url   string
	conn  *amqp.Connection
	mu    sync.RWMutex
	close chan struct{}

	queues    []models.QueueDef
	exchanges []models.ExchangeDef
	binds     []models.BindDef

	pool *ChannelPool

	// onReconnect dipanggil setelah reconnect sukses + topology restored.
	// Dipakai oleh Client untuk re-register consumers yang mati saat
	// connection drop. Tanpa ini, consumers permanen dead setelah
	// reconnect (msgs channel closed, goroutine exit, nobody restart).
	onReconnect     func()
	onReconnectOnce sync.Once
}

func NewRabbitMQConn(url string) (*RabbitMQConn, error) {
	c := &RabbitMQConn{
		url:   url,
		close: make(chan struct{}),
	}

	if err := c.connect(); err != nil {
		return nil, err
	}

	go c.reconnectLoop()

	return c, nil

}

// RegisterOnReconnect mendaftarkan callback yang dipanggil setiap kali
// reconnect sukses. Consumer menggunakan callback ini untuk re-register
// diri setelah connection drop. Hanya satu callback yang diizinkan
// (overwrite jika dipanggil beberapa kali).
func (c *RabbitMQConn) RegisterOnReconnect(fn func()) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onReconnect = fn
}

func (c *RabbitMQConn) connect() error {
	conn, err := amqp.Dial(c.url)

	if err != nil {
		return fmt.Errorf("[RABBITMQ] failed to connect to RabbitMQ: %w", err)
	}

	// Post-connect health check. Beberapa skenario network half-open
	// bisa membuat Dial() sukses tapi connection langsung broken
	// (firewall timeout, broker baru restart, dst). Detect early.
	if conn.IsClosed() {
		_ = conn.Close()
		return fmt.Errorf("[RABBITMQ] connection closed immediately after dial (network half-open?)")
	}

	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()

	logger.LogInfo("RabbitMQ connected")

	return nil
}

func (c *RabbitMQConn) reconnectLoop() {
	for {
		notify := c.conn.NotifyClose(make(chan *amqp.Error))

		select {
		case err := <-notify:
			if err != nil {
				logger.LogError("RabbitMQ connection closed", err)
				c.reconnect()
			}
		case <-c.close:
			return
		}
	}
}

func (c *RabbitMQConn) reconnect() {
	backoff := time.Second

	for {
		select {
		case <-c.close:
			return
		default:
		}

		log.Println("[RABBITMQ] Attempting to reconnect to RabbitMQ...")
		time.Sleep(backoff)

		if err := c.connect(); err != nil {
			log.Printf("[RABBITMQ] Reconnection failed: %v", err)
			if backoff < 30*time.Second {
				backoff *= 2
			}

			continue
		}

		c.restoreTopology()

		if c.pool != nil {
			c.pool.Rebuild()
		}

		// Panggil callback re-register consumers. Tanpa ini, semua
		// consumer goroutine yang exit saat msgs channel close tidak
		// akan pernah di-restart → pipeline processing mati permanen.
		c.mu.RLock()
		cb := c.onReconnect
		c.mu.RUnlock()
		if cb != nil {
			logger.LogInfo("Re-registering consumers after reconnect")
			cb()
		}

		log.Println("[RABBITMQ] Reconnected to RabbitMQ successfully")
		return
	}
}

func (c *RabbitMQConn) NewChannel() (*amqp.Channel, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.conn == nil || c.conn.IsClosed() {
		return nil, fmt.Errorf("[RABBITMQ] connection is not established")
	}

	return c.conn.Channel()
}

func (c *RabbitMQConn) Close() {
	// Close the channel pool first
	if c.pool != nil {
		c.pool.Close()
	}

	// Stop reconnect loop
	close(c.close)

	// Close the connection
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		c.conn.Close()
	}
}
