package rabbitmq

import (
	"context"

	"github.com/livingdolls/go-blockchain-simulate/app/models"
	"github.com/rabbitmq/amqp091-go"
)

type Client struct {
	conn *RabbitMQConn
}

// NewClient membuat koneksi RabbitMQ baru dengan URL dan ukuran channel pool.
// URL contoh: amqp://guest:guest@localhost:5672/
func NewClient(url string, poolSize int) (*Client, error) {
	conn, err := NewRabbitMQConn(url)

	if err != nil {
		return nil, err
	}

	_, err = NewChannelPool(conn, poolSize)

	if err != nil {
		return nil, err
	}

	return &Client{conn: conn}, nil
}

func (c *Client) DeclareQueue(q models.QueueDef) error {
	return c.conn.DeclareQueue(q)
}

func (c *Client) DeclareExchange(e models.ExchangeDef) error {
	return c.conn.DeclareExchange(e)
}

func (c *Client) Bind(b models.BindDef) error {
	return c.conn.Bind(b)
}

func (c *Client) Publish(ctx context.Context, exchange, routingKey string, body []byte) error {
	return c.conn.Publish(ctx, exchange, routingKey, body)
}

func (c *Client) Consume(queue string, workers int, handler func(amqp091.Delivery)) error {
	return c.conn.Consume(queue, workers, handler)
}

func (c *Client) ConsumeWithContext(ctx context.Context, queueName string, workerCount int, handler HandlerFunc) error {
	return c.conn.ConsumeWithContext(ctx, queueName, workerCount, handler)
}

// RegisterOnReconnect mendaftarkan callback yang dipanggil setiap kali
// reconnect sukses. Consumer menggunakan callback ini untuk re-register.
func (c *Client) RegisterOnReconnect(fn func()) {
	c.conn.RegisterOnReconnect(fn)
}

func (c *Client) Close() {
	c.conn.Close()
}
