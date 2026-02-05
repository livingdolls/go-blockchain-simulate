package worker

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/livingdolls/go-blockchain-simulate/app/dto"
	"github.com/livingdolls/go-blockchain-simulate/app/entity"
	"github.com/livingdolls/go-blockchain-simulate/app/publisher"
	"github.com/livingdolls/go-blockchain-simulate/logger"
	"github.com/livingdolls/go-blockchain-simulate/rabbitmq"
	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type NotificationWSConsumer struct {
	client            *rabbitmq.Client
	publisherWS       *publisher.PublisherWS
	mu                sync.Mutex
	isRunning         bool
	stopChan          chan struct{}
	workerCount       int
	processingTimeout time.Duration

	// stats
	statsMu     sync.RWMutex
	stats       map[string]DeliveryStats
	totalSent   int64
	totalFailed int64
}

type DeliveryStats struct {
	UserAddress     string
	TotalSent       int
	TotalDelivered  int
	TotalFailed     int
	LastDeliveredAt int64
}

func NewNotificationWebSocketConsumer(
	client *rabbitmq.Client,
	publisherWS *publisher.PublisherWS,
	workerCount int,
) *NotificationWSConsumer {
	return &NotificationWSConsumer{
		client:            client,
		publisherWS:       publisherWS,
		stopChan:          make(chan struct{}),
		workerCount:       workerCount,
		processingTimeout: 30 * time.Second,
		stats:             make(map[string]DeliveryStats),
	}
}

func (n *NotificationWSConsumer) Start() error {
	n.mu.Lock()

	if n.isRunning {
		n.mu.Unlock()
		return nil
	}

	n.isRunning = true
	n.mu.Unlock()

	logger.LogInfo("Starting notification Websocket consumer", zap.Int("worker_count", n.workerCount))

	return n.client.Consume(
		rabbitmq.NotificationRealTimeQueue,
		n.workerCount,
		n.handleMessage,
	)

}

func (n *NotificationWSConsumer) handleMessage(msg amqp091.Delivery) {
	defer func() {
		if err := msg.Ack(false); err != nil {
			logger.LogError("[NOTIFICATION_WS_CONSUMER] Failed to acknowledge notification WebSocket message: %v\n", err)
		}
	}()

	logger.LogDebug("[NOTIFICATION_WS_CONSUMER] Received notification WebSocket message", zap.ByteString("body", msg.Body))

	var notification dto.NotificationEvent

	if err := json.Unmarshal(msg.Body, &notification); err != nil {
		logger.LogError("[NOTIFICATION_WS_CONSUMER] Failed to unmarshal notification WebSocket message: %v\n", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), n.processingTimeout)
	defer cancel()

	n.deliverNotification(ctx, notification)
}

func (n *NotificationWSConsumer) deliverNotification(ctx context.Context, notification dto.NotificationEvent) {
	hasWSChannel := false

	for _, ch := range notification.Channels {
		if ch == dto.ChannelWebSocket {
			hasWSChannel = true
			break
		}
	}

	if !hasWSChannel {
		logger.LogDebug("[NOTIFICATION_WS_CONSUMER] Notification does not have WebSocket channel, skipping delivery", zap.String("notification_id", notification.ID))
		return
	}

	if n.publisherWS != nil && notification.RecipientAddress != "" {
		// create ws message
		wsMsg := map[string]interface{}{
			"id":               notification.ID,
			"type":             notification.Type,
			"priority":         notification.Priority,
			"title":            notification.Title,
			"message":          notification.Message,
			"data":             notification.Data,
			"timestamp":        time.Now().Unix(),
			"related_tx_id":    notification.RelatedTxID,
			"related_block_id": notification.RelatedBlockID,
		}

		// send via websocket publisher
		n.publisherWS.PublishToAddress(
			notification.RecipientAddress,
			entity.MessageType("notification."+string(notification.Type)),
			wsMsg,
		)

		logger.LogInfo("[NOTIFICATION_WS_CONSUMER] Delivered notification via WebSocket", zap.String("notification_id", notification.ID), zap.String("recipient", notification.RecipientAddress))

		// update stats
		n.updateDeliveryStats(notification.RecipientAddress, true)
	} else {
		logger.LogWarn("WebSocket publisher not available or no recipient")
		n.updateDeliveryStats(notification.RecipientAddress, false)
	}
}

func (n *NotificationWSConsumer) updateDeliveryStats(address string, success bool) {
	n.statsMu.Lock()
	defer n.statsMu.Unlock()

	stats, exists := n.stats[address]

	if !exists {
		stats = DeliveryStats{
			UserAddress: address,
		}
	}

	stats.TotalSent++

	if success {
		stats.TotalDelivered++
		stats.LastDeliveredAt = time.Now().Unix()
		n.totalSent++
	} else {
		stats.TotalFailed++
		n.totalFailed++
	}

	n.stats[address] = stats
}

func (n *NotificationWSConsumer) GetDeliveryStats() map[string]DeliveryStats {
	n.statsMu.RLock()
	defer n.statsMu.RUnlock()

	statsCopy := make(map[string]DeliveryStats)
	for k, v := range n.stats {
		statsCopy[k] = v
	}

	return statsCopy
}

func (n *NotificationWSConsumer) GetTotalStats() map[string]int64 {
	n.statsMu.RLock()
	defer n.statsMu.RUnlock()

	return map[string]int64{
		"total_sent":   n.totalSent,
		"total_failed": n.totalFailed,
	}
}

func (n *NotificationWSConsumer) Stop() {
	n.mu.Lock()

	if !n.isRunning {
		n.mu.Unlock()
		return
	}

	n.isRunning = false
	n.mu.Unlock()

	close(n.stopChan)

	logger.LogInfo("Stopped notification WebSocket consumer")
}

func (n *NotificationWSConsumer) IsRunning() bool {
	n.mu.Lock()
	defer n.mu.Unlock()

	return n.isRunning
}
