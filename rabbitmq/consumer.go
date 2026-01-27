package rabbitmq

import (
	"log"

	"github.com/rabbitmq/amqp091-go"
)

func (c *RabbitMQConn) Consume(queue string, workers int, handler func(amqp091.Delivery)) error {
	for i := 0; i < workers; i++ {
		ch, err := c.pool.Get()

		if err != nil {
			return err
		}

		msgs, err := ch.Consume(
			queue,
			"",
			false,
			false,
			false,
			false,
			nil,
		)

		if err != nil {
			return err
		}

		go func(ch *amqp091.Channel) {
			defer c.pool.Put(ch)
			for msg := range msgs {
				handler(msg)
			}

			log.Printf("[RABBITMQ] Stopped consuming from queue: %s", queue)
		}(ch)
	}

	return nil
}
