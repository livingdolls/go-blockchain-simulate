package worker

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

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

	log.Println("[LEDGER_PERSISTENCE_CONSUMER] Starting ledger persistence consumer...")

	return l.client.Consume(
		rabbitmq.LedgerPresistenceQueue,
		l.workerCount,
		l.handleMessage,
	)
}

func (l *LedgerPersistenceConsumer) handleMessage(msg amqp091.Delivery) {
	defer func() {
		if err := msg.Ack(false); err != nil {
			log.Printf("[LEDGER_PERSISTENCE_CONSUMER] Failed to ack message: %v", err)
		}
	}()

	var batch dto.LedgerBatchEvent

	if err := json.Unmarshal(msg.Body, &batch); err != nil {
		log.Printf("[LEDGER_PERSISTENCE_CONSUMER] Failed to unmarshal ledger batch event: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), l.processingTimeout)
	defer cancel()

	if err := l.persistLedgerEntries(ctx, batch); err != nil {
		log.Printf("[LEDGER_PERSISTENCE_CONSUMER] Failed to persist ledger entries for block %d: %v", batch.BlockNumber, err)
		return
	}

	log.Printf("[LEDGER_PERSISTENCE_CONSUMER] Successfully persisted ledger entries for block %d", batch.BlockNumber)
}

func (l *LedgerPersistenceConsumer) persistLedgerEntries(ctx context.Context, batch dto.LedgerBatchEvent) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if len(batch.Entries) == 0 {
		log.Printf("[LEDGER_PRESISTENCE_CONSUMER] No entries to presist for block %d", batch.BlockNumber)
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

	log.Printf("[LEDGER_PRESISTENCE_CONSUMER] Persisting %d ledger entries for block %d", len(entries), batch.BlockNumber)

	err := l.ledgerRepo.BulkCreate(entries)
	if err != nil {
		log.Printf("[LEDGER_PRESISTENCE_CONSUMER] Error presisting entries : %v", err)
		return err
	}

	log.Printf("[LEDGER_PERSISTENCE_CONSUMER] ✅ Successfully inserted %d ledger entries (blockID=%d)", len(entries), batch.BlockID)
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

	log.Println("[LEDGER_PRESISTENCE_CONSUMER] stoping ledger presistence...")
	close(l.stopChan)
	log.Println("[LEDGER_PERSISTENCE_CONSUMER] Stopped ledger persistence consumer.")
}
