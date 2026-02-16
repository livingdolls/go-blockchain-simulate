package rabbitmq

import (
	"context"
	"fmt"
	"log"

	"github.com/livingdolls/go-blockchain-simulate/logger"
	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type HandlerFunc func(amqp091.Delivery)

// Deprecated: use ConsumeWithContext instead
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

func (c *RabbitMQConn) ConsumeWithContext(ctx context.Context, queueName string, workerCount int, handler HandlerFunc) error {
	for i := 0; i < workerCount; i++ {
		ch, err := c.pool.Get()

		if err != nil {
			return err
		}

		consumerTag := fmt.Sprintf("consumer-%s-%d", queueName, i)

		msgs, err := ch.Consume(
			queueName,
			consumerTag,
			false,
			false,
			false,
			false,
			nil,
		)

		if err != nil {
			c.pool.Put(ch)
			return err
		}

		go func(
			ch *amqp091.Channel,
			tag string,
			msgs <-chan amqp091.Delivery,
		) {
			defer func() {
				err := ch.Cancel(tag, false)
				if err != nil {
					logger.LogError("Failed to cancel consumer", err, zap.String("consumerTag", tag))
				}

				c.pool.Put(ch)

				logger.LogInfo("Stopped consuming from queue", zap.String("queue", queueName), zap.String("consumerTag", tag))
			}()

			for {
				select {
				case <-ctx.Done():
					logger.LogInfo("Context cancelled, stopping consumer", zap.String("consumerTag", tag))
					return
				case msg, ok := <-msgs:
					if !ok {
						logger.LogInfo("Message channel closed, stopping consumer", zap.String("consumerTag", tag))
						return
					}

					func() {
						defer func() {
							if r := recover(); r != nil {
								logger.LogError("Panic in consumer handler", fmt.Errorf("panic: %v", r), zap.String("consumerTag", tag))
								msg.Nack(false, true)
							}
						}()

						handler(msg)
					}()
				}
			}
		}(ch, consumerTag, msgs)
	}

	return nil
}
