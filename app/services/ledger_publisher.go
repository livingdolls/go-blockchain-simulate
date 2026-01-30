package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/livingdolls/go-blockchain-simulate/app/dto"
	"github.com/livingdolls/go-blockchain-simulate/logger"
	"github.com/livingdolls/go-blockchain-simulate/rabbitmq"
)

type LedgerPublisher interface {
	PublishLedgerBatch(ctx context.Context, blockID int64, blockNumber int, entries []dto.LedgerEntryEvent, minerAddress string) error
	PublishLedgerEntry(ctx context.Context, entry dto.LedgerEntryEvent) error
}

type ledgerPublisher struct {
	rmqClient *rabbitmq.Client
}

func NewLedgerPublisher(rmqClient *rabbitmq.Client) LedgerPublisher {
	return &ledgerPublisher{
		rmqClient: rmqClient,
	}
}

// PublishLedgerBatch implements [LedgerPublisher].
// Routing Key: ledger.batch
// Exchange: ledger (topic)
// Queue : ledger.etries
func (l *ledgerPublisher) PublishLedgerBatch(ctx context.Context, blockID int64, blockNumber int, entries []dto.LedgerEntryEvent, minerAddress string) error {
	batch := dto.LedgerBatchEvent{
		BlockID:      blockID,
		BlockNumber:  blockNumber,
		TotalEntries: len(entries),
		Entries:      entries,
		Timestamp:    time.Now().Unix(),
		MinerAddress: minerAddress,
	}

	body, err := json.Marshal(batch)
	if err != nil {
		return fmt.Errorf("[LEDGER_PUBLISHER] failed to marshal ledger batch : %w", err)
	}

	if err := l.rmqClient.Publish(
		ctx,
		rabbitmq.LedgerExchange,
		rabbitmq.LedgerBatchKey,
		body,
	); err != nil {
		return fmt.Errorf("[LEDGER_PUBLISHER] failed to publish ledger batch: %w", err)
	}

	logger.LogInfo(fmt.Sprintf("[LEDGER_PUBLISHER] Published batch for block #%d with %d entries", blockNumber, len(entries)))

	return nil
}

// PublishLedgerEntry implements [LedgerPublisher].
// routing key: ledger.entry
func (l *ledgerPublisher) PublishLedgerEntry(ctx context.Context, entry dto.LedgerEntryEvent) error {
	body, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("[LEDGER_PUBLISHER] failed to marshal ledger entry: %w", err)
	}

	if err := l.rmqClient.Publish(
		ctx,
		rabbitmq.LedgerExchange,
		rabbitmq.LedgerEntryKey,
		body,
	); err != nil {
		return fmt.Errorf("[LEDGER_PUBLISHER] failed to publish ledger entry: %w", err)
	}

	logger.LogInfo(fmt.Sprintf("[LEDGER_PUBLISHER] Published ledger entry for address %s, amount %.4f", entry.Address, entry.Amount))

	return nil
}
