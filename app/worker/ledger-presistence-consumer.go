package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/livingdolls/go-blockchain-simulate/logger"

	"github.com/livingdolls/go-blockchain-simulate/app/dto"
	"github.com/livingdolls/go-blockchain-simulate/app/repository"
	"github.com/livingdolls/go-blockchain-simulate/rabbitmq"
	"github.com/rabbitmq/amqp091-go"
)

type LedgerPersistenceConsumer struct {
	client            *rabbitmq.Client
	ledgerRepo        repository.LedgerRepository
	mu                sync.Mutex
	isRunning         bool
	stopChan          chan struct{}
	workerCount       int
	processingTimeout time.Duration
}

func NewLedgerPersistenceConsumer(rmqClient *rabbitmq.Client, ledgerRepo repository.LedgerRepository, workerCount int) *LedgerPersistenceConsumer {
	return &LedgerPersistenceConsumer{
		client:            rmqClient,
		ledgerRepo:        ledgerRepo,
		stopChan:          make(chan struct{}),
		workerCount:       workerCount,
		processingTimeout: 30 * time.Second,
	}
}

func (l *LedgerPersistenceConsumer) Start() error {
	l.mu.Lock()
	if l.isRunning {
		l.mu.Unlock()
		return nil
	}

	l.isRunning = true
	l.mu.Unlock()

	logger.LogInfo("Starting ledger persistence consumer")

	return l.client.Consume(
		rabbitmq.LedgerPresistenceQueue,
		l.workerCount,
		l.handleMessage,
	)
}

func (l *LedgerPersistenceConsumer) handleMessage(msg amqp091.Delivery) {
	defer func() {
		if err := msg.Ack(false); err != nil {
			logger.LogError("Failed to ack message", err)
		}
	}()

	var batch dto.LedgerBatchEvent

	if err := json.Unmarshal(msg.Body, &batch); err != nil {
		logger.LogError("Failed to unmarshal ledger batch event", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), l.processingTimeout)
	defer cancel()

	if err := l.persistLedgerEntries(ctx, batch); err != nil {
		logger.LogError("Failed to persist ledger entries", err)
		return
	}

	logger.LogInfo(fmt.Sprintf("Successfully persisted ledger entries for block %d", batch.BlockNumber))
}

func (l *LedgerPersistenceConsumer) persistLedgerEntries(ctx context.Context, batch dto.LedgerBatchEvent) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if len(batch.Entries) == 0 {
		logger.LogInfo(fmt.Sprintf("No entries to persist for block %d", batch.BlockNumber))
		return nil
	}

	entries := make([]repository.LedgerEntry, 0, len(batch.Entries))
	for _, entry := range batch.Entries {
		entries = append(entries, repository.LedgerEntry{
			BlockID:      batch.BlockID,
			TxID:         entry.TxID,
			Address:      entry.Address,
			Amount:       entry.Amount,
			BalanceAfter: entry.BalanceAfter,
		})
	}

	logger.LogInfo(fmt.Sprintf("Persisting %d ledger entries for block %d", len(entries), batch.BlockNumber))

	err := l.ledgerRepo.BulkCreate(entries)
	if err != nil {
		logger.LogError("Error persisting entries", err)
		return err
	}

	logger.LogInfo(fmt.Sprintf("Successfully inserted %d ledger entries (blockID=%d)", len(entries), batch.BlockID))
	return nil
}

func (l *LedgerPersistenceConsumer) Stop() {
	l.mu.Lock()

	if !l.isRunning {
		l.mu.Unlock()
		return
	}

	l.isRunning = false
	l.mu.Unlock()

	logger.LogInfo("Stopping ledger persistence")
	close(l.stopChan)
	logger.LogInfo("Stopped ledger persistence consumer")
}
