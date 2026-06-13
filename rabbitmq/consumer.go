package rabbitmq

import (
	"context"
	"errors"
	"fmt"

	"github.com/livingdolls/go-blockchain-simulate/logger"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type HandlerFunc func(amqp.Delivery)

// Deprecated: use ConsumeWithContext instead
func (c *RabbitMQConn) Consume(queue string, workers int, handler func(amqp.Delivery)) error {
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

		go func(ch *amqp.Channel) {
			defer c.pool.Put(ch)
			for msg := range msgs {
				handler(msg)
			}
			// Tidak log "stopped" - ini terjadi setiap shutdown (termasuk
			// hot reload Air) dan penuh noise. Caller bisa detect via
			// channel close event kalau butuh.
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
			ch *amqp.Channel,
			tag string,
			msgs <-chan amqp.Delivery,
		) {
			defer func() {
				// Cancel consumer. Error "channel/connection is not open"
				// adalah NORMAL saat graceful shutdown (channel di-close
				// duluan oleh Close()). Jangan log sebagai error - ini
				// misleading operator seolah ada masalah.
				if err := ch.Cancel(tag, false); err != nil {
					var amqpErr *amqp.Error
					if errors.As(err, &amqpErr) && amqpErr.Code == amqp.ChannelError {
						// Channel sudah closed - expected saat shutdown.
						// Log sebagai debug, bukan error.
						logger.LogInfo("Consumer channel already closed",
							zap.String("consumerTag", tag))
					} else {
						logger.LogError("Failed to cancel consumer", err, zap.String("consumerTag", tag))
					}
				}

				c.pool.Put(ch)
			}()

			for {
				select {
				case <-ctx.Done():
					return
				case msg, ok := <-msgs:
					if !ok {
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
