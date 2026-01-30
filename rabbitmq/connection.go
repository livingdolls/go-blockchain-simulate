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

func (c *RabbitMQConn) connect() error {
	conn, err := amqp.Dial(c.url)

	if err != nil {
		return fmt.Errorf("[RABBITMQ] failed to connect to RabbitMQ: %w", err)
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
	close(c.close)
	c.mu.Lock()
	if c.conn != nil {
		c.conn.Close()
	}
	defer c.mu.Unlock()
}
