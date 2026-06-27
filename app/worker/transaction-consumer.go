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

// DLQEnvelope membungkus pesan asli yang gagal diproses beserta
// metadata error untuk inspeksi manual di transaction.dlq.
type DLQEnvelope struct {
	OriginalBody json.RawMessage `json:"original_body"`
	FailureStage string          `json:"failure_stage"`
	Error        string          `json:"error"`
	FailedAt     time.Time       `json:"failed_at"`
	RoutingKey   string          `json:"routing_key"`
	Exchange     string          `json:"exchange"`
}

type TransactionConsumer struct {
	client            *rabbitmq.Client
	txService         services.TransactionService
	mu                sync.Mutex
	isRunning         bool
	stopChan          chan struct{}
	workerCount       int
	processingTimeout time.Duration
	parentCtx         context.Context
	parentCancel      context.CancelFunc
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

// sendToDLQ mem-publish envelope pesan gagal ke DLQ exchange.
// Error dari publish di-log tapi tidak menggangu alur: pesan asli tetap
// di-Ack agar tidak terjadi requeue loop pada broker.
func (tc *TransactionConsumer) sendToDLQ(ctx context.Context, delivery amqp091.Delivery, stage string, processErr error) {
	envelope := DLQEnvelope{
		OriginalBody: json.RawMessage(delivery.Body),
		FailureStage: stage,
		Error:        processErr.Error(),
		FailedAt:     time.Now().UTC(),
		RoutingKey:   delivery.RoutingKey,
		Exchange:     delivery.Exchange,
	}

	body, err := json.Marshal(envelope)
	if err != nil {
		logger.LogError("gagal membuat envelope DLQ", err)
		return
	}

	if err := tc.client.Publish(ctx, rabbitmq.DLQExchange, rabbitmq.DLXRoutingKey, body); err != nil {
		logger.LogError("gagal publish ke DLQ", err)
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
	// Derive cancellable context dari input agar handler bisa di-abort
	// saat Stop() dipanggil.
	tc.parentCtx, tc.parentCancel = context.WithCancel(ctx)
	tc.mu.Unlock()

	logger.LogInfo(fmt.Sprintf("Starting with %d workers", tc.workerCount))

	handler := func(delivery amqp091.Delivery) {
		// Pakai parent ctx agar handler dapat dibatalkan saat shutdown.
		// Batas maksimum processing time = processingTimeout, tapi juga
		// berhenti lebih awal bila ctx parent di-cancel.
		ctx, cancel := context.WithTimeout(tc.parentCtx, tc.processingTimeout)

		defer cancel()

		// parse message
		var msg TransactionMessage

		if err := json.Unmarshal(delivery.Body, &msg); err != nil {
			logger.LogError("Failed to parse message", err)
			tc.sendToDLQ(ctx, delivery, "parse", err)
			delivery.Ack(false)
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
				msg.Nonce,
				msg.Signature,
				msg.Amount,
			)
		default:
			err = fmt.Errorf("unknown transaction type: %s", msg.Type)
		}

		if err != nil {
			logger.LogError(fmt.Sprintf("Error processing %s transaction", msg.Type), err)
			tc.sendToDLQ(ctx, delivery, "process", err)
			delivery.Ack(false)
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

// Stop menunggu in-flight handler selesai lalu menutup stopChan.
// Cancellation dilakukan via parent ctx (handler pakai
// context.WithTimeout(parentCtx, ...)), sehingga handler yang
// select terhadap ctx.Done() akan keluar early.
func (tc *TransactionConsumer) Stop() {
	tc.mu.Lock()

	if !tc.isRunning {
		tc.mu.Unlock()
		logger.LogInfo("Transaction consumer is not running")
		return
	}

	tc.isRunning = false
	stopChan := tc.stopChan
	parentCancel := tc.parentCancel
	tc.mu.Unlock()

	// Cancel parent ctx agar handler yang sedang berjalan keluar early.
	if parentCancel != nil {
		parentCancel()
	}

	// Tutup stopChan agar goroutine consumer keluar dari loop.
	select {
	case <-stopChan:
		// sudah ditutup
	default:
		close(stopChan)
	}

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
