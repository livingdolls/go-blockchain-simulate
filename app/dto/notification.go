package dto

import (
	"crypto/rand"
	"fmt"
	"time"
)

type NotificationPriority string

const (
	PriorityHigh   NotificationPriority = "high"
	PriorityMedium NotificationPriority = "medium"
	PriorityLow    NotificationPriority = "low"
)

type NotificationType string

const (
	TypeTransactionConfirmed  NotificationType = "TRANSACTION_CONFIRMED"
	TypeTransactionSubmitted  NotificationType = "TRANSACTION_SUBMITTED"
	TypeBlockConfirmed        NotificationType = "BLOCK_CONFIRMED"
	TypeRewardEarned          NotificationType = "REWARD_EARNED"
	TypeBalanceUpdated        NotificationType = "BALANCE_UPDATED"
	TypeTransactionBlockMined NotificationType = "TRANSACTION_BLOCK_MINED"
)

type NotificationChannel string

const (
	ChannelWebSocket NotificationChannel = "ws"
	ChannelEmail     NotificationChannel = "email"
	ChannelSMS       NotificationChannel = "sms"
	ChannelAudit     NotificationChannel = "audit"
)

type NotificationEvent struct {
	ID       string               `json:"id"`
	Type     NotificationType     `json:"type"`
	Priority NotificationPriority `json:"priority"`

	RecipientAddress string `json:"recipient_address"`
	RelatedTxID      *int64 `json:"related_tx_id,omitempty"`
	RelatedBlockID   *int64 `json:"related_block_id,omitempty"`

	Title   string                 `json:"title"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data"`

	Channels  []NotificationChannel `json:"channels"` // ws, email, sms, audit
	Timestamp int64                 `json:"timestamp"`
	CreatedAt int64                 `json:"created_at"`
	ExpiresAt int64                 `json:"expires_at"`
}

type TransactionConfirmedData struct {
	TxID        int64   `json:"tx_id"`
	FromAddress string  `json:"from_address"`
	ToAddress   string  `json:"to_address"`
	Amount      float64 `json:"amount"`
	Fee         float64 `json:"fee"`
	TxHash      string  `json:"tx_hash"`
	Status      string  `json:"status"`
	BlockNumber int64   `json:"block_number"`
	Timestamp   int64   `json:"timestamp"`
	ConfirmTime string  `json:"confirm_time"`
}

type RewardEarnedData struct {
	BlockNumber int64   `json:"block_number"`
	Amount      float64 `json:"amount"`
	Source      string  `json:"source"` // BLOCK REWARD, STAKING REWARD, etc.
	MinerAddr   string  `json:"miner_address"`
	Timestamp   int64   `json:"timestamp"`
}

type BlockConfirmedData struct {
	BlockNumber   int64   `json:"block_number"`
	TxCount       int     `json:"tx_count"`
	TotalFees     float64 `json:"total_fees"`
	BlockHash     string  `json:"block_hash"`
	Confirmations int     `json:"confirmations"`
	Timestamp     int64   `json:"timestamp"`
}

type BalanceUpdatedData struct {
	NewYTEBalance float64 `json:"new_yte_balance"`
	NewUSDBalance float64 `json:"new_usd_balance"`
	ChangeType    string  `json:"change_type"` // INCREASE, DECREASE
	RelatedTxID   *int64  `json:"related_tx_id,omitempty"`
	Timestamp     int64   `json:"timestamp"`
}

// NewNotificationEvent creates a new notification event
func NewNotificationEvent(
	notificationType NotificationType,
	priority NotificationPriority,
	recipientAddr string,
	title string,
	message string,
	channels []NotificationChannel,
) *NotificationEvent {
	now := time.Now()

	return &NotificationEvent{
		ID:               generateNotificationID(),
		Type:             notificationType,
		Priority:         priority,
		RecipientAddress: recipientAddr,
		Title:            title,
		Message:          message,
		Channels:         channels,
		Data:             make(map[string]interface{}),
		Timestamp:        now.Unix(),
		CreatedAt:        now.Unix(),
		ExpiresAt:        now.AddDate(0, 0, 1).Unix(),
	}
}

// generateNotificationID creates a unique notification ID
func generateNotificationID() string {
	return time.Now().Format("2006-01-02T15:04:05.000") + "-" + generateRandomString(8)
}

// generateRandomString creates a cryptographically secure random string
func generateRandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	randomBytes := make([]byte, n)

	if _, err := rand.Read(randomBytes); err != nil {
		// Fallback jika ada error
		for i := range b {
			b[i] = letters[i%len(letters)]
		}
		return string(b)
	}

	for i := range randomBytes {
		b[i] = letters[randomBytes[i]%byte(len(letters))]
	}

	return string(b)
}

// IsValidPriority checks if the given priority is valid
func (p NotificationPriority) IsValid() bool {
	switch p {
	case PriorityHigh, PriorityMedium, PriorityLow:
		return true
	}
	return false
}

// IsValidType checks if the given type is valid
func (t NotificationType) IsValid() bool {
	switch t {
	case TypeTransactionConfirmed, TypeTransactionSubmitted, TypeBlockConfirmed,
		TypeRewardEarned, TypeBalanceUpdated, TypeTransactionBlockMined:
		return true
	}
	return false
}

// IsValidChannel checks if the given channel is valid
func (c NotificationChannel) IsValid() bool {
	switch c {
	case ChannelWebSocket, ChannelEmail, ChannelSMS, ChannelAudit:
		return true
	}
	return false
}

// SetData sets the data payload for the notification
func (ne *NotificationEvent) SetData(data interface{}) error {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid data format: expected map[string]interface{}")
	}
	ne.Data = dataMap
	return nil
}

type NotificationDeliveryStatus struct {
	NotificationID string
	Channel        NotificationChannel
	Status         string
	Attempts       int
	LastAttemptAt  int64
	DeliveredAt    int64
	ErrorMessage   string
}
