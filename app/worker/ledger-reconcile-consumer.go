package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/livingdolls/go-blockchain-simulate/logger"

	"github.com/livingdolls/go-blockchain-simulate/app/dto"
	"github.com/livingdolls/go-blockchain-simulate/app/models"
	"github.com/livingdolls/go-blockchain-simulate/app/repository"
	"github.com/livingdolls/go-blockchain-simulate/rabbitmq"
	"github.com/rabbitmq/amqp091-go"
)

type RecoilConfig struct {
	WorkerCount       int
	ReconcileWorkers  int
	ProcessingTimeout time.Duration
	MaxDiscrepancies  int
}

type LedgerReconcileConsumer struct {
	client          *rabbitmq.Client
	walletRepo      repository.UserWalletRepository
	ledgerRepo      repository.LedgerRepository
	discrepancyRepo repository.DiscrepancyRepository
	mu              sync.Mutex
	isRunning       bool
	stopCtx         context.Context
	stopCancel      context.CancelFunc
	cfg             RecoilConfig
	reconcileQueue  chan dto.LedgerBatchEvent
	discrepancies   []dto.BalanceReconciliation
	discrepanciesMu sync.RWMutex
}

func NewLedgerReconcileConsumer(rmqClient *rabbitmq.Client, walletRepo repository.UserWalletRepository, ledgerRepo repository.LedgerRepository, discrepancyRepo repository.DiscrepancyRepository, cfg RecoilConfig) *LedgerReconcileConsumer {
	ctx, cancel := context.WithCancel(context.Background())
	return &LedgerReconcileConsumer{
		client:          rmqClient,
		walletRepo:      walletRepo,
		ledgerRepo:      ledgerRepo,
		discrepancyRepo: discrepancyRepo,
		stopCtx:         ctx,
		stopCancel:      cancel,
		cfg:             cfg,
		reconcileQueue:  make(chan dto.LedgerBatchEvent, 1000),
		discrepancies:   make([]dto.BalanceReconciliation, 0, cfg.MaxDiscrepancies),
	}
}

func (l *LedgerReconcileConsumer) Start() error {
	l.mu.Lock()
	if l.isRunning {
		l.mu.Unlock()
		return nil
	}

	l.isRunning = true
	l.mu.Unlock()

	logger.LogInfo("Starting reconcile consumer")

	// start reconcile worker pool
	for i := 0; i < l.cfg.ReconcileWorkers; i++ {
		go l.reconcileWorker(i)
	}

	return l.client.Consume(
		rabbitmq.LedgerReconcileQueue,
		l.cfg.WorkerCount,
		l.handleMessage,
	)
}

func (l *LedgerReconcileConsumer) handleMessage(msg amqp091.Delivery) {
	defer func() {
		if err := msg.Ack(false); err != nil {
			logger.LogError("Failed to ack message", err)
		}
	}()

	var batch dto.LedgerBatchEvent
	if err := json.Unmarshal(msg.Body, &batch); err != nil {
		logger.LogError("Failed to unmarshal ledger batch", err)
		return
	}

	// reconcile setiap 50 blocks
	if batch.BlockNumber%50 == 0 {
		select {
		case l.reconcileQueue <- batch:
			logger.LogInfo(fmt.Sprintf("Reconciliation enqueued for block #%d", batch.BlockNumber))
		default:
			logger.LogInfo(fmt.Sprintf("Reconciliation queue full for block #%d (will reconcile at next checkpoint)", batch.BlockNumber))
		}
	}

	logger.LogInfo(fmt.Sprintf(
		"Processed block #%d with %d entries",
		batch.BlockNumber,
		len(batch.Entries),
	))
}

// worker untuk reconcile balances
func (l *LedgerReconcileConsumer) reconcileWorker(id int) {
	logger.LogInfo(fmt.Sprintf("Reconcile worker %d started", id))

	for {
		select {
		case <-l.stopCtx.Done():
			logger.LogInfo(fmt.Sprintf("Reconcile worker %d stopping", id))
			return
		case batch := <-l.reconcileQueue:
			ctx, cancel := context.WithTimeout(l.stopCtx, l.cfg.ProcessingTimeout)
			l.reconcileBalances(ctx, batch)
			cancel()
		}
	}
}

// core logic untuk reconcile balances
func (l *LedgerReconcileConsumer) reconcileBalances(ctx context.Context, batch dto.LedgerBatchEvent) {
	// get unique addresses from batch entries
	addressMap := make(map[string]bool)

	for _, entry := range batch.Entries {
		addressMap[entry.Address] = true
	}

	addresses := make([]string, 0, len(addressMap))
	for addr := range addressMap {
		addresses = append(addresses, addr)
	}

	//check context done
	select {
	case <-ctx.Done():
		logger.LogInfo(fmt.Sprintf("Context done before reconciliation for block #%d", batch.BlockNumber))
		return
	default:
	}

	// get wallet balances
	wallets, err := l.walletRepo.GetMultipleByAddress(addresses)
	if err != nil {
		logger.LogError("Failed to get wallets", err)
		return
	}

	//  get ledger entries double check
	ledgerEntries, err := l.ledgerRepo.GetEntriesByBlockID(batch.BlockID)
	if err != nil {
		logger.LogError("Failed to get ledger entries by block ID", err)

		// fallback: gunakan entries dari batch
		logger.LogInfo(fmt.Sprintf("Falling back to batch entries (%d entries)", len(batch.Entries)))

		ledgerEntries = l.batchEntriesToLedgerEntries(batch.Entries)
		l.flagBlockForManualReview(batch.BlockID, batch.BlockNumber, "LEDGER_QUERY_FAILED")

	}

	ledgerMap := make(map[string]float64)
	lastEntryIDMap := make(map[string]int64)

	for _, entry := range ledgerEntries {
		ledgerMap[entry.Address] = entry.BalanceAfter

		if entry.ID > lastEntryIDMap[entry.Address] {
			lastEntryIDMap[entry.Address] = entry.ID
		}
	}

	// check missing wallet entries
	for _, entry := range ledgerEntries {
		walletExists := false
		for _, wallet := range wallets {
			if wallet.UserAddress == entry.Address {
				walletExists = true
				break
			}
		}

		if !walletExists {
			logger.LogInfo(
				fmt.Sprintf("Missing wallet address %s", entry.Address),
			)

			discrepancies := dto.BalanceReconciliation{
				Address:         entry.Address,
				ExpectedBalance: ledgerMap[entry.Address],
				ActualBalance:   0,
				Difference:      -ledgerMap[entry.Address],
				LastEntryID:     lastEntryIDMap[entry.Address],
				BlockNumber:     batch.BlockNumber,
				Timestamp:       time.Now().Unix(),
			}
			l.storeDiscrepancy(discrepancies)
		}
	}

	// check setiap address
	for _, wallet := range wallets {

		select {
		case <-ctx.Done():
			logger.LogInfo(fmt.Sprintf("Context done during reconciliation for block #%d", batch.BlockNumber))
			return
		default:
		}

		//compare wallet balance dengan ledger balance
		expectedBalance := ledgerMap[wallet.UserAddress]

		// compare balances
		if math.Abs(expectedBalance-wallet.YTEBalance) > 1e-9 {
			discrepancy := dto.BalanceReconciliation{
				Address:         wallet.UserAddress,
				ExpectedBalance: expectedBalance,
				ActualBalance:   wallet.YTEBalance,
				Difference:      wallet.YTEBalance - expectedBalance,
				LastEntryID:     0,
				BlockNumber:     batch.BlockNumber,
				Timestamp:       time.Now().Unix(),
			}

			l.storeDiscrepancy(discrepancy)

			logger.LogInfo(
				fmt.Sprintf("DISCREPANCY: %s - Ledger: %.8f, Wallet: %.8f, Diff: %.8f",
					wallet.UserAddress,
					expectedBalance,
					wallet.YTEBalance,
					discrepancy.Difference,
				))
		}
	}

	logger.LogInfo(
		fmt.Sprintf("Reconciliation complete for block #%d - Checked %d addresses",
			batch.BlockNumber,
			len(wallets),
		))
}

func (l *LedgerReconcileConsumer) storeDiscrepancy(discrepancy dto.BalanceReconciliation) {
	l.discrepanciesMu.Lock()
	defer l.discrepanciesMu.Unlock()

	// check if discrepancy already exists
	for _, existing := range l.discrepancies {
		if existing.Address == discrepancy.Address && existing.BlockNumber == discrepancy.BlockNumber && existing.ExpectedBalance == discrepancy.ExpectedBalance && existing.ActualBalance == discrepancy.ActualBalance {
			logger.LogInfo(fmt.Sprintf("Discrepancy already recorded for %s at block #%d, skipping duplicate",
				discrepancy.Address, discrepancy.BlockNumber))
			return
		}
	}

	// bounded slice
	if len(l.discrepancies) >= l.cfg.MaxDiscrepancies {
		// remove oldest
		copy(l.discrepancies, l.discrepancies[1:])
		l.discrepancies[len(l.discrepancies)-1] = discrepancy
		logger.LogInfo("Removed oldest discrepancy to maintain max size")
	} else {
		l.discrepancies = append(l.discrepancies, discrepancy)
	}

	// simpan ke db
	dbDiscrepancy := models.BalanceDiscrepancy{
		Address:         discrepancy.Address,
		BlockNumber:     discrepancy.BlockNumber,
		ExpectedBalance: discrepancy.ExpectedBalance,
		ActualBalance:   discrepancy.ActualBalance,
		Timestamp:       discrepancy.Timestamp,
	}

	if err := l.discrepancyRepo.StoreDiscrepancy(dbDiscrepancy); err != nil {
		logger.LogError("Failed to store discrepancy to db", err)
	}

	logger.LogInfo(fmt.Sprintf("Stored discrepancy for address %s at block #%d", discrepancy.Address, discrepancy.BlockNumber))
}

func (l *LedgerReconcileConsumer) batchEntriesToLedgerEntries(entries []dto.LedgerEntryEvent) []repository.LedgerEntryWithID {
	ledgerEntries := make([]repository.LedgerEntryWithID, 0, len(entries))

	for _, entry := range entries {
		ledgerEntries = append(ledgerEntries, repository.LedgerEntryWithID{
			ID:           entry.EntryID,
			BlockID:      entry.BlockID,
			TxID:         entry.TxID,
			Address:      entry.Address,
			Amount:       entry.Amount,
			BalanceAfter: entry.BalanceAfter,
		})
	}

	return ledgerEntries
}

func (l *LedgerReconcileConsumer) flagBlockForManualReview(blockID int64, blockNumber int, reason string) {
	logger.LogInfo(fmt.Sprintf("Block #%d (ID: %d) flagged for manual review: %s", blockNumber, blockID, reason))

	// TODO :: Store ke db untuk manual review queue/ integrate alerting system
}

func (l *LedgerReconcileConsumer) GetDiscrepancies() []dto.BalanceReconciliation {
	l.discrepanciesMu.RLock()
	defer l.discrepanciesMu.RUnlock()

	// return a copy to
	out := make([]dto.BalanceReconciliation, len(l.discrepancies))
	copy(out, l.discrepancies)
	return out
}

func (l *LedgerReconcileConsumer) Stop() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.isRunning {
		return
	}

	logger.LogInfo("Stopping reconcile consumer")
	l.stopCancel()
	close(l.reconcileQueue)
	l.isRunning = false
}
