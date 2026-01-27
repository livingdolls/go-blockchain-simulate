package rabbitmq

import (
	"context"

	"github.com/livingdolls/go-blockchain-simulate/app/models"
	"github.com/rabbitmq/amqp091-go"
)

type Client struct {
	conn *RabbitMQConn
}

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

func (c *Client) Close() {
	c.conn.Close()
}
