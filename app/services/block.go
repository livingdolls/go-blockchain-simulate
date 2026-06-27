package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/livingdolls/go-blockchain-simulate/app/dto"
	"github.com/livingdolls/go-blockchain-simulate/app/entity"
	"github.com/livingdolls/go-blockchain-simulate/app/models"
	"github.com/livingdolls/go-blockchain-simulate/app/publisher"
	"github.com/livingdolls/go-blockchain-simulate/app/repository"
	"github.com/livingdolls/go-blockchain-simulate/logger"
	"github.com/livingdolls/go-blockchain-simulate/utils"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type BlockService interface {
	GenerateBlock() (models.Block, error)
	GetBlocks(limit, offset int) ([]models.Block, error)
	GetBlockByID(id int64) (models.Block, error)
	GetBlockByBlockNumber(id int64) (models.Block, error)
	GetDetailsByBlockNumber(id int64) (models.Block, error)
	CheckBlockchainIntegrity() error
	GetTransactionByBlockNumber(ctx context.Context, blockNumber int64) ([]models.Transaction, error)
	SearchBlocksByHash(ctx context.Context, hash string) ([]models.Block, error)
	GetBlocksInRange(ctx context.Context, from, to int64) ([]models.Block, error)
	GetBlockStats(ctx context.Context) (dto.BlockStatsResponse, error)
	SearchBlocksByMinerAddress(ctx context.Context, address string, limit, offset int) ([]models.Block, error)
}

// shortAddr memotong address untuk display di notifikasi.
func shortAddr(addr string) string {
	if len(addr) <= 10 {
		return addr
	}
	return addr[:6] + "..." + addr[len(addr)-4:]
}

type blockService struct {
	blockRepo              repository.BlockRepository
	walletRepo             repository.UserWalletRepository
	balanceRepo            repository.UserBalanceRepository
	txRepo                 repository.TransactionRepository
	userRepo               repository.UserRepository
	candle                 CandleService
	market                 MarketEngineService
	publisherWS            *publisher.PublisherWS
	pricingPublisher       MarketPricingPublisher
	ledgerPublisher        LedgerPublisher
	rewardPublisher        RewardPublisher
	notificationPublisher  publisher.NotificationPublisher
}

func NewBlockService(blockRepo repository.BlockRepository, walletRepo repository.UserWalletRepository, balanceRepo repository.UserBalanceRepository, txRepo repository.TransactionRepository, userRepo repository.UserRepository, candle CandleService, market MarketEngineService, publisherWS *publisher.PublisherWS, pricingPublisher MarketPricingPublisher, ledgerPublisher LedgerPublisher, rewardPublisher RewardPublisher, notificationPublisher publisher.NotificationPublisher) BlockService {
	return &blockService{
		blockRepo:              blockRepo,
		walletRepo:             walletRepo,
		balanceRepo:            balanceRepo,
		txRepo:                 txRepo,
		userRepo:               userRepo,
		candle:                 candle,
		market:                 market,
		publisherWS:            publisherWS,
		pricingPublisher:       pricingPublisher,
		ledgerPublisher:        ledgerPublisher,
		rewardPublisher:        rewardPublisher,
		notificationPublisher:  notificationPublisher,
	}
}

func (s *blockService) GenerateBlock() (models.Block, error) {
	// ========================================
	// PHASE 1: Read-only validation (NO LOCKS)
	// ========================================

	// Get last block (read-only)
	lastBlock, err := s.blockRepo.GetLastBlock()
	if err != nil {
		return models.Block{}, fmt.Errorf("get last block: %w", err)
	}

	// get current market state
	var currentMarket models.MarketEngine
	if s.market != nil {
		currentMarket, err = s.market.GetState()
		if err != nil {
			// If error, use default price
			if !strings.Contains(err.Error(), "no rows in result set") {
				return models.Block{}, fmt.Errorf("get market state: %w", err)
			}
			currentMarket.Price = 100.0 // default only on error
		} else if currentMarket.Price == 0 {
			// Fallback if price is 0
			currentMarket.Price = 100.0
		}
	} else {
		// default market state if service not available
		currentMarket.Price = 100.0
	}

	// Get pending transactions (read-only, limit to 100 to prevent timeout)
	pendingTxs, err := s.txRepo.GetPendingTransactions(100)
	if err != nil {
		return models.Block{}, fmt.Errorf("get pending transactions: %w", err)
	}

	if len(pendingTxs) == 0 {
		return models.Block{}, entity.ErrNoPendingTransactions
	}

	var buyVolume, sellVolume float64

	for _, t := range pendingTxs {
		if strings.EqualFold(t.Type, "BUY") {
			buyVolume += t.Amount
		} else if strings.EqualFold(t.Type, "SELL") {
			sellVolume += t.Amount
		}
	}

	// Collect unique addresses
	uniqueAddresses := make(map[string]bool)
	for _, t := range pendingTxs {
		uniqueAddresses[t.FromAddress] = true
		uniqueAddresses[t.ToAddress] = true
	}

	addresses := make([]string, 0, len(uniqueAddresses)+1) // +1 MINER Address
	for addr := range uniqueAddresses {
		addresses = append(addresses, addr)
	}

	// Added miner acount
	addresses = append(addresses, entity.MinerAccountAddress)
	minerAddr := strings.ToLower(entity.MinerAccountAddress)

	// Get all users at once (read-only)
	users, err := s.userRepo.GetMultipleByAddress(addresses)
	if err != nil {
		return models.Block{}, fmt.Errorf("get multiple users: %w", err)
	}

	// Build user cache
	userCache := make(map[string]models.User)
	for _, u := range users {
		userCache[strings.ToLower(u.Address)] = u
	}

	// get all wallets
	wallets, err := s.walletRepo.GetMultipleByAddress(addresses)
	if err != nil {
		return models.Block{}, fmt.Errorf("get multiple wallets: %w", err)
	}

	// build wallet cache
	walletCache := make(map[string]models.UserWallet)
	for _, w := range wallets {
		walletCache[strings.ToLower(w.UserAddress)] = w
	}

	// Pre-validate in-memory (no DB)
	balances := make(map[string]float64)
	for _, addr := range addresses {
		lookupAddr := strings.ToLower(addr)
		wallet, exists := walletCache[lookupAddr]

		if exists {
			balances[lookupAddr] = wallet.YTEBalance
		} else {
			balances[lookupAddr] = 0
		}
	}

	for _, t := range pendingTxs {
		sender, exists := userCache[strings.ToLower(t.FromAddress)]
		if !exists {
			return models.Block{}, fmt.Errorf("sender not found: %s", t.FromAddress)
		}

		// calculato total deduction amount + fee
		totalDeduction := t.Amount + t.Fee

		if balances[strings.ToLower(sender.Address)] < totalDeduction {
			return models.Block{}, fmt.Errorf("insufficient balance for address %s: need %.8f (amount: %.8f + fee: %.8f), have %.8f",
				sender.Address, totalDeduction, t.Amount, t.Fee, balances[strings.ToLower(sender.Address)])
		}

		from := strings.ToLower(t.FromAddress)
		to := strings.ToLower(t.ToAddress)
		balances[from] -= totalDeduction // - amount + fee
		balances[to] += t.Amount
	}

	// MINING PHASE : Prof of Work

	// Get all blocks to calculate next difficulty
	allBlocks, err := s.blockRepo.GetAllBlocks()
	if err != nil {
		return models.Block{}, fmt.Errorf("get all blocks: %w", err)
	}

	// calculate difficulty for next block
	difficulty := utils.CalculateNextDifficulty(allBlocks)

	// calculate merkle root
	merkleRoot := utils.CalculateMerkleRoot(pendingTxs)
	logger.LogDebug("Calculated Merkle Root", zap.String("merkle_root", merkleRoot))

	// Calculate block reward
	nextBlockNumber := lastBlock.BlockNumber + 1
	blockReward := utils.CalculateBlockReward(int64(nextBlockNumber))

	// Tentukan timestamp SEBELUM mining agar hash bisa di-recover
	// persis sama dengan RecalculateBlockHash saat integritas dicek.
	blockTimestamp := time.Now().Unix()

	// Perform mining (this can take 5-60 seconds depending on difficulty)
	logger.LogInfo("Starting mining process",
		zap.Int64("block_number", int64(nextBlockNumber)),
		zap.Int64("difficulty", int64(difficulty)),
		zap.String("merkle_root", merkleRoot),
		zap.Float64("block_reward", blockReward),
	)

	miningResult := utils.MineBlock(lastBlock.BlockNumber+1, lastBlock.CurrentHash, pendingTxs, difficulty, blockTimestamp)

	// check if mining was successful
	if miningResult.Hash == "" {
		return models.Block{}, fmt.Errorf("mining failed to find a valid nonce")
	}

	logger.LogInfo("Mining complete", zap.String("hash", miningResult.Hash))

	// ========================================
	// PHASE 2: Write operations (SHORT TRANSACTION)
	// ========================================

	tx, err := s.blockRepo.BeginTx()
	if err != nil {
		return models.Block{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Lock last block and verify no new block was created
	lastBlockLocked, err := s.blockRepo.GetLastBlockForUpdateWithTx(tx)
	if err != nil {
		return models.Block{}, fmt.Errorf("lock last block: %w", err)
	}

	if lastBlockLocked.BlockNumber != lastBlock.BlockNumber {
		return models.Block{}, fmt.Errorf("new block created while processing, please retry")
	}

	// get locked wallets for update
	lockedWallets, err := s.walletRepo.GetMultipleByAddressWithTx(tx, addresses)
	if err != nil {
		return models.Block{}, fmt.Errorf("lock multiple wallets: %w", err)
	}

	// build current wallet cache
	currentWallets := make(map[string]models.UserWallet)
	for _, w := range lockedWallets {
		currentWallets[strings.ToLower(w.UserAddress)] = w
		logger.LogWarn("Block: locked wallet",
			zap.String("address", w.UserAddress),
			zap.Float64("yte_balance", w.YTEBalance),
		)
	}

	currentBalances := make(map[string]float64)
	for _, addr := range addresses {
		lookupAddr := strings.ToLower(addr)
		if wallet, exists := currentWallets[lookupAddr]; exists {
			currentBalances[lookupAddr] = wallet.YTEBalance
		} else {
			currentBalances[lookupAddr] = 0
		}
	}
	logger.LogWarn("Block: initial balances", zap.Any("balances", currentBalances))

	// verify miner address
	if _, exists := currentBalances[minerAddr]; !exists {
		return models.Block{}, fmt.Errorf("%s not found in locked users", entity.MinerAccountAddress)
	}

	// Prepare basic info for block creation
	var ledgerEntries []repository.LedgerEntry
	var txIDs []int64
	totalFees := 0.00000000

	// Hitung balance per transaksi dengan accumulate (bukan overwrite).
	// Sebelumnya: txBalanceChanges[addr] = currentBalances[addr] - deduction
	// → setiap iterasi OVERWRITE nilai sebelumnya, bukan accumulate.
	// Jika address yang sama punya 2 transaksi, hanya yang terakhir yang
	// dipakai → miner fee terpotong, sender balance salah.
	// Fix: update currentBalances langsung (running balance).
	for _, t := range pendingTxs {
		totalDeduction := t.Amount + t.Fee
		from := strings.ToLower(t.FromAddress)
		to := strings.ToLower(t.ToAddress)
		currentBalances[from] -= totalDeduction
		currentBalances[to] += t.Amount
		currentBalances[minerAddr] += t.Fee

		totalFees += t.Fee
		txIDs = append(txIDs, t.ID)
	}

	// Create block FIRST to get blockID
	newBlock := models.Block{
		BlockNumber:  nextBlockNumber,
		PreviousHash: lastBlock.CurrentHash,
		CurrentHash:  miningResult.Hash,
		Nonce:        miningResult.Nonce,
		Difficulty:   miningResult.Difficulty,
		Timestamp:    blockTimestamp,
		MerkleRoot:   merkleRoot,
		MinerAddress: entity.MinerAccountAddress,
		BlockReward:  blockReward,
		TotalFees:    totalFees,
	}

	blockID, err := s.blockRepo.CreateWithTx(tx, newBlock)
	if err != nil {
		return models.Block{}, fmt.Errorf("create block: %w", err)
	}
	newBlock.ID = blockID

	// NOW create ledger entries. BalanceAfter menggunakan currentBalances
	// yang sudah ter-update per-tx di atas (running balance).
	for _, t := range pendingTxs {
		txID := t.ID
		txIDPtr := &txID
		totalDeduction := t.Amount + t.Fee

		ledgerEntries = append(ledgerEntries,
			repository.LedgerEntry{
				BlockID:      blockID,
				TxID:         txIDPtr,
				Address:      t.FromAddress,
				Amount:       -totalDeduction,
				BalanceAfter: currentBalances[strings.ToLower(t.FromAddress)],
			},
			repository.LedgerEntry{
				BlockID:      blockID,
				TxID:         txIDPtr,
				Address:      t.ToAddress,
				Amount:       t.Amount,
				BalanceAfter: currentBalances[strings.ToLower(t.ToAddress)],
			},
			repository.LedgerEntry{
				BlockID:      blockID,
				TxID:         txIDPtr,
				Address:      entity.MinerAccountAddress,
				Amount:       t.Fee,
				BalanceAfter: currentBalances[minerAddr],
			},
		)
	}

	// Bulk insert block-transaction links (1 query instead of N)
	err = s.blockRepo.BulkInsertBlockTransactionsWithTx(tx, blockID, txIDs)
	if err != nil {
		return models.Block{}, fmt.Errorf("bulk insert block transactions: %w", err)
	}

	// Bulk update transaction status (1 query instead of N)
	err = s.txRepo.BulkMarkConfirmedWithTx(tx, txIDs)
	if err != nil {
		return models.Block{}, fmt.Errorf("bulk mark confirmed: %w", err)
	}

	// Bulk update user balances using individual UPDATEs
	// Sebelumnya pakai CASE WHEN yang bisa gagal di edge case
	// (sqlx.In expansion dengan args kompleks).
	for addr, bal := range currentBalances {
		logger.LogWarn("Block: updating wallet balance",
			zap.String("address", addr),
			zap.Float64("new_balance", bal),
		)
		err = s.walletRepo.UpdateWalletWithTx(tx, addr, bal)
		if err != nil {
			return models.Block{}, fmt.Errorf("update wallet %s: %w", addr, err)
		}
	}

	// Process USD Balance Updates For BUY and SELL tx

	var buyerAddresses, sellerAddresses []string
	usdBalances := make(map[string]models.UserBalance)

	for _, t := range pendingTxs {
		if strings.EqualFold(t.Type, "BUY") {
			// Buyer to_address receives YTE Pays USD
			buyerAddresses = append(buyerAddresses, strings.ToLower(t.ToAddress))
		} else if strings.EqualFold(t.Type, "SELL") {
			// seller : from_address selss YTE receives USD
			sellerAddresses = append(sellerAddresses, strings.ToLower(t.FromAddress))
		}
	}

	// get all USD with lock
	allUSDAddresses := append(buyerAddresses, sellerAddresses...)
	allUSDAddresses = append(allUSDAddresses, minerAddr) // miner gets fee

	if len(allUSDAddresses) > 0 {
		lockedUSDBalances, err := s.balanceRepo.GetMultipleByAddressWithTxForUpdate(tx, allUSDAddresses)

		if err != nil {
			return models.Block{}, fmt.Errorf("lock multiple USD balances: %w", err)
		}

		for _, ub := range lockedUSDBalances {
			usdBalances[strings.ToLower(ub.UserAddress)] = ub
		}

		// ensure all address have USD balance record
		for _, addr := range allUSDAddresses {
			if _, exists := usdBalances[addr]; !exists {
				if err := s.balanceRepo.UpsertEmptyIfNotExistsWithTx(tx, addr); err != nil {
					return models.Block{}, fmt.Errorf("upsert empty USD balance: %w", err)
				}

				// refetch after upsert
				balance, err := s.balanceRepo.GetForUpdateWithTx(tx, addr)
				if err != nil {
					return models.Block{}, fmt.Errorf("refetch USD balance after upsert: %w", err)
				}
				usdBalances[addr] = balance
			}
		}
	}

	for _, t := range pendingTxs {
		if strings.EqualFold(t.Type, "BUY") {
			buyerAddr := strings.ToLower(t.ToAddress)

			// BUY: user membeli YTE dari sistem, membayar dalam USD.
			// Harga = amount * market price (konsisten dengan SELL).
			// Sebelumnya: totalCost = t.Amount + t.Fee (tanpa price conversion)
			// → BUY di harga 1:1 terlepas dari market price → arbitrage.
			usdCost := t.Amount * currentMarket.Price
			usdFee := t.Fee * currentMarket.Price
			totalUSDCost := usdCost + usdFee

			buyerBalance := usdBalances[buyerAddr]
			balanceAfter := buyerBalance.USDBalance - totalUSDCost

			usdBalances[buyerAddr] = models.UserBalance{
				UserAddress:     buyerAddr,
				USDBalance:      balanceAfter,
				LockedBalance:   buyerBalance.LockedBalance,
				TotalDeposited:  buyerBalance.TotalDeposited,
				TotalWithdrawn:  buyerBalance.TotalWithdrawn + totalUSDCost,
				TotalTraded:     buyerBalance.TotalTraded + usdCost,
				LastTransaction: buyerBalance.LastTransaction,
			}

			// miner fee receives USD (dalam USD, bukan YTE)
			minerBalance := usdBalances[minerAddr]
			minerBalance.USDBalance += usdFee
			usdBalances[minerAddr] = minerBalance

		} else if strings.EqualFold(t.Type, "SELL") {
			sellerAddr := strings.ToLower(t.FromAddress)

			// use market price to calculate USD received
			usdAmount := t.Amount * currentMarket.Price
			usdFee := t.Fee * currentMarket.Price

			sellerBalance := usdBalances[sellerAddr]
			balanceAfter := sellerBalance.USDBalance + usdAmount

			usdBalances[sellerAddr] = models.UserBalance{
				UserAddress:     sellerAddr,
				USDBalance:      balanceAfter,
				LockedBalance:   sellerBalance.LockedBalance,
				TotalDeposited:  sellerBalance.TotalDeposited + usdAmount,
				TotalWithdrawn:  sellerBalance.TotalWithdrawn,
				TotalTraded:     sellerBalance.TotalTraded + usdAmount,
				LastTransaction: sellerBalance.LastTransaction,
			}

			// miner fee receives USD
			minerBalance := usdBalances[minerAddr]
			minerBalance.USDBalance += usdFee
			usdBalances[minerAddr] = minerBalance
		}
	}

	// bulk update all usd balances
	if err := s.balanceRepo.BulkUpdateBalancesWithTx(tx, usdBalances); err != nil {
		return models.Block{}, fmt.Errorf("bulk update USD balances: %w", err)
	}

	// End process USD

	var marketState models.MarketEngine
	if s.market != nil {
		if marketState, err = s.market.ApplyBlockPricingWithTx(tx, blockID, buyVolume, sellVolume, len(pendingTxs)); err != nil {
			return models.Block{}, fmt.Errorf("apply market pricing: %w", err)
		}
	}

	// store market tick
	var marketTick models.MarketTick
	if s.market != nil {
		marketTick = models.MarketTick{
			BlockID:    blockID,
			Price:      marketState.Price,
			BuyVolume:  buyVolume,
			SellVolume: sellVolume,
			TxCount:    len(pendingTxs),
			CreatedAt:  time.Now().Unix(),
		}
	}

	// Commit (total transaction time: < 2 seconds)
	if err := tx.Commit(); err != nil {
		return models.Block{}, fmt.Errorf("commit transaction: %w", err)
	}

	ctx := context.Background()

	// PHASE 3: Async Event Publishing (POST-COMMIT)

	// publish ledger batch event
	if s.ledgerPublisher != nil {
		ledgerEvents := make([]dto.LedgerEntryEvent, 0, len(ledgerEntries))

		for _, entry := range ledgerEntries {
			event := dto.LedgerEntryEvent{
				Address:      entry.Address,
				Amount:       entry.Amount,
				BalanceAfter: entry.BalanceAfter,
			}
			if entry.TxID != nil {
				event.TxID = entry.TxID
			}

			ledgerEvents = append(ledgerEvents, event)
		}

		if err := s.ledgerPublisher.PublishLedgerBatch(
			ctx,
			blockID,
			newBlock.BlockNumber,
			ledgerEvents,
			newBlock.MinerAddress,
		); err != nil {
			logger.LogWarn("Failed to publish ledger batch", zap.Error(err))
		}
	}

	// publish market pricing event
	if s.pricingPublisher != nil && marketState.ID != 0 {
		if err := s.pricingPublisher.PublishPricingEvent(
			ctx,
			blockID,
			newBlock.BlockNumber,
			marketState,
			marketTick,
			newBlock.MinerAddress,
		); err != nil {
			logger.LogWarn("Failed to publish market pricing event", zap.Error(err))
		}
	}

	//broadcast new block mined
	if s.publisherWS != nil {
		s.publisherWS.Publish(entity.EventTypeBlockMined, newBlock)
	}

	// Publish notification: BLOCK_CONFIRMED untuk semua subscriber
	if s.notificationPublisher != nil {
		blockNotif := dto.NewNotificationEvent(
			dto.TypeBlockConfirmed,
			dto.PriorityMedium,
			entity.MinerAccountAddress,
			fmt.Sprintf("Block #%d Berhasil Ditambang", newBlock.BlockNumber),
			fmt.Sprintf("Block %d ditambang dengan %d transaksi, total fee %.8f YTE",
				newBlock.BlockNumber, len(pendingTxs), newBlock.TotalFees),
			[]dto.NotificationChannel{dto.ChannelWebSocket},
		)
		blockNotif.RelatedBlockID = &blockID
		blockNotif.Data = map[string]interface{}{
			"block_number": newBlock.BlockNumber,
			"tx_count":     len(pendingTxs),
			"total_fees":   newBlock.TotalFees,
			"block_hash":   newBlock.CurrentHash,
		}
		_ = s.notificationPublisher.Publish(ctx, *blockNotif)
	}

	// publish reward calculation event
	if s.rewardPublisher != nil {
		rewardCalcEvent := dto.RewardCalculationEvent{
			BlockID:             blockID,
			BlockNumber:         newBlock.BlockNumber,
			MinerAddress:        newBlock.MinerAddress,
			BlockReward:         newBlock.BlockReward,
			TransactionCount:    len(pendingTxs),
			TotalTransactionFee: newBlock.TotalFees,
			MarketPrice:         marketState.Price,
			Timestamp:           time.Now().Unix(),
		}

		if err := s.rewardPublisher.PublishRewardCalculation(ctx, rewardCalcEvent); err != nil {
			logger.LogWarn("Failed to publish reward calculation event", zap.Error(err))
		}
	}

	// load transactions
	transactions, err := s.txRepo.GetTransactionsByBlockID(blockID)
	if err != nil {
		logger.LogWarn("Failed to load transactions", zap.Error(err))
	} else {
		newBlock.Transactions = transactions
	}

	// send notifycation to websocket
	if s.publisherWS != nil && len(newBlock.Transactions) > 0 {
		for _, tx := range newBlock.Transactions {
			payload := tx
			if tx.FromAddress != entity.MinerAccountAddress {
				s.publisherWS.PublishToAddress(strings.ToLower(tx.FromAddress), entity.EventTransactionUpdate, payload)
			}

			if tx.ToAddress != entity.MinerAccountAddress {
				s.publisherWS.PublishToAddress(strings.ToLower(tx.ToAddress), entity.EventTransactionUpdate, payload)
			}

			// Publish TRANSACTION_CONFIRMED notification ke pengirim
			if s.notificationPublisher != nil && tx.FromAddress != entity.MinerAccountAddress {
				txNotif := dto.NewNotificationEvent(
					dto.TypeTransactionConfirmed,
					dto.PriorityHigh,
					strings.ToLower(tx.FromAddress),
					"Transaksi Terkonfirmasi",
					fmt.Sprintf("Transaksi %.8f YTE ke %s telah dikonfirmasi di block #%d",
						tx.Amount, shortAddr(tx.ToAddress), newBlock.BlockNumber),
					[]dto.NotificationChannel{dto.ChannelWebSocket},
				)
				txNotif.RelatedTxID = &tx.ID
				txNotif.RelatedBlockID = &blockID
				txNotif.Data = map[string]interface{}{
					"tx_id":         tx.ID,
					"from_address":  tx.FromAddress,
					"to_address":    tx.ToAddress,
					"amount":        tx.Amount,
					"fee":           tx.Fee,
					"type":          tx.Type,
					"block_number":  newBlock.BlockNumber,
					"block_hash":    newBlock.CurrentHash,
				}
				_ = s.notificationPublisher.Publish(ctx, *txNotif)
			}
		}
	}

	minerWallet, _ := s.walletRepo.GetByAddress(entity.MinerAccountAddress)
	logger.LogBlockEvent(
		int64(newBlock.BlockNumber),
		"mined",
		zap.String("hash", newBlock.CurrentHash),
		zap.String("merkle_root", newBlock.MerkleRoot),
		zap.Int64("nonce", newBlock.Nonce),
		zap.Int64("difficulty", int64(newBlock.Difficulty)),
		zap.Int("transaction_count", len(pendingTxs)),
		zap.Float64("total_fees", totalFees),
		zap.Float64("block_reward", blockReward),
		zap.Float64("total_earned", blockReward+totalFees),
		zap.Float64("miner_balance_before", minerWallet.YTEBalance),
		zap.Float64("current_supply", utils.GetCurrentSupply(int64(nextBlockNumber))),
		zap.Float64("max_supply", utils.GetMaxSupply()),
		zap.Int64("next_halving_block", utils.GetNextHalvingBlock(int64(nextBlockNumber))),
		zap.Int64("blocks_until_halving", utils.GetBlocksUntilHalving(int64(nextBlockNumber))),
		zap.Duration("mining_time", miningResult.Duration),
	)

	return newBlock, nil
}

func (s *blockService) GetBlocks(limit, offset int) ([]models.Block, error) {
	return s.blockRepo.GetBlocks(limit, offset)
}

func (s *blockService) GetBlockByID(id int64) (models.Block, error) {
	return s.blockRepo.GetBlockByID(id)
}

func (s *blockService) GetBlockByBlockNumber(id int64) (models.Block, error) {
	return s.blockRepo.GetBlockByBlockNumber(id)
}

func (s *blockService) CheckBlockchainIntegrity() error {
	blocks, err := s.blockRepo.GetAllBlocks() // ambil semua block

	if err != nil {
		return err
	}

	err = utils.CheckBlockchainIntegrity(blocks)

	return err
}

func (s *blockService) GetDetailsByBlockNumber(id int64) (models.Block, error) {
	var block models.Block

	block, err := s.blockRepo.GetBlockByBlockNumber(id)
	if err != nil {
		return models.Block{}, err
	}

	// Populate transactions for the block
	tx, err := s.txRepo.GetTransactionsByBlockID(block.ID)

	if err != nil {
		return models.Block{}, err
	}

	block.Transactions = tx

	return block, nil
}

func (s *blockService) GetTransactionByBlockNumber(ctx context.Context, blockNumber int64) ([]models.Transaction, error) {
	return s.blockRepo.GetTransactionsByBlockNumber(ctx, blockNumber)
}

func (s *blockService) SearchBlocksByHash(ctx context.Context, hash string) ([]models.Block, error) {
	return s.blockRepo.SearchByHash(ctx, hash)
}

func (s *blockService) GetBlocksInRange(ctx context.Context, from, to int64) ([]models.Block, error) {
	return s.blockRepo.GetBlocksInRange(ctx, from, to)
}

func (s *blockService) GetBlockStats(ctx context.Context) (dto.BlockStatsResponse, error) {
	var (
		blockStats  models.BlockStats
		latestBlock models.Block
		countHour   int64
	)

	// error group
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		var err error
		blockStats, err = s.blockRepo.GetBlockStats(ctx)
		return err
	})

	g.Go(func() error {
		var err error
		latestBlock, err = s.blockRepo.GetLatestBlockInfo(ctx)
		return err
	})

	g.Go(func() error {
		var err error
		countHour, err = s.blockRepo.GetBlockCountLastHour(ctx)
		return err
	})

	// wait all goroutine finish
	if err := g.Wait(); err != nil {
		return dto.BlockStatsResponse{}, err
	}

	response := dto.BlockStatsResponse{
		TotalBlocks:       int(blockStats.TotalBlocks),
		AverageDifficulty: blockStats.AverageDifficulty,
		TotalTransactions: int(blockStats.TotalTransactions),
		TotalFees:         blockStats.TotalFees,
		AvgTxPerBlock:     blockStats.AvgTxPerBlock,
		TotalBlockRewards: float64(blockStats.AverageBlockReward * float64(blockStats.TotalBlocks)),
		LatestBlock: dto.LatestBlockInfo{
			BlockNumber:  int64(latestBlock.BlockNumber),
			Hash:         latestBlock.CurrentHash,
			Timestamp:    latestBlock.Timestamp,
			Transactions: latestBlock.Transactions,
			MinerAddress: latestBlock.MinerAddress,
			BlockReward:  latestBlock.BlockReward,
			TotalFees:    latestBlock.TotalFees,
		},
		LastHourBlockCount: countHour,
	}

	return response, nil
}

func (s *blockService) SearchBlocksByMinerAddress(ctx context.Context, address string, limit, offset int) ([]models.Block, error) {
	// validate limit and offset
	if limit <= 0 {
		limit = 10
	} else if limit > 100 {
		limit = 100
	}

	if offset < 0 {
		offset = 0
	}

	return s.blockRepo.SearchByMinerAddress(ctx, address, limit, offset)
}
