package publisher

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/livingdolls/go-blockchain-simulate/app/dto"
	"github.com/livingdolls/go-blockchain-simulate/logger"
	"github.com/livingdolls/go-blockchain-simulate/rabbitmq"
	"go.uber.org/zap"
)

// NotificationPublisher mem-publish NotificationEvent ke RabbitMQ
// untuk diproses oleh NotificationWSConsumer.
type NotificationPublisher interface {
	Publish(ctx context.Context, event dto.NotificationEvent) error
}

type notificationPublisher struct {
	client *rabbitmq.Client
}

func NewNotificationPublisher(client *rabbitmq.Client) NotificationPublisher {
	return &notificationPublisher{client: client}
}

func (p *notificationPublisher) Publish(ctx context.Context, event dto.NotificationEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal notification event: %w", err)
	}

	routingKey := fmt.Sprintf("notification.%s", string(event.Type))

	if err := p.client.Publish(ctx, rabbitmq.NotificationExchange, routingKey, body); err != nil {
		logger.LogError("Gagal publish notification event", err,
			zap.String("type", string(event.Type)),
			zap.String("recipient", event.RecipientAddress),
		)
		return fmt.Errorf("publish notification: %w", err)
	}

	logger.LogDebug("Notification event published",
		zap.String("type", string(event.Type)),
		zap.String("recipient", event.RecipientAddress),
		zap.String("id", event.ID),
	)

	return nil
}
