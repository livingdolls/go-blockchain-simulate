package worker

import (
	"context"
	"encoding/json"
	"fmt"
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

type RetryConfig struct {
	MaxRetries        int
	InitialBackoff    time.Duration
	MaxBackoff        time.Duration
	BackoffMultiplier float64
}

type NotificationWSConsumer struct {
	client            *rabbitmq.Client
	publisherWS       *publisher.PublisherWS
	mu                sync.Mutex
	isRunning         bool
	stopChan          chan struct{}
	workerCount       int
	processingTimeout time.Duration
	retryConfig       RetryConfig

	// stats
	statsMu        sync.RWMutex
	stats          map[string]DeliveryStats
	totalSent      int64
	totalDelivered int64
	totalFailed    int64
	totalRetried   int64
}

type DeliveryStats struct {
	UserAddress     string
	TotalSent       int
	TotalDelivered  int
	TotalFailed     int
	TotalRetried    int
	LastDeliveredAt int64
	LastFailedAt    int64
	AvgLatencyMs    float64
}

type MessageMetadata struct {
	NotificationID string
	Attempts       int
	FirstAttemptAt int64
	LastAttemptAt  int64
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
		retryConfig: RetryConfig{
			MaxRetries:        3,
			InitialBackoff:    100 * time.Millisecond,
			MaxBackoff:        10 * time.Second,
			BackoffMultiplier: 2.0,
		},
		stats: make(map[string]DeliveryStats),
	}
}

func (n *NotificationWSConsumer) SetRetryConfig(config RetryConfig) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.retryConfig = config
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
	startTime := time.Now()

	// defer ACK handling
	defer func() {
		if err := msg.Ack(false); err != nil {
			logger.LogError("[NOTIFICATION_WS_CONSUMER] Failed to acknowledge notification WebSocket message: %v\n", err)
		}
	}()

	logger.LogDebug("[NOTIFICATION_WS_CONSUMER] Received notification WebSocket message", zap.ByteString("body", msg.Body))

	// validate
	if len(msg.Body) == 0 {
		logger.LogWarn("[NOTIFICATION_WS_CONSUMER] Empty notification WebSocket message body")
		return
	}

	var notification dto.NotificationEvent

	if err := json.Unmarshal(msg.Body, &notification); err != nil {
		logger.LogError("[NOTIFICATION_WS_CONSUMER] Failed to unmarshal notification WebSocket message: %v\n", err)
		return
	}

	// validate notification
	if err := n.validateNotification(&notification); err != nil {
		logger.LogError("[NOTIFICATION_WS_CONSUMER] Invalid notification WebSocket message: %v\n", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), n.processingTimeout)
	defer cancel()

	metaData := &MessageMetadata{
		NotificationID: notification.ID,
		FirstAttemptAt: startTime.Unix(),
	}

	n.deliverNotificationWithRetry(ctx, notification, metaData)

	// record latency
	latency := time.Since(startTime).Milliseconds()
	n.recordDeliveryMetrics(notification.RecipientAddress, latency)
}

func (n *NotificationWSConsumer) validateNotification(notification *dto.NotificationEvent) error {
	if notification.ID == "" {
		return fmt.Errorf("notification ID is empty")
	}

	if notification.RecipientAddress == "" {
		return fmt.Errorf("recipient address is empty")
	}

	if !notification.Type.IsValid() {
		return fmt.Errorf("invalid notification type: %s", notification.Type)
	}

	if !notification.Priority.IsValid() {
		return fmt.Errorf("invalid notification priority: %s", notification.Priority)
	}

	if len(notification.Channels) == 0 {
		return fmt.Errorf("no notification channels specified")
	}

	for _, ch := range notification.Channels {
		if !ch.IsValid() {
			return fmt.Errorf("invalid notification channel: %s", ch)
		}
	}

	if notification.Title == "" && notification.Message == "" {
		return fmt.Errorf("notification title and message are both empty")
	}

	return nil
}

func (n *NotificationWSConsumer) deliverNotificationWithRetry(ctx context.Context, notification dto.NotificationEvent, metaData *MessageMetadata) {
	// check if ws channel is requested
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

	if n.publisherWS == nil {
		logger.LogWarn("WebSocket publisher not available")
		n.updateDeliveryStats(notification.RecipientAddress, false, 0)
		return
	}

	backoff := n.retryConfig.InitialBackoff

	for attemp := 0; attemp <= n.retryConfig.MaxRetries; attemp++ {
		select {
		case <-ctx.Done():
			logger.LogWarn("[NOTIFICATION_WS_CONSUMER] Context cancelled while delivering notification via WebSocket ", zap.String("notification_id", notification.ID), zap.Int("attemp", attemp+1))
			n.updateDeliveryStats(notification.RecipientAddress, false, 0)
			return
		default:
		}

		metaData.Attempts = attemp + 1
		metaData.LastAttemptAt = time.Now().Unix()

		if err := n.deliverNotification(ctx, notification, attemp); err != nil {
			logger.LogWarn("[NOTIFICATION_WS_CONSUMER] Failed to deliver notification via WebSocket", zap.String("notification_id", notification.ID), zap.Int("attemp", attemp+1), zap.Error(err))

			if attemp == n.retryConfig.MaxRetries {
				logger.LogWarn("[NOTIFICATION_WS_CONSUMER] Max retries reached, giving up on notification delivery", zap.String("notification_id", notification.ID))
				n.updateDeliveryStats(notification.RecipientAddress, false, 0)
				return
			}

			// wait before retrying
			select {
			case <-ctx.Done():
				return
			case <-time.After(backoff):
				// continue to next attempt
			}

			// calculate next backoff
			backoff = time.Duration(float64(backoff) * n.retryConfig.BackoffMultiplier)
			if backoff > n.retryConfig.MaxBackoff {
				backoff = n.retryConfig.MaxBackoff
			}

			n.incrementRetryStats(notification.RecipientAddress)
			continue
		}

		// success
		logger.LogInfo("[NOTIFICATION_WS_CONSUMER] Delivered notification via WebSocket", zap.String("notification_id", notification.ID), zap.String("recipient", notification.RecipientAddress))

		n.updateDeliveryStats(notification.RecipientAddress, true, 0)
		return
	}
}

func (n *NotificationWSConsumer) deliverNotification(ctx context.Context, notification dto.NotificationEvent, attempt int) error {
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
		"attempt":          attempt + 1,
	}

	// send via WebSocket publisher
	deliveryChan := make(chan error, 1)

	go func() {
		n.publisherWS.PublishToAddress(
			notification.RecipientAddress,
			entity.MessageType("notification."+string(notification.Type)),
			wsMsg,
		)
		deliveryChan <- nil
	}()

	// wait for delivery or context done
	select {
	case err := <-deliveryChan:
		return err
	case <-ctx.Done():
		return fmt.Errorf("context cancelled while delivering notification via WebSocket")
	}
}

func (n *NotificationWSConsumer) recordDeliveryMetrics(address string, latencyMs int64) {
	n.statsMu.Lock()
	defer n.statsMu.Unlock()

	stats, exists := n.stats[address]

	if !exists {
		stats = DeliveryStats{
			UserAddress: address,
		}
	}

	// update average latency
	if stats.TotalDelivered > 0 {
		stats.AvgLatencyMs = (stats.AvgLatencyMs*float64(stats.TotalDelivered) + float64(latencyMs)) / float64(stats.TotalDelivered+1)
	} else {
		stats.AvgLatencyMs = float64(latencyMs)
	}

	n.stats[address] = stats
}

func (n *NotificationWSConsumer) updateDeliveryStats(address string, success bool, latencyMs int64) {
	n.statsMu.Lock()
	defer n.statsMu.Unlock()

	stats, exists := n.stats[address]
	if !exists {
		stats = DeliveryStats{
			UserAddress: address,
		}
	}

	// 1. Update Global Counter
	n.totalSent++

	// 2. Update Per-User Stats
	stats.TotalSent++

	if success {
		// Update average latency (Moving Average)
		if stats.TotalDelivered > 0 {
			stats.AvgLatencyMs = (stats.AvgLatencyMs*float64(stats.TotalDelivered) + float64(latencyMs)) / float64(stats.TotalDelivered+1)
		} else {
			stats.AvgLatencyMs = float64(latencyMs)
		}

		stats.TotalDelivered++
		stats.LastDeliveredAt = time.Now().Unix()
		n.totalDelivered++
	} else {
		stats.TotalFailed++
		stats.LastFailedAt = time.Now().Unix()
		n.totalFailed++
	}

	n.stats[address] = stats
}
func (n *NotificationWSConsumer) incrementRetryStats(address string) {
	n.statsMu.Lock()
	defer n.statsMu.Unlock()

	stats, exists := n.stats[address]

	if !exists {
		stats = DeliveryStats{
			UserAddress: address,
		}
	}

	stats.TotalRetried++
	n.totalRetried++

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
		"total_sent":      n.totalSent,
		"total_delivered": n.totalDelivered,
		"total_failed":    n.totalFailed,
		"total_retried":   n.totalRetried,
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
