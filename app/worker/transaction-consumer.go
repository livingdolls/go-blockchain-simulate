package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/livingdolls/go-blockchain-simulate/logger"

	"github.com/livingdolls/go-blockchain-simulate/app/services"
	"github.com/livingdolls/go-blockchain-simulate/rabbitmq"
	"github.com/rabbitmq/amqp091-go"
)

type TransactionMessage struct {
	Type      string  `json:"type"`
	Address   string  `json:"address"`
	ToAddress string  `json:"to_address"`
	Amount    float64 `json:"amount"`
	Nonce     string  `json:"nonce"`
	Signature string  `json:"signature"`
}

type TransactionConsumer struct {
	client            *rabbitmq.Client
	txService         services.TransactionService
	mu                sync.Mutex
	isRunning         bool
	stopChan          chan struct{}
	workerCount       int
	processingTimeout time.Duration
}

func NewTransactionConsumer(
	client *rabbitmq.Client,
	txService services.TransactionService,
	workerCount int,
) *TransactionConsumer {
	return &TransactionConsumer{
		client:            client,
		txService:         txService,
		stopChan:          make(chan struct{}),
		workerCount:       workerCount,
		processingTimeout: 30 * time.Second,
	}
}

// jalankan transaction consumer dengan worker
func (tc *TransactionConsumer) Start(ctx context.Context) error {
	tc.mu.Lock()

	if tc.isRunning {
		tc.mu.Unlock()
		log.Println("[TRANSACTION_CONSUMER] is running")
		return nil
	}

	tc.isRunning = true
	tc.mu.Unlock()

	logger.LogInfo(fmt.Sprintf("Starting with %d workers", tc.workerCount))

	handler := func(delivery amqp091.Delivery) {
		ctx, cancel := context.WithTimeout(context.Background(), tc.processingTimeout)

		defer cancel()

		// parse message
		var msg TransactionMessage

		if err := json.Unmarshal(delivery.Body, &msg); err != nil {
			logger.LogError("Failed to parse message", err)

			// negative acknowledge, without requeue (discard message)
			delivery.Nack(false, false)
			return
		}

		logger.LogInfo(fmt.Sprintf("Receiver transaction: type=%s, From=%s, Amount=%.8f",
			msg.Type, msg.Address, msg.Amount))
		// proses transaksi

		var err error

		switch msg.Type {
		case "SEND":
			_, err = tc.txService.SendWithSignature(
				ctx,
				msg.Address,
				msg.ToAddress,
				msg.Amount,
				msg.Nonce,
				msg.Signature,
			)
		case "BUY":
			_, err = tc.txService.Buy(
				ctx,
				msg.Address,
				msg.Signature,
				msg.Nonce,
				msg.Amount,
			)
		case "SELL":
			_, err = tc.txService.Sell(
				ctx,
				msg.Address,
				msg.Signature,
				msg.Nonce,
				msg.Amount,
			)
		default:
			err = fmt.Errorf("unknown transaction type: %s", msg.Type)
		}

		if err != nil {
			logger.LogError(fmt.Sprintf("Error processing %s transaction", msg.Type), err)
			// negative acknowledge, with requeue
			delivery.Nack(false, true)
			return
		}

		// successfully processed
		logger.LogInfo(fmt.Sprintf("Successfully processed %s transaction from %s", msg.Type, msg.Address))

		// Positive acknowledge
		delivery.Ack(false)
	}

	// start consuming with multiple workers
	if err := tc.client.Consume(rabbitmq.TransactionPendingQueue, tc.workerCount, handler); err != nil {
		tc.isRunning = false
		return fmt.Errorf("[TRANSACTION_CONSUMER] failed to start consuming: %w", err)
	}

	logger.LogInfo("Transaction consumer started")
	return nil
}

func (tc *TransactionConsumer) Stop() {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	if !tc.isRunning {
		logger.LogInfo("Transaction consumer is not running")
		return
	}

	tc.isRunning = false
	close(tc.stopChan)

	logger.LogInfo("Transaction consumer stopped")
}

func (tc *TransactionConsumer) IsRunning() bool {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	return tc.isRunning
}

func (tc *TransactionConsumer) SetProcessingTimeout(timeout time.Duration) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	tc.processingTimeout = timeout
}
