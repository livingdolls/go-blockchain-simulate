package rabbitmq

import (
	"context"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

func (c *RabbitMQConn) Publish(ctx context.Context, exchange, routingKey string, body []byte) error {
	ch, err := c.pool.Get()

	if err != nil {
		return err
	}

	defer c.pool.Put(ch)

	return ch.PublishWithContext(
		ctx,
		exchange,
		routingKey,
		false,
		false,
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
			Timestamp:   time.Now(),
		},
	)
}
